package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mot-bot/pkg/mot"
	"mot-bot/pkg/telegram"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get Telegram bot token
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Get MOT API credentials
	clientID := os.Getenv("MOT_CLIENT_ID")
	if clientID == "" {
		log.Fatal("MOT_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("MOT_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("MOT_CLIENT_SECRET environment variable is required")
	}

	apiKey := os.Getenv("MOT_API_KEY")
	if apiKey == "" {
		log.Fatal("MOT_API_KEY environment variable is required")
	}

	// Create MOT client
	motClient := mot.NewClient(clientID, clientSecret, apiKey)

	// Create and start bot
	bot, err := telegram.NewBot(token, motClient)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create context that will be canceled on SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	// Start the bot
	log.Println("Starting bot...")
	if err := bot.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Bot error: %v", err)
	}
}
