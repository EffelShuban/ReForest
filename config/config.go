package config

import (
	"os"
	"time"
)

type Config struct{
	DBURL string
	GRPCPort string
	JWTSecret string
	TokenExpiry time.Duration
}

func LoadConfig() *Config{
	return &Config{
		DBURL: os.Getenv("DATABASE_URL"),
		GRPCPort: ":50051",
		JWTSecret: os.Getenv("JWT_SECRET"),
		TokenExpiry: time.Hour * 24,
	}
}