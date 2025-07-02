// internal/modules/realtime/ws_manager.go
package realtime

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WSManager struct {
	conns map[string]*websocket.Conn
	mu    sync.RWMutex
}

var manager = &WSManager{conns: make(map[string]*websocket.Conn)}

func Add(userID string, conn *websocket.Conn) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Đóng connection cũ nếu có
	if oldConn, exists := manager.conns[userID]; exists {
		oldConn.Close()
		log.Printf("Closed old connection for user %s", userID)
	}

	manager.conns[userID] = conn
	log.Printf("Added connection for user %s. Total connections: %d", userID, len(manager.conns))
}

func Remove(userID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if conn, exists := manager.conns[userID]; exists {
		conn.Close()
		delete(manager.conns, userID)
		log.Printf("Removed connection for user %s. Total connections: %d", userID, len(manager.conns))
	}
}

func Send(userID, msg string) error {
	manager.mu.RLock()
	conn, ok := manager.conns[userID]
	manager.mu.RUnlock()

	if !ok {
		log.Printf("No connection found for user %s", userID)
		return fmt.Errorf("user %s not connected", userID)
	}

	// Lock để tránh concurrent write
	manager.mu.Lock()
	err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
	manager.mu.Unlock()

	if err != nil {
		log.Printf("Failed to send message to user %s: %v", userID, err)
		// Remove broken connection
		Remove(userID)
		return err
	}

	log.Printf("Message sent successfully to user %s: %s", userID, msg)
	return nil
}

// Thêm function để check user online
func IsUserOnline(userID string) bool {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	_, exists := manager.conns[userID]
	return exists
}

// Thêm function để get tất cả connected users (for debugging)
func GetConnectedUsers() []string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	users := make([]string, 0, len(manager.conns))
	for userID := range manager.conns {
		users = append(users, userID)
	}
	return users
}
