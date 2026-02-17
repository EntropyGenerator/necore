package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Config(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file: ", err)
	}
	return os.Getenv(key)
}
