# MOT Checker Telegram Bot

A Telegram bot that checks UK vehicle MOT history using the DVSA MOT History API.

## Prerequisites

- Go 1.21 or later
- A Telegram bot token (get it from [@BotFather](https://t.me/BotFather))
- A DVSA MOT History API key and OAuth2 credentials (register at [DVSA Developer Portal](https://developer-portal.driver-vehicle-licensing.api.gov.uk/))

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/mot-bot.git
cd mot-bot
```

2. Install dependencies:
```bash
go mod tidy
```

3. Create a `.env` file in the project root with the following content:
```bash
# Telegram Bot Token (get from @BotFather)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here

# MOT API Credentials
MOT_CLIENT_ID=your_client_id_here
MOT_CLIENT_SECRET=your_client_secret_here
MOT_API_KEY=your_api_key_here
MOT_TOKEN_URL=your_token_url_here
```

Alternatively, you can set these as environment variables:
```bash
export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
export MOT_CLIENT_ID="your_client_id"
export MOT_CLIENT_SECRET="your_client_secret"
export MOT_API_KEY="your_api_key"
export MOT_TOKEN_URL="your_token_url_here"
```

## Running the Bot

```bash
go run cmd/bot/main.go
```

## Usage

1. Start a chat with your bot on Telegram
2. Send `/start` to get a welcome message
3. Send a UK vehicle registration number (e.g., "AB12CDE") to get its MOT history

## Features

- Retrieves vehicle details including make, model, and registration date
- Shows complete MOT history with test dates and results
- Displays any defects found during MOT tests
- Handles errors gracefully with user-friendly messages

## Commands

- `/start` - Get a welcome message
- `/help` - Get usage instructions

## License

MIT 