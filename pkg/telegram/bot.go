package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	"mot-bot/pkg/mot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot       *tgbotapi.BotAPI
	motClient mot.ClientInterface
}

func NewBot(token string, motClient mot.ClientInterface) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{
		bot:       bot,
		motClient: motClient,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if !update.Message.IsCommand() {
				// Handle registration number
				registration := strings.TrimSpace(update.Message.Text)
				if err := b.handleRegistration(ctx, update.Message.Chat.ID, registration); err != nil {
					log.Printf("Error handling registration: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, I couldn't process that registration number. Please try again.")
					if _, err := b.bot.Send(msg); err != nil {
						log.Printf("Error sending error message: %v", err)
					}
				}
				continue
			}

			// Handle commands
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the MOT Checker Bot! Send me a UK vehicle registration number to check its MOT history.")
				if _, err := b.bot.Send(msg); err != nil {
					log.Printf("Error sending start message: %v", err)
				}
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Simply send me a UK vehicle registration number to check its MOT history.")
				if _, err := b.bot.Send(msg); err != nil {
					log.Printf("Error sending help message: %v", err)
				}
			}
		}
	}
}

func (b *Bot) handleRegistration(ctx context.Context, chatID int64, registration string) error {
	vehicle, err := b.motClient.GetVehicleByRegistration(ctx, registration)
	if err != nil {
		return fmt.Errorf("failed to get vehicle data: %w", err)
	}

	// Use the new formatted output
	response := vehicle.FormatVehicleInfo()

	msg := tgbotapi.NewMessage(chatID, response)
	if _, err := b.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
