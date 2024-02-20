package helper

import "os"

// Config is get env var
func Config(key string) string {
	return os.Getenv(key)
}
