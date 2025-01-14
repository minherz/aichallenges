package utils

import (
	"os"
)

func GetEnvOrDefault(name, defaultValue string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return defaultValue
}
