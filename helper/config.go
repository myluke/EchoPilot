package helper

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}

// Config is get env var
func Config(key string) string {
	return os.Getenv(key)
}
