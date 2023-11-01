package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	rootPath := os.Getenv("ROOT_PATH")
	if rootPath == "" {
		log.Fatal("MONGO_URI not set in environment")
	}
	indexerEngine(rootPath)
}
