package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gundam_feedback_bot/bot"
	"gundam_feedback_bot/logger"
	"log"
)

const configFile = "config/logger_config.json"

func main() {
	botLogger, err := logger.InitializeLoggerFromConfig(configFile)
	if err != nil {
		log.Fatal("Ошибка при инициализации логгера:", err)
	}
	defer func(botLogger *logger.Logger) {
		err := botLogger.Close()
		if err != nil {
			log.Fatal("Ошибка при закрытие логгера:", err)
		}
	}(botLogger)

	bh, err := bot.NewBotHandler(botLogger)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка запуска бота: %v", err))
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bh.Bot.GetUpdatesChan(u)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка получения обновлений: %v", err))
	}

	bh.HandleUpdates(updates)
}
