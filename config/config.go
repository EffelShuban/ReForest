package config

import (
	"os"
)

type Config struct {
	DBDSN           string
	MongoURI        string
	AuthGRPCPort    string
	TreeGRPCPort    string
	FinanceGRPCPort string
	JWTSecret       string
	MailtrapHost    string
	MailtrapPort    string
	MailtrapUser    string
	MailtrapPass    string
	MailtrapFrom    string
	RabbitMQURL     string
	XenditAPIKey    string
	AllowedOrigin   string
}

func Load() *Config {
	return &Config{
		DBDSN:           getEnv("DB_DSN", "postgres://user:password@localhost:5432/reforest_db?sslmode=disable"),
		MongoURI:        getEnv("MONGO_URI", "mongodb://root:password@localhost:27017"),
		AuthGRPCPort:    getEnv("AUTH_GRPC_PORT", ":50051"),
		TreeGRPCPort:    getEnv("TREE_GRPC_PORT", ":50052"),
		FinanceGRPCPort: getEnv("FINANCE_GRPC_PORT", ":50053"),
		JWTSecret:       getEnv("JWT_SECRET", "secret"),
		MailtrapHost:    getEnv("MAILTRAP_HOST", "smtp.mailtrap.io"),
		MailtrapPort:    getEnv("MAILTRAP_PORT", "2525"),
		MailtrapUser:    getEnv("MAILTRAP_USER", ""),
		MailtrapPass:    getEnv("MAILTRAP_PASS", ""),
		MailtrapFrom:    getEnv("MAILTRAP_FROM", "no-reply@reforest.com"),
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		XenditAPIKey:    getEnv("XENDIT_API_KEY", ""),
		AllowedOrigin:   getEnv("ALLOWED_ORIGIN", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}