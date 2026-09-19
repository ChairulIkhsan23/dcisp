package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	GinMode                  string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	DBSSLMode                string
	RedisHost                string
	RedisPort                string
	RedisPassword            string
	JWTSecret                string
	JWTAccessDurationMinutes int
	JWTRefreshDurationDays   int
}

// Memuat seluruh konfigurasi aplikasi dari environment variable dan file .env.
func LoadConfig() *Config {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:                     getEnv("PORT", "8080"),
		GinMode:                  getEnv("GIN_MODE", "debug"),
		DBHost:                   getEnv("DB_HOST", "127.0.0.1"),
		DBPort:                   getEnv("DB_PORT", "5432"),
		DBUser:                   getEnv("DB_USER", "postgres"),
		DBPassword:               getEnv("DB_PASSWORD", "postgres_secret_pass"),
		DBName:                   getEnv("DB_NAME", "dcisp_db"),
		DBSSLMode:                getEnv("DB_SSLMODE", "disable"),
		RedisHost:                getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:                getEnv("REDIS_PORT", "6381"),
		RedisPassword:            getEnv("REDIS_PASSWORD", ""),
		JWTSecret:                getEnv("JWT_SECRET", "super_secret_jwt_key_dcisp_2026_32bytes_long!"),
		JWTAccessDurationMinutes: getEnvAsInt("JWT_ACCESS_DURATION_MINUTES", 15),
		JWTRefreshDurationDays:   getEnvAsInt("JWT_REFRESH_DURATION_DAYS", 7),
	}
}

// Mengambil nilai environment variable berdasarkan kunci dengan nilai bawaan sebagai cadangan.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

// Mengambil nilai environment variable dan mengonversinya menjadi tipe integer.
func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
