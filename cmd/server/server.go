package main

import (
	"github.com/evok02/jcrawler/internal/server"
	"github.com/joho/godotenv"
	"log"
)

// TODO: fix issues with handlers

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Running...")
	log.Fatal(server.Run("localhost:1337").Error())
}
