# Gundam Feedback Bot

Gundam Feedback Bot is a simple Telegram bot designed to forward messages and photos from users to specified chat IDs. It serves as a convenient tool for collecting feedback and sharing media with a group of people.

## Features

- **Message Forwarding**: The bot forwards text messages sent by users to the specified chat IDs. It also sends a confirmation message to the original sender.

- **Photo Forwarding**: When users send photos, the bot forwards the photos with captions to the designated chat IDs. It also sends a confirmation message to the sender.

- **Command Handling**: The bot handles specific commands:
    - `/start`: Sends a welcome message to the user.
    - `/info`: Provides information about the bot.

## How to Use

To use the Gundam Feedback Bot, follow these steps:

1. **Create a Telegram Bot**: If you don't have a Telegram bot already, create one by following the [official Telegram documentation](https://core.telegram.org/bots#creating-a-new-bot).

2. **Set Up Environment Variables**: Create a `.env` file and add the following variables:
   BOT_TOKEN=your_telegram_bot_token
   CHAT_IDS=comma_separated_chat_ids

3. **Specify Chat IDs**: Replace `comma_separated_chat_ids` in the `.env` file with the chat IDs of the groups or individuals to which you want to forward the messages and photos. Separate multiple chat IDs with commas.

4. **Install Dependencies**: Make sure you have Go installed on your system. Navigate to the project directory and run:
   go mod download

5. **Configure Logger**: Edit the `config/logger_config.json` file to customize the logging behavior, such as log file path, log level, etc.

6. **Load Responses**: Customize the responses to commands in the `responses/responses.json` file.

7. **Run the Bot**: Execute the following command to start the bot:
   go run main.go


## Disclaimer

This Telegram bot is developed for educational and non-commercial purposes. The bot may not be suitable for production environments without further improvements and security considerations.

---

 🤖🚀
