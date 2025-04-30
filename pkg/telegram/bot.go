package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	"mot-bot/pkg/mot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxMessageLength = 4096
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

// splitMessage splits a long message into chunks of maxMessageLength characters
func (b *Bot) splitMessage(text string) []string {
	if len(text) <= maxMessageLength {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")
	currentChunk := strings.Builder{}

	for _, line := range lines {
		// If adding this line would exceed the limit, start a new chunk
		if currentChunk.Len()+len(line)+1 > maxMessageLength {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		// Add the line to the current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)
	}

	// Add the last chunk if it's not empty
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// sendMessage sends a message to a chat, splitting it into chunks if necessary
func (b *Bot) sendMessage(chatID int64, text string) error {
	chunks := b.splitMessage(text)
	for i, chunk := range chunks {
		msg := tgbotapi.NewMessage(chatID, chunk)
		if i > 0 {
			msg.Text = fmt.Sprintf("(Part %d/%d)\n%s", i+1, len(chunks), chunk)
		}
		if _, err := b.bot.Send(msg); err != nil {
			return fmt.Errorf("failed to send message part %d: %w", i+1, err)
		}
	}
	return nil
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
					if err := b.sendMessage(update.Message.Chat.ID, "Sorry, I couldn't process that registration number. Please try again."); err != nil {
						log.Printf("Error sending error message: %v", err)
					}
				}
				continue
			}

			// Handle commands
			switch update.Message.Command() {
			case "start":
				if err := b.sendMessage(update.Message.Chat.ID, "Welcome to the MOT Checker Bot! Send me a UK vehicle registration number to check its MOT history."); err != nil {
					log.Printf("Error sending start message: %v", err)
				}
			case "help":
				if err := b.sendMessage(update.Message.Chat.ID, "Simply send me a UK vehicle registration number to check its MOT history."); err != nil {
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

	return b.sendMessage(chatID, response)
}
