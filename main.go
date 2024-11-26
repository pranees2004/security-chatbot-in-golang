package main

import (
	"SecurityChatbot/src"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Setting up database...")
	if err := src.SetupDatabase(); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	log.Println("Database setup complete.")
	log.Println("Loading intents...")
	err := src.LoadIntents("data.json")
	if err != nil {
		log.Fatalf("Error loading intents: %v", err)
	}
	log.Println("Intents loaded.")
	log.Println("Setting up chatbot...")
	client := src.MakeWaClient()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
