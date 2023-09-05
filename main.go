package main

import (
	"fmt"
	"gundam_feedback_bot/loader"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gundam_feedback_bot/bot"
)

func main() {
	err := loader.LoadLoggerFromConfig()
	if err != nil {
		log.Println(err)
	}

	bh, err := bot.NewBotHandler()
	if err != nil {
		loader.BotLogger.Log(fmt.Sprintf("Ошибка запуска бота: %v", err))
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bh.Bot.GetUpdatesChan(u)
	if err != nil {
		loader.BotLogger.Log(fmt.Sprintf("Ошибка получения обновлений: %v", err))
	}

	bh.HandleUpdates(updates)
}
