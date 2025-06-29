// ============================================================================
// internal/kafka/producer.go - HOÀN THIỆN PRODUCER
// ============================================================================
package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

var producer sarama.SyncProducer

// InitProducer khởi tạo Kafka producer
func InitProducer(brokers []string) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	var err error
	producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}

	log.Println("✅ Kafka producer initialized successfully")
	return nil
}

// Publish gửi message lên Kafka topic
func Publish(topic string, data []byte) error {
	if producer == nil {
		return fmt.Errorf("producer not initialized")
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(data),
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		log.Printf("❌ Failed to send message to topic %s: %v", topic, err)
		return err
	}

	log.Printf("📤 Message sent to topic: %s, partition: %d, offset: %d", topic, partition, offset)
	return nil
}

// CloseProducer đóng producer
func CloseProducer() error {
	if producer != nil {
		return producer.Close()
	}
	return nil
}
