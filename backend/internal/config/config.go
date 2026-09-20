package config

import (
	"errors"
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
	R2AccountID              string
	R2AccessKeyID            string
	R2SecretAccessKey        string
	R2BucketName             string
	R2PublicURL              string
}

// Memuat seluruh konfigurasi aplikasi dari environment variable dan memvalidasi konfigurasi wajib.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, errors.New("konfigurasi wajib tidak ditemukan: variabel lingkungan JWT_SECRET harus diatur dan tidak boleh kosong")
	}

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
		JWTSecret:                jwtSecret,
		JWTAccessDurationMinutes: getEnvAsInt("JWT_ACCESS_DURATION_MINUTES", 15),
		JWTRefreshDurationDays:   getEnvAsInt("JWT_REFRESH_DURATION_DAYS", 7),
		R2AccountID:              getEnv("R2_ACCOUNT_ID", "local_dev_account"),
		R2AccessKeyID:            getEnv("R2_ACCESS_KEY_ID", "local_dev_key"),
		R2SecretAccessKey:        getEnv("R2_SECRET_ACCESS_KEY", "local_dev_secret"),
		R2BucketName:             getEnv("R2_BUCKET_NAME", "dcisp-vault"),
		R2PublicURL:              getEnv("R2_PUBLIC_URL", "https://pub-r2.dcisp.internal"),
	}, nil
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
