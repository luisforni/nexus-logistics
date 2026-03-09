package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {

	Port   string
	AppEnv string

	DatabaseURL string

	RedisURL      string
	RedisPassword string
	RedisDB       int

	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	EthereumRPCURL  string
	ContractAddress string
	ChainID         int64

	OptimizerHost string
	OptimizerPort string

	OtelEndpoint string
	LogLevel     string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig()
	v.AutomaticEnv()

	v.SetDefault("BACKEND_PORT", "8080")
	v.SetDefault("APP_ENV", "production")
	v.SetDefault("JWT_EXPIRY_MINUTES", 15)
	v.SetDefault("JWT_REFRESH_EXPIRY_HOURS", 168)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("OPTIMIZER_HOST", "optimizer")
	v.SetDefault("OPTIMIZER_PORT", "9090")
	v.SetDefault("ETHEREUM_CHAIN_ID", 11155111)
	v.SetDefault("LOG_LEVEL", "info")

	jwtSecret := v.GetString("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	dbURL := v.GetString("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return &Config{
		Port:             v.GetString("BACKEND_PORT"),
		AppEnv:           v.GetString("APP_ENV"),
		DatabaseURL:      dbURL,
		RedisURL:         v.GetString("REDIS_URL"),
		RedisPassword:    v.GetString("REDIS_PASSWORD"),
		RedisDB:          v.GetInt("REDIS_DB"),
		JWTSecret:        jwtSecret,
		JWTExpiry:        time.Duration(v.GetInt("JWT_EXPIRY_MINUTES")) * time.Minute,
		JWTRefreshExpiry: time.Duration(v.GetInt("JWT_REFRESH_EXPIRY_HOURS")) * time.Hour,
		EthereumRPCURL:   v.GetString("ETHEREUM_RPC_URL"),
		ContractAddress:  v.GetString("SHIPMENT_CONTRACT_ADDRESS"),
		ChainID:          v.GetInt64("ETHEREUM_CHAIN_ID"),
		OptimizerHost:    v.GetString("OPTIMIZER_HOST"),
		OptimizerPort:    v.GetString("OPTIMIZER_PORT"),
		OtelEndpoint:     v.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
		LogLevel:         v.GetString("LOG_LEVEL"),
	}, nil
}
