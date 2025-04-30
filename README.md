# MOT Checker Telegram Bot

A Telegram bot that checks MOT history and vehicle details for UK-registered vehicles.

## Features

- Check MOT history for any UK-registered vehicle
- View vehicle tax status and due date
- View vehicle wheelplan and Euro status
- View date of last V5C issued

## Prerequisites

- Go 1.21 or later
- A Telegram bot token (get it from [@BotFather](https://t.me/BotFather))
- A UK MOT API key (get it from [GOV.UK](https://developer-portal.driver-vehicle-licensing.api.gov.uk/))
- A UK Vehicle Enquiry Service API key (get it from [GOV.UK](https://developer-portal.driver-vehicle-licensing.api.gov.uk/))

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

MOT_API_KEY=your_mot_api_key
MOT_CLIENT_ID=your_mot_client_id
MOT_CLIENT_SECRET=your_mot_client_secret
MOT_TOKEN_URL=you_mot_token_url

VES_API_KEY=your_ves_api_key
VES_API_BASE_URL=https://driver-vehicle-licensing.api.gov.uk/vehicle-enquiry/v1
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/hitriy/tg-mot-bot.git
cd cd-mot-bot
```

2. Install dependencies:
```bash
go mod tody
```

3. Create a `.env` file with your API keys and tokens.

4. Run the bot:
```bash
go run cmd/bot/main.go
```

## Usage

1. Start a chat with your bot on Telegram
2. Send a UK vehicle registration number
3. The bot will respond with:
   - MOT history
   - Tax status and due date
   - Wheelplan
   - Euro status
   - Date of last V5C issued

## License

MIT 