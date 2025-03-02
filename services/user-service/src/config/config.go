package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config holds all configuration for the service
type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	JWT     JWTConfig
	Kafka   KafkaConfig
	AuthSvc AuthServiceConfig
	Logging LoggingConfig
	CORS    CORSConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// MongoDBConfig holds MongoDB-related configuration
type MongoDBConfig struct {
	URI         string
	DBName      string
	Timeout     time.Duration
	MaxPoolSize uint64
	MinPoolSize uint64
}

// JWTConfig holds JWT validation configuration
type JWTConfig struct {
	Secret string
	Issuer string
}

// KafkaConfig holds Kafka-related configuration
type KafkaConfig struct {
	Brokers         []string
	GroupID         string
	ClientID        string
	AutoOffsetReset string
	Topics          KafkaTopics
}

// KafkaTopics holds Kafka topic names
type KafkaTopics struct {
	UserEvents string
	AuthEvents string
	TeamEvents string
}

// AuthServiceConfig holds Auth Service connection details
type AuthServiceConfig struct {
	URL string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found")
	}

	// Use viper to load environment variables
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Return the config
	return &Config{
		Server: ServerConfig{
			Port:    viper.GetString("PORT"),
			GinMode: viper.GetString("GIN_MODE"),
		},
		MongoDB: MongoDBConfig{
			URI:         viper.GetString("MONGO_URI"),
			DBName:      viper.GetString("MONGO_DB_NAME"),
			Timeout:     time.Duration(viper.GetInt("MONGO_TIMEOUT")) * time.Second,
			MaxPoolSize: viper.GetUint64("MONGO_MAX_POOL_SIZE"),
			MinPoolSize: viper.GetUint64("MONGO_MIN_POOL_SIZE"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
			Issuer: viper.GetString("JWT_ISSUER"),
		},
		Kafka: KafkaConfig{
			Brokers:         viper.GetStringSlice("KAFKA_BROKERS"),
			GroupID:         viper.GetString("KAFKA_GROUP_ID"),
			ClientID:        viper.GetString("KAFKA_CLIENT_ID"),
			AutoOffsetReset: viper.GetString("KAFKA_AUTO_OFFSET_RESET"),
			Topics: KafkaTopics{
				UserEvents: viper.GetString("KAFKA_TOPIC_USER_EVENTS"),
				AuthEvents: viper.GetString("KAFKA_TOPIC_AUTH_EVENTS"),
				TeamEvents: viper.GetString("KAFKA_TOPIC_TEAM_EVENTS"),
			},
		},
		AuthSvc: AuthServiceConfig{
			URL: viper.GetString("AUTH_SERVICE_URL"),
		},
		Logging: LoggingConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
		CORS: CORSConfig{
			AllowedOrigins: viper.GetString("CORS_ALLOWED_ORIGINS"),
		},
	}, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("PORT", "8001")
	viper.SetDefault("GIN_MODE", "debug")

	// MongoDB defaults
	viper.SetDefault("MONGO_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGO_DB_NAME", "slidoclone_users")
	viper.SetDefault("MONGO_TIMEOUT", 10)
	viper.SetDefault("MONGO_MAX_POOL_SIZE", 100)
	viper.SetDefault("MONGO_MIN_POOL_SIZE", 5)

	// JWT defaults
	viper.SetDefault("JWT_SECRET", "your_jwt_secret_here")
	viper.SetDefault("JWT_ISSUER", "slido-clone-auth")

	// Kafka defaults
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("KAFKA_GROUP_ID", "user-service-group")
	viper.SetDefault("KAFKA_CLIENT_ID", "user-service")
	viper.SetDefault("KAFKA_AUTO_OFFSET_RESET", "earliest")

	// Kafka topic defaults
	viper.SetDefault("KAFKA_TOPIC_USER_EVENTS", "user.events")
	viper.SetDefault("KAFKA_TOPIC_AUTH_EVENTS", "auth.events")
	viper.SetDefault("KAFKA_TOPIC_TEAM_EVENTS", "team.events")

	// Auth Service defaults
	viper.SetDefault("AUTH_SERVICE_URL", "http://localhost:3001")

	// Logging defaults
	viper.SetDefault("LOG_LEVEL", "debug")

	// CORS defaults
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")
}

// String returns a string representation of the config
func (c *Config) String() string {
	return fmt.Sprintf(`
Server:
  Port: %s
  GinMode: %s
MongoDB:
  URI: %s
  DBName: %s
  Timeout: %v
  MaxPoolSize: %d
  MinPoolSize: %d
JWT:
  Secret: %s
  Issuer: %s
Kafka:
  Brokers: %v
  GroupID: %s
  ClientID: %s
  AutoOffsetReset: %s
  Topics:
    UserEvents: %s
    AuthEvents: %s
    TeamEvents: %s
AuthService:
  URL: %s
Logging:
  Level: %s
CORS:
  AllowedOrigins: %s
`,
		c.Server.Port,
		c.Server.GinMode,
		c.MongoDB.URI,
		c.MongoDB.DBName,
		c.MongoDB.Timeout,
		c.MongoDB.MaxPoolSize,
		c.MongoDB.MinPoolSize,
		maskString(c.JWT.Secret),
		c.JWT.Issuer,
		c.Kafka.Brokers,
		c.Kafka.GroupID,
		c.Kafka.ClientID,
		c.Kafka.AutoOffsetReset,
		c.Kafka.Topics.UserEvents,
		c.Kafka.Topics.AuthEvents,
		c.Kafka.Topics.TeamEvents,
		c.AuthSvc.URL,
		c.Logging.Level,
		c.CORS.AllowedOrigins,
	)
}

// maskString masks a string for logging purposes
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
