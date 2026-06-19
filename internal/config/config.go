package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret    string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	Port         string
	CORSOrigin   []string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	MailFrom     string
}

func requireEnv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(key + " is required")

	}
	return value
}

func LoadEnv() *Config {
	_ = godotenv.Load()

	corsOrigins := os.Getenv("CORS_ORIGINS")
	var origins []string
	if corsOrigins != "" {
		origins = strings.Split(corsOrigins, ",")
	}

	return &Config{
		JWTSecret:    requireEnv("JWT_SECRET"),
		DBHost:       requireEnv("DB_HOST"),
		DBPort:       requireEnv("DB_PORT"),
		DBUser:       requireEnv("DB_USER"),
		DBPassword:   requireEnv("DB_PASSWORD"),
		DBName:       requireEnv("DB_NAME"),
		Port:         requireEnv("PORT"),
		CORSOrigin:   origins,
		SMTPHost:     requireEnv("SMTP_HOST"),
		SMTPPort:     requireEnv("SMTP_PORT"),
		SMTPUsername: requireEnv("SMTP_USERNAME"),
		SMTPPassword: requireEnv("SMTP_PASSWORD"),
		MailFrom:     requireEnv("MAIL_FROM"),
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
