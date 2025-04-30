package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mot-bot/pkg/mot"
	"mot-bot/pkg/telegram"
	"mot-bot/pkg/ves"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load environment variables
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	motAPIKey := os.Getenv("MOT_API_KEY")
	if motAPIKey == "" {
		log.Fatal("MOT_API_KEY environment variable is not set")
	}

	motClientID := os.Getenv("MOT_CLIENT_ID")
	if motClientID == "" {
		log.Fatal("MOT_CLIENT_ID environment variable is not set")
	}

	motClientSecret := os.Getenv("MOT_CLIENT_SECRET")
	if motClientSecret == "" {
		log.Fatal("MOT_CLIENT_SECRET environment variable is not set")
	}

	vesAPIKey := os.Getenv("VES_API_KEY")
	if vesAPIKey == "" {
		log.Fatal("VES_API_KEY environment variable is not set")
	}

	vesBaseURL := os.Getenv("VES_API_BASE_URL")
	if vesBaseURL == "" {
		log.Fatal("VES_API_BASE_URL environment variable is not set")
	}

	// Create clients
	motClient := mot.NewClient(motClientID, motClientSecret, motAPIKey)
	vesClient := ves.NewClient(vesBaseURL, vesAPIKey)

	// Create bot
	bot, err := telegram.NewBot(token, motClient, vesClient)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create context that will be cancelled on SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Start bot
	log.Println("Starting bot...")
	if err := bot.Start(ctx); err != nil {
		log.Printf("Bot stopped with error: %v", err)
	}
}
