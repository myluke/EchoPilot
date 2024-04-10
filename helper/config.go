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
func Config(key string) string {
	return os.Getenv(key)
}
