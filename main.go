package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"gundam_feedback_bot/logger"
	"log"
	"os"
	"strconv"
	"strings"
)

var botToken string
var adminIDs []int64
var responses map[string]string

type BotHandler struct {
	Bot          *tgbotapi.BotAPI
	Logger       *logger.Logger
	SenderChatID int64
}

// Загрузка ответов из файла JSON
func loadResponsesFromFile(filename string) error {
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &responses)
	if err != nil {
		return err
	}

	return nil
}

// Инициализация бота
func newBotHandler(logger *logger.Logger) (*BotHandler, error) {
	// Загрузка .env файла
	err := godotenv.Load()
	if err != nil {
		logger.Log(fmt.Sprintf("ошибка загрузки .env файла: %v", err))
	}

	// Получение BOT_TOKEN
	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		logger.Log("токен бота не найден в .env файле")
		return nil, fmt.Errorf("токен бота не найден в .env файле")
	}

	// Получение ADMIN_IDS
	adminIDStrings := os.Getenv("ADMIN_IDS")
	if adminIDStrings == "" {
		logger.Log("chat IDs не найдены в .env файле")
		return nil, fmt.Errorf("chat IDs не найдены в .env файле")
	}

	adminIDs = make([]int64, 0) // Очищаем пакетную переменную
	for _, idStr := range strings.Split(adminIDStrings, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка при парсинге chat ID: %v", err)
		}
		adminIDs = append(adminIDs, id) // Добавляем обработанные ID в пакетную переменную
	}

	// Создаем экземпляр бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %v", err)
	}

	bot.Debug = false
	logger.Log(fmt.Sprintf("Успешная авторизация на аккаунте %s", bot.Self.UserName))

	// Загрузка ответов из файла JSON
	if err := loadResponsesFromFile("responses/responses.json"); err != nil {
		return nil, fmt.Errorf("ошибка при загрузке ответов из файла JSON: %v", err)
	}

	return &BotHandler{
		Bot:          bot,
		Logger:       logger,
		SenderChatID: 0, // Изменить на chat ID отправителя, если это значение известно заранее
	}, nil
}

// Обработчик входящих обновлений
func (bh *BotHandler) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Пересылка сообщения каждому chat ID из списка adminIDs
		for _, id := range adminIDs {
			if update.Message.IsCommand() {
				bh.handleCommand(update.Message)
			} else if update.Message.Photo != nil {
				bh.handlePhoto(update.Message, id)
			} else if update.Message.Text != "" {
				bh.handleText(update.Message, id)
			}
		}
	}
}

// Обработчик команд от пользователей
func (bh *BotHandler) handleCommand(msg *tgbotapi.Message) {
	command := msg.Command()
	chatID := msg.Chat.ID

	switch command {
	case "start":
		response := responses[strings.ToLower(command)]
		bh.sendMessage(chatID, response)
	case "info":
		response := responses[strings.ToLower(command)]
		bh.sendMessage(chatID, response)
	default:
		bh.sendMessage(chatID, "Неизвестная команда.")
	}
}

// Обработчик входящих текстовых сообщений от пользователей
func (bh *BotHandler) handleText(msg *tgbotapi.Message, chatID int64) {
	if msg.Text != "" {
		text := fmt.Sprintf("Текст от @%s\n\n%s", msg.From.UserName, msg.Text)
		bh.sendMessage(chatID, text)

		// Отправка подтверждающего сообщения только пользователю, который отправил текстовое сообщение
		bh.sendConfirmationToUser(int64(msg.From.ID))
	}
}

// Обработчик входящих фотографий от пользователей
func (bh *BotHandler) handlePhoto(msg *tgbotapi.Message, chatID int64) {
	if msg.Photo != nil && len(*msg.Photo) > 0 {
		photoID := (*msg.Photo)[len(*msg.Photo)-1].FileID
		caption := fmt.Sprintf("Картинка от @%s\n\n%s", msg.From.UserName, msg.Caption)
		bh.forwardPhoto(chatID, photoID, caption)

		// Отправка подтверждающего сообщения только пользователю, который отправил фотографию
		bh.sendConfirmationToUser(int64(msg.From.ID))
	}
}

// Отправка сообщения пользователю с указанным chat ID
func (bh *BotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bh.Bot.Send(msg)
	if err != nil {
		bh.Logger.Log(fmt.Sprintf("Ошибка отправки сообщения в чат ID %d: %v", chatID, err))
	} else {
		bh.Logger.Log(fmt.Sprintf("Сообщение успешно отправлено в чат ID %d", chatID))
	}
}

// Пересылка фотографии с подписью в указанный chat ID
func (bh *BotHandler) forwardPhoto(chatID int64, fileID string, caption string) {
	msg := tgbotapi.NewPhotoShare(chatID, fileID)
	msg.Caption = caption
	_, err := bh.Bot.Send(msg)
	if err != nil {
		bh.Logger.Log(fmt.Sprintf("Ошибка пересылки сообщения в чат ID %d: %v", chatID, err))
	} else {
		bh.Logger.Log(fmt.Sprintf("Сообщение успешно переслано в чат ID %d", chatID))
	}
}

// Отправка подтверждающего сообщения пользователю о том, что его сообщение было успешно отправлено
func (bh *BotHandler) sendConfirmationToUser(chatID int64) {
	if chatID != bh.SenderChatID {
		text := "Ваше сообщение отправлено, спасибо!"
		msg := tgbotapi.NewMessage(chatID, text)
		_, err := bh.Bot.Send(msg)
		if err != nil {
			bh.Logger.Log(fmt.Sprintf("Ошибка отправки подтверждения сообщения в chat ID %d: %v", chatID, err))
		} else {
			bh.Logger.Log(fmt.Sprintf("Подтверждение сообщения успешно отправлено в chat ID %d", chatID))
		}
	}
}

func main() {
	configFile := "config/logger_config.json"

	botLogger, err := logger.InitializeLoggerFromConfig(configFile)
	if err != nil {
		log.Fatal("Ошибка при инициализации логгера:", err)
	}
	defer botLogger.Close()

	bh, err := newBotHandler(botLogger)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка запуска бота: %v", err))
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bh.Bot.GetUpdatesChan(u)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка получения обновлений: %v", err))
	}

	bh.handleUpdates(updates)
}
