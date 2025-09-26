package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	WeatherAPIKey  string
	WeatherStackKey string
	
	DatabasePath string
	
	ServerPort string
	DebugMode  bool
	
	MaxRequests int
	WaitTime    time.Duration
	APITimeout time.Duration
}

func LoadConfig() *Config {

	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env dosyası bulunamadı, sistem environment variables kullanılıyor: %v", err)
	}
	
	config := &Config{
		WeatherAPIKey:   getEnv("WEATHER_API_KEY", "b417cbe563c444f98a0124504252409"),
		WeatherStackKey: getEnv("WEATHER_STACK_KEY", "e8919ef8c0246a634fb92cf4567c3681"),
		
		DatabasePath: getEnv("DATABASE_PATH", "weather.sqlite"),
		
		ServerPort: getEnv("SERVER_PORT", "8000"),
		DebugMode:  getEnvAsBool("DEBUG_MODE", false),
		
		MaxRequests: getEnvAsInt("MAX_REQUESTS", 10),
		WaitTime:    getEnvAsDuration("WAIT_TIME", "5s"),
		
		APITimeout: getEnvAsDuration("API_TIMEOUT", "10s"),
	}
	
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	
	return 5 * time.Second // Fallback
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
