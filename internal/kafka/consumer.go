// =============================================================================
// internal/kafka/consumer.go
// 🎧 KAFKA CONSUMER: Lắng nghe và xử lý events từ các topics
// =============================================================================
package kafka

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// ConsumerHandler interface để xử lý messages
// 🔌 INTERFACE: Cho phép inject different handlers
type ConsumerHandler interface {
	HandleMessage(message *sarama.ConsumerMessage) error
}

type Consumer struct {
	consumer sarama.ConsumerGroup // Kafka consumer group
	topics   []string             // Danh sách topics cần lắng nghe
	handler  ConsumerHandler      // Handler để xử lý messages
	ctx      context.Context      // Context để control lifecycle
	cancel   context.CancelFunc   // Function để cancel context
	wg       sync.WaitGroup       // WaitGroup để đợi goroutines
}

// NewConsumer creates a new Kafka consumer
// 🏗️ CONSTRUCTOR: Khởi tạo consumer với config
func NewConsumer(brokers []string, groupID string, topics []string, handler ConsumerHandler) (*Consumer, error) {
	// ⚙️ KAFKA CONFIG: Cấu hình consumer
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin // Load balancing
	config.Consumer.Offsets.Initial = sarama.OffsetNewest                       // Chỉ đọc messages mới
	config.Consumer.Group.Session.Timeout = 10 * time.Second                    // Timeout 10s
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second                  // Heartbeat 3s

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		consumer: consumer,
		topics:   topics,
		handler:  handler,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start starts consuming messages
// 🚀 BẮT ĐẦU LẮNG NGHE: Goroutine chạy liên tục để consume messages
func (c *Consumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done(): // 🛑 Nhận signal stop
				return
			default:
				// 🔄 CONSUME LOOP: Liên tục lắng nghe messages
				if err := c.consumer.Consume(c.ctx, c.topics, &consumerGroupHandler{handler: c.handler}); err != nil {
					log.Printf("❌ Consumer error: %v", err)
				}
			}
		}
	}()

	// 🛡️ GRACEFUL SHUTDOWN: Lắng nghe OS signals
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	log.Println("🛑 Terminating consumer...")
	c.Stop()
}

// Stop stops the consumer
// 🛑 DỪNG CONSUMER: Clean shutdown
func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
	if err := c.consumer.Close(); err != nil {
		log.Printf("❌ Failed to close consumer: %v", err)
	}
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
// 🎯 SARAMA HANDLER: Required interface để implement
type consumerGroupHandler struct {
	handler ConsumerHandler
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim xử lý messages từ một partition
// 📨 XỬ LÝ MESSAGES: Loop qua tất cả messages trong claim
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// 🎯 GỌI HANDLER: Delegate việc xử lý cho custom handler
		if err := h.handler.HandleMessage(message); err != nil {
			log.Printf("❌ Failed to handle message: %v", err)
			continue // ⚡ CONTINUE ON ERROR: Không stop toàn bộ consumer
		}
		// ✅ MARK MESSAGE: Báo Kafka là đã xử lý xong
		session.MarkMessage(message, "")
	}
	return nil
}
