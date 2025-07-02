// internal/modules/realtime/ws_handler.go
package realtime

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSHandler(c *gin.Context) {
	userID := c.Query("user_id") // hoặc lấy từ JWT
	if userID == "" {
		c.JSON(400, gin.H{"error": "user_id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	log.Printf("User %s connected to WebSocket", userID)
	Add(userID, conn)
	defer func() {
		Remove(userID)
		conn.Close()
		log.Printf("User %s disconnected from WebSocket", userID)
	}()

	// Set up ping/pong to detect broken connections
	conn.SetReadLimit(512)
	conn.SetPongHandler(func(string) error {
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", userID, err)
			}
			break
		}
	}
}
