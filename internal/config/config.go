// config package provides loading configuration settings
package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARN: Could not load .env file: %v. Relying on OS environment variables.", err)
	} else {
		log.Printf("INFO: Loaded configuration from .env file.")
	}
}