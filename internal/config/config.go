package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:",squash"`
	Database DatabaseConfig `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
	Kafka    KafkaConfig    `mapstructure:",squash"`
	JWT      JWTConfig      `mapstructure:",squash"`
	Cache    CacheConfig    `mapstructure:",squash"`
}

type ServerConfig struct {
	Port            int           `mapstructure:"SERVER_PORT"`
	Environment     string        `mapstructure:"ENVIRONMENT"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DBName   string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type KafkaConfig struct {
	Brokers  []string `mapstructure:"KAFKA_BROKERS"`
	Topic    string   `mapstructure:"KAFKA_TOPIC"`
	GroupID  string   `mapstructure:"KAFKA_GROUP_ID"`
}

type JWTConfig struct {
	SecretKey       string        `mapstructure:"JWT_SECRET_KEY"`
	AccessDuration  time.Duration `mapstructure:"JWT_ACCESS_DURATION"`
	RefreshDuration time.Duration `mapstructure:"JWT_REFRESH_DURATION"`
}

func Load() (*Config, error) {
	// Set default values
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SHUTDOWN_TIMEOUT", "10s")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "rayda")
	viper.SetDefault("DB_SSLMODE", "disable")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("KAFKA_TOPIC", "events")
	viper.SetDefault("KAFKA_GROUP_ID", "rayda-service")

	viper.SetDefault("JWT_ACCESS_DURATION", "15m")
	viper.SetDefault("JWT_REFRESH_DURATION", "24h")

	// Read from .env file if it exists
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Read from environment variables
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// Set JWT secret key if not set in config
	if config.JWT.SecretKey == "" {
		if secret := os.Getenv("JWT_SECRET_KEY"); secret != "" {
			config.JWT.SecretKey = secret
		} else {
			// In production, this should be set via environment variable
			config.JWT.SecretKey = "default-secret-key-change-in-production"
		}
	}

	return &config, nil
}
