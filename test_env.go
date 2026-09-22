package main

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
)

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	godotenv.Load(".env.local", ".env")
	fmt.Printf("LLM_PROVIDER: '%s'\n", getEnvOrDefault("LLM_PROVIDER", ""))
	fmt.Printf("LLM_BASE_URL: '%s'\n", getEnvOrDefault("LLM_BASE_URL", "fallback"))
}
