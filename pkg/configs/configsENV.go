package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type ConectionConfigs struct {
	PostgresCF *DataBasePostgresConfig
	KafkaCF    *KafkaConfig
	RedisCF    *RedisConfig
	ServerCF   *ServerConfig
	SMTPCF     *STMPConfig
}
type STMPConfig struct {
	SmtpHost     string
	SmtpPort     string
	SmtpUsername string
	SmtpPassword string
	FromName     string
	FromEmail    string
	BaseURL      string
}
type ServerConfig struct {
	Port      string
	Host      string
	GinMode   string
	JWTSecret string
	JWTExpiry time.Duration
}

type DataBasePostgresConfig struct {
	Host     string
	Port     string
	UserName string
	Password string
	DBName   string
	SSLMode  string
}

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topics  KafkaTopics
}

type KafkaTopics struct {
	BookingCreated   string
	BookingUpdated   string
	BookingCancelled string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func LoadConfigs() (*ConectionConfigs, error) { // Try to load .env file from multiple locations
	envPaths := []string{
		"/app/.env",                              // Docker container path
		".env",                                   // Local development
		"../.env",                                // One level up
		filepath.Join(os.Getenv("HOME"), ".env"), // Home directory
	}

	envLoaded := false
	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			fmt.Printf("🔍 Trying to load .env from: %s\n", envPath)
			if err := godotenv.Load(envPath); err == nil {
				fmt.Printf("✅ Successfully loaded .env from: %s\n", envPath)
				envLoaded = true
				break
			} else {
				fmt.Printf("⚠️ Failed to load .env from %s: %v\n", envPath, err)
			}
		}
	}

	if !envLoaded {
		fmt.Println("ℹ️ No .env file loaded, using environment variables and defaults")
	}

	// Log environment variables for debugging
	fmt.Printf("📊 Environment variables:\n")
	fmt.Printf("  KAFKA_BROKERS: '%s'\n", os.Getenv("KAFKA_BROKERS"))
	fmt.Printf("  DB_HOST: '%s'\n", os.Getenv("DB_HOST_POSTGRES"))
	fmt.Printf("  API_PORT: '%s'\n", os.Getenv("API_PORT"))

	cfg := &ConectionConfigs{
		ServerCF: &ServerConfig{
			Port:      getEnv("API_PORT", "8080"),
			Host:      getEnv("API_HOST", "0.0.0.0"),
			GinMode:   getEnv("GIN_MODE", "debug"),
			JWTSecret: getEnv("JWT_SECRET", "abc123"),
			JWTExpiry: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		SMTPCF: &STMPConfig{
			SmtpHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SmtpPort:     getEnv("SMTP_PORT", "587"),
			SmtpUsername: getEnv("SMTP_USERNAME", "truongvankhangthanthanh@gmail.com"),
			SmtpPassword: getEnv("SMTP_PASSWORD", "hothingoctram12032004"),
			FromName:     getEnv("FROM_NAME", "Consultation Booking System"),
			FromEmail:    getEnv("FROM_Email", "truongvankhangthanthanh@gmail.com"),
			BaseURL:      getEnv("BASE_URL", "http://localhost:8899"),
		},
		PostgresCF: &DataBasePostgresConfig{
			Host:     getEnv("DB_HOST_POSTGRES", "localhost"),
			Port:     getEnv("DB_PORT_POSTGRES", "5432"),
			UserName: getEnv("DB_USER_POSTGRES", "postgres"),
			Password: getEnv("DB_PASSWORD_POSTGRES", "khangmc1502@"),
			DBName:   getEnv("DB_NAME_POSTGRES", "Create_APPP"),
			SSLMode:  getEnv("DB_SSL_MODE_POSTGRES", "disable"),
		},
		// DBMC: &DatabaseMsqlConfig{
		// 	Host:     getEnv("DB_HOST", "localhost"),
		// 	Port:     getEnv("DB_PORT", "1443"),
		// 	UserName: getEnv("DB_USER", "sa"),
		// 	Password: getEnv("DB_PASSWORD", "khangmc1502@"),
		// 	DBName:   getEnv("DB_NAME", "ecommerce"),
		// },
		KafkaCF: &KafkaConfig{
			Brokers: getEnvSlice("KAFKA_BROKERS", []string{"kafka:9092"}),
			GroupID: getEnv("KAFKA_GROUP_ID", "CB_System"),
			Topics: KafkaTopics{
				BookingCreated:   getEnv("KAFKA_TOPIC_BOOKING_EVENTS", "booking-events"),
				BookingUpdated:   getEnv("KAFKA_TOPIC_USER_NOTIFICATIONS", "user-notifications"),
				BookingCancelled: getEnv("KAFKA_TOPIC_USER_NOTIFICATIONS", "user-notifications"),
			},
		},
		RedisCF: &RedisConfig{
			Host:     getEnv("RD_HOST", ""),
			Port:     getEnv("RD_PORT", ""),
			Password: getEnv("RD_PASSWORD", ""),
		},
	}

	if len(cfg.KafkaCF.Brokers) == 0 || (len(cfg.KafkaCF.Brokers) == 1 && cfg.KafkaCF.Brokers[0] == "") {
		return nil, fmt.Errorf("❌ No Kafka brokers configured. Please set KAFKA_BROKERS environment variable")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	var cleaned []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	fmt.Printf("💬 Raw KAFKA_BROKERS from env: '%s'\n", os.Getenv("KAFKA_BROKERS"))

	// 💡 Nếu cleaned rỗng, fallback về default
	if len(cleaned) == 0 {
		return defaultValue
	}

	return cleaned
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("⚠️ Invalid duration for %s: %s, using default: %v", key, value, defaultValue)
		return defaultValue
	}
	return duration
}

// func getEnvInt(key string, defaultValue int) int {
// 	value := strings.TrimSpace(os.Getenv(key))
// 	if value == "" {
// 		return defaultValue
// 	}
// 	var i int
// 	_, err := fmt.Sscanf(value, "%d", &i)
// 	if err != nil {
// 		log.Printf("⚠️ Invalid integer for %s: %s, using default: %d", key, value, defaultValue)
// 		return defaultValue
// 	}
// 	return i
// }
