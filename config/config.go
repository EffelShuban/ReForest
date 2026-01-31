package config

import "os"

type Config struct {
	DBDSN        string
	MongoDSN     string
	JWTSecret    string
	AuthGRPCPort string
	TreeGRPCPort string
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

	return &Config{
		DBDSN:        dsn,
		MongoDSN:     mongoDSN,
		JWTSecret:    jwtSecret,
		AuthGRPCPort: authGrpcPort,
		TreeGRPCPort: treeGrpcPort,
	}
}