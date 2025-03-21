package config

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Processor failed to load .env: %v", err)
	}

	return nil
}
