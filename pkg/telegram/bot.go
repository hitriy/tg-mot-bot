package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"mot-bot/pkg/db"
	"mot-bot/pkg/mot"
	"mot-bot/pkg/ves"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxMessageLength = 4096
)

type Bot struct {
	bot       *tgbotapi.BotAPI
	motClient mot.ClientInterface
	vesClient ves.ClientInterface
	logger    *db.Logger
	adminList string
}

func NewBot(bot *tgbotapi.BotAPI, motClient mot.ClientInterface, vesClient ves.ClientInterface, logger *db.Logger, adminList string) *Bot {
	return &Bot{
		bot:       bot,
		motClient: motClient,
		vesClient: vesClient,
		logger:    logger,
		adminList: adminList,
	}
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
		msg.ParseMode = "Markdown"
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
			case "stats":
				if err := b.handleStats(update.Message); err != nil {
					log.Printf("Error handling stats command: %v", err)
				}
			}
		}
	}
}

func (b *Bot) handleRegistration(ctx context.Context, chatID int64, registration string) error {
	// Get data from both APIs concurrently
	motChan := make(chan *mot.VehicleResponse)
	vesChan := make(chan *ves.Vehicle)
	errChan := make(chan error, 2)

	go func() {
		vehicle, err := b.motClient.GetVehicleByRegistration(ctx, registration)
		if err != nil {
			errChan <- fmt.Errorf("MOT API error: %w", err)
			return
		}
		motChan <- vehicle
	}()

	go func() {
		vehicle, err := b.vesClient.GetVehicleByRegistration(ctx, registration)
		if err != nil {
			errChan <- fmt.Errorf("VES API error: %w", err)
			return
		}
		vesChan <- vehicle
	}()

	// Wait for both responses
	var motVehicle *mot.VehicleResponse
	var vesVehicle *ves.Vehicle
	var err error

	for i := 0; i < 2; i++ {
		select {
		case motVehicle = <-motChan:
		case vesVehicle = <-vesChan:
		case err = <-errChan:
			return err
		}
	}

	// Format combined response
	response := formatCombinedResponse(motVehicle, vesVehicle)

	// Get user information
	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}
	chat, err := b.bot.GetChat(chatConfig)

	// Prepare user information for logging
	userID := chatID // Default to chat ID if we can't get user info
	username := "unknown"

	if err == nil {
		if chat.IsPrivate() {
			userID = chat.ID // For private chats, chat ID is the same as user ID
			// Try to get the best available user identifier
			switch {
			case chat.UserName != "":
				username = chat.UserName
			case chat.FirstName != "" && chat.LastName != "":
				username = fmt.Sprintf("%s %s", chat.FirstName, chat.LastName)
			case chat.FirstName != "":
				username = chat.FirstName
			}
		} else {
			// For group chats, try to get the sender's information
			username = fmt.Sprintf("group_chat_%d", chatID)
		}
	}

	// Log the request with user ID
	if err := b.logger.LogRequest(userID, username, registration, response); err != nil {
		log.Printf("Failed to log request: %v", err)
	}

	return b.sendMessage(chatID, response)
}

func (b *Bot) handleStats(message *tgbotapi.Message) error {
	// Check if user is admin
	if !b.isAdmin(message.From.ID, message.From.UserName) {
		return b.sendMessage(message.Chat.ID, "Sorry, this command is only available to administrators.")
	}

	// Get stats from database
	stats, err := b.logger.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	// Format response
	response := fmt.Sprintf("📊 *Bot Usage Statistics*\n\n"+
		"Last 24 hours: `%d` requests\n"+
		"Last 30 days: `%d` requests\n"+
		"All time: `%d` requests",
		stats.LastDay, stats.LastMonth, stats.AllTime)

	return b.sendMessage(message.Chat.ID, response)
}

// isAdmin checks if the given user (by ID or username) is in the admin list
func (b *Bot) isAdmin(userID int64, username string) bool {
	if b.adminList == "" {
		return false
	}

	admins := strings.Fields(b.adminList) // Split by whitespace
	userIDStr := strconv.FormatInt(userID, 10)
	username = strings.TrimPrefix(username, "@") // Remove @ if present

	for _, admin := range admins {
		admin = strings.TrimPrefix(admin, "@") // Remove @ if present
		if admin == username || admin == userIDStr {
			return true
		}
	}

	return false
}

func formatCombinedResponse(motVehicle *mot.VehicleResponse, vesVehicle *ves.Vehicle) string {
	var sb strings.Builder

	// Basic vehicle info
	sb.WriteString(fmt.Sprintf("🚗 *Vehicle Information*\n\n"))
	sb.WriteString(fmt.Sprintf("📝 *Registration:* `%s`\n", motVehicle.Registration))
	sb.WriteString(fmt.Sprintf("🏭 *Make:* `%s`\n", motVehicle.Make))
	sb.WriteString(fmt.Sprintf("🚘 *Model:* `%s`\n", motVehicle.Model))
	sb.WriteString(fmt.Sprintf("📅 *First Registered:* `%s`\n", motVehicle.FirstUsedDate))
	sb.WriteString(fmt.Sprintf("⛽ *Fuel Type:* `%s`\n", motVehicle.FuelType))
	sb.WriteString(fmt.Sprintf("🎨 *Colour:* `%s`\n", motVehicle.PrimaryColour))
	sb.WriteString(fmt.Sprintf("🛞 *Wheelplan:* `%s`\n", vesVehicle.Wheelplan))
	sb.WriteString(fmt.Sprintf("🌍 *Euro Status:* `%s`\n", vesVehicle.EuroStatus))
	sb.WriteString(fmt.Sprintf("📄 *Last V5C Issued:* `%s`\n", vesVehicle.DateOfLastV5CIssued.Format("02.01.2006")))

	// Tax information
	sb.WriteString("\n💰 *Tax Information*\n\n")
	sb.WriteString(fmt.Sprintf("📊 *Status:* `%s`\n", vesVehicle.TaxStatus))
	if !vesVehicle.TaxDueDate.IsZero() {
		sb.WriteString(fmt.Sprintf("📅 *Due Date:* `%s`\n", vesVehicle.TaxDueDate.Format("02.01.2006")))
	}

	// MOT history
	sb.WriteString("\n🔧 *MOT History*\n\n")
	for _, test := range motVehicle.MotTests {
		// Format date to DD.MM.YYYY
		testDate := test.CompletedDate
		if len(testDate) >= 10 {
			// Parse the date from YYYY-MM-DD format
			parsedDate, err := time.Parse("2006-01-02", testDate[:10])
			if err == nil {
				testDate = parsedDate.Format("02.01.2006")
			}
		}
		sb.WriteString(fmt.Sprintf("📅 *Test Date:* `%s`\n", testDate))

		// Set appropriate emoji for test result
		var resultEmoji string
		if strings.ToUpper(test.TestResult) == "FAILED" {
			resultEmoji = "❌"
		} else {
			resultEmoji = "✅"
		}
		sb.WriteString(fmt.Sprintf("%s *Result:* `%s`\n", resultEmoji, test.TestResult))

		if test.OdometerValue != "" {
			sb.WriteString(fmt.Sprintf("📏 *Mileage:* `%s %s`\n", test.OdometerValue, test.OdometerUnit))
		}
		if len(test.Defects) > 0 {
			sb.WriteString("⚠️ *Defects:*\n")
			for _, defect := range test.Defects {
				// Convert defect type to emoji
				var defectEmoji string
				switch defect.Type {
				case "FAIL":
					defectEmoji = "❌"
				case "ADVISORY":
					defectEmoji = "⚠️"
				case "USER ENTERED":
					defectEmoji = "📝"
				default:
					defectEmoji = "ℹ️"
				}
				sb.WriteString(fmt.Sprintf("  %s `%s`\n", defectEmoji, defect.Text))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
