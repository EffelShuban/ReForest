package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN           string
	MongoDSN        string
	JWTSecret       string
	AuthGRPCPort    string
	TreeGRPCPort    string
	FinanceGRPCPort string
	XenditAPIKey    string
	RabbitMQURL     string
	MailtrapHost    string
	MailtrapPort    string
	MailtrapUser    string
	MailtrapPass    string
	MailtrapFrom    string
	RabbitMQURL     string
}

func Load() *Config {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=forest_guard port=5432 sslmode=disable"
	}

	mongoDSN := os.Getenv("MONGO_DSN")
	if mongoDSN == "" {
		mongoDSN = "mongodb://root:password@localhost:27017/reforest_db?authSource=admin"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key"
	}

	authGrpcPort := os.Getenv("AUTH_GRPC_PORT")
	if authGrpcPort == "" {
		authGrpcPort = ":50051"
	}

	treeGrpcPort := os.Getenv("TREE_GRPC_PORT")
	if treeGrpcPort == "" {
		treeGrpcPort = ":50052"
	}

	financeGrpcPort := os.Getenv("FINANCE_GRPC_PORT")
	if financeGrpcPort == "" {
		financeGrpcPort = ":50053"
	}

	xenditAPIKey := os.Getenv("XENDIT_API_KEY")
	if xenditAPIKey == "" {
		xenditAPIKey = "xnd_public_development_G3NFqweuzg_Ke9pjaBtZBSIExYCVP0yQA77ZeETSbcI_4yMx5MFc60MROYi0Je"
	}

	mailtrapHost := os.Getenv("MAILTRAP_HOST")
	if mailtrapHost == "" {
		mailtrapHost = "sandbox.smtp.mailtrap.io"
	}

	mailtrapPort := os.Getenv("MAILTRAP_PORT")
	if mailtrapPort == "" {
		mailtrapPort = "2525"
	}

	mailtrapUser := os.Getenv("MAILTRAP_USER")
	if mailtrapUser == "" {
		mailtrapUser = "username"
	}

	mailtrapPass := os.Getenv("MAILTRAP_PASS")
	if mailtrapPass == "" {
		mailtrapPass = "password"
	}

	mailtrapFrom := os.Getenv("MAILTRAP_FROM")
	if mailtrapFrom == "" {
		mailtrapFrom = "no-reply@example.com"
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	return &Config{
		DBDSN:              os.Getenv("DB_DSN"),
		MongoDSN:           os.Getenv("MONGO_DSN"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AuthGRPCPort:       os.Getenv("AUTH_GRPC_PORT"),
		TreeGRPCPort:       os.Getenv("TREE_GRPC_PORT"),
		FinanceGRPCPort: os.Getenv("FINANCE_GRPC_PORT"),
		XenditAPIKey:    os.Getenv("XENDIT_API_KEY"),
		MailtrapHost:    mailtrapHost,
		MailtrapPort:    mailtrapPort,
		MailtrapUser:    mailtrapUser,
		MailtrapPass:    mailtrapPass,
		MailtrapFrom:    mailtrapFrom,
		RabbitMQURL:     os.Getenv("RABBITMQ_URL"),
	}
}