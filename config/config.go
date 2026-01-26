package config

import "os"

type Config struct {
	DBDSN        string
	JWTSecret    string
	AuthGRPCPort string
}

func Load() *Config {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=forest_guard port=5432 sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key"
	}

	authGrpcPort := os.Getenv("AUTH_GRPC_PORT")
	if authGrpcPort == "" {
		authGrpcPort = ":50051"
	}

	return &Config{
		DBDSN:        dsn,
		JWTSecret:    jwtSecret,
		AuthGRPCPort: authGrpcPort,
	}
}