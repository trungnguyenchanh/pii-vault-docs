package config

import (
	"os"
)

// Config chứa cấu hình runtime của Admin API, nạp từ biến môi trường.
type Config struct {
	Addr               string
	DatabaseURL        string
	JWKSURL            string
	JWTIssuer          string
	JWTAudience        string
	CORSAllowedOrigins string
}

// Load đọc cấu hình từ env, áp giá trị mặc định khi cần.
func Load() Config {
	return Config{
		Addr:               env("ADMIN_API_ADDR", ":8080"),
		DatabaseURL:        env("DATABASE_URL", ""),
		JWKSURL:            env("JWKS_URL", ""),
		JWTIssuer:          env("JWT_ISSUER", ""),
		JWTAudience:        env("JWT_AUDIENCE", "pii-admin"),
		CORSAllowedOrigins: env("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
