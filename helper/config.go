package helper

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env file")
	}
}

// Config is get env var
func Config(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}
