package config

import "os"

type Config struct {
	DBDSN                    string
	MongoDSN                 string
	JWTSecret                string
	AuthGRPCPort             string
	TreeGRPCPort             string
	AuthServiceURL           string
	TreeManagementServiceURL string

	MailtrapHost string
	MailtrapPort string
	MailtrapUser string
	MailtrapPass string
	MailtrapFrom string
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

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "localhost:50051"
	}

	treeServiceURL := os.Getenv("TREE_MANAGEMENT_SERVICE_URL")
	if treeServiceURL == "" {
		treeServiceURL = "localhost:50052"
	}

	mailtrapHost := os.Getenv("MAILTRAP_HOST")
	mailtrapPort := os.Getenv("MAILTRAP_PORT")
	if mailtrapPort == "" {
		mailtrapPort = "2525"
	}
	mailtrapUser := os.Getenv("MAILTRAP_USER")
	mailtrapPass := os.Getenv("MAILTRAP_PASS")
	mailtrapFrom := os.Getenv("MAILTRAP_FROM")
	if mailtrapFrom == "" {
		mailtrapFrom = "noreply@example.com"
	}

	return &Config{
		DBDSN:                    dsn,
		MongoDSN:                 mongoDSN,
		JWTSecret:                jwtSecret,
		AuthGRPCPort:             authGrpcPort,
		TreeGRPCPort:             treeGrpcPort,
		AuthServiceURL:           authServiceURL,
		TreeManagementServiceURL: treeServiceURL,
		MailtrapHost:             mailtrapHost,
		MailtrapPort:             mailtrapPort,
		MailtrapUser:             mailtrapUser,
		MailtrapPass:             mailtrapPass,
		MailtrapFrom:             mailtrapFrom,
	}
}
