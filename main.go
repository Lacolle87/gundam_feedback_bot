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

func main() {
	// Загрузка переменных окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла: ", err)
	}

	// Получение токена бота из переменных окружения
	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Токен бота не найден в .env файле.")
	}

	// Получение admin chat IDs из переменных окружения
	adminIDStrings := os.Getenv("ADMIN_IDS")
	if adminIDStrings == "" {
		log.Fatal("Chat IDs не найдены в .env файле.")
	}

	// Разделение строки на отдельные chat IDs
	adminIDStringsSlice := strings.Split(adminIDStrings, ",")
	for _, idStr := range adminIDStringsSlice {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Fatalf("Ошибка при парсинге chat ID: %v", err)
		}
		adminIDs = append(adminIDs, id)
	}

	// Загрузка конфигурации логгера из файла
	configFile := "config/logger_config.json"
	loggerConfig, err := logger.LoadLoggerConfig(configFile)
	if err != nil {
		log.Fatal("Ошибка при загрузке конфигурации логгера:", err)
	}

	// Настройка логгера
	botLogger, err := logger.SetupLogger(loggerConfig)
	if err != nil {
		log.Fatal("Ошибка при инициализации логгера:", err)
	}
	defer botLogger.Close()

	// Создание нового бота с использованием токена
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Ошибка создания бота: ", err)
	}

	// Включение отладочной информации
	bot.Debug = false
	log.Printf("Успешная авторизация на аккаунте %s", bot.Self.UserName)

	if err := loadResponsesFromFile("responses/responses.json"); err != nil {
		log.Fatal("Ошибка при загрузке ответов из файла JSON:", err)
	}

	// Создание объекта для получения обновлений от Telegram
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Получение канала обновлений
	updates, err := bot.GetUpdatesChan(u)

	// Обработка входящих обновлений
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Определение chatID отправителя
		senderChatID := update.Message.Chat.ID

		// Пересылка сообщения каждому chat ID из списка adminIDs
		for _, id := range adminIDs {
			if update.Message.IsCommand() {
				handleCommand(bot, update.Message, botLogger, senderChatID)
			} else if update.Message.Photo != nil {
				handlePhoto(bot, update.Message, botLogger, id, senderChatID)
			} else if update.Message.Text != "" {
				handleText(bot, update.Message, botLogger, id, senderChatID)
			}
		}
	}
}

// Загрузка ответов из JSON
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

// handleCommand обрабатывает команды от пользователей.
// В зависимости от команды, отправляет соответствующее сообщение.
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, botLogger *logger.Logger, senderUserID int64) {
	command := msg.Command()
	chatID := msg.Chat.ID

	switch command {
	case "start":
		response := responses[strings.ToLower(command)]
		sendMessage(bot, chatID, response, botLogger, senderUserID)
	case "info":
		response := responses[strings.ToLower(command)]
		sendMessage(bot, chatID, response, botLogger, senderUserID)
	default:
		sendMessage(bot, chatID, "Неизвестная команда.", botLogger, senderUserID)
	}
}

// handlePhoto обрабатывает входящие фотографии от пользователей.
// Пересылает фотографию в указанный chat ID с информацией об отправителе.
func handlePhoto(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, botLogger *logger.Logger, chatID, senderUserID int64) {
	if msg.Photo != nil && len(*msg.Photo) > 0 {
		photoID := (*msg.Photo)[len(*msg.Photo)-1].FileID
		caption := fmt.Sprintf("Картинка от @%s\n\n%s", msg.From.UserName, msg.Caption)
		forwardPhoto(bot, chatID, photoID, caption, botLogger, senderUserID)
	}
}

// handleText обрабатывает входящие текстовые сообщения от пользователей.
// Пересылает текстовое сообщение в указанный chat ID с информацией об отправителе.
func handleText(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, botLogger *logger.Logger, chatID, senderUserID int64) {
	if msg.Text != "" {
		text := fmt.Sprintf("Текст от @%s\n\n%s", msg.From.UserName, msg.Text)
		sendMessage(bot, chatID, text, botLogger, senderUserID)
	}
}

// sendMessage отправляет сообщение пользователю с указанным chat ID.
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string, botLogger *logger.Logger, senderUserID int64) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка отправки сообщения в чат ID %d: %v", chatID, err))
	} else {
		botLogger.Log(fmt.Sprintf("Сообщение успешно отправлено в чат ID %d", chatID))
	}

	// Отправить подтверждающее сообщение только пользователю, который отправил сообщение
	if chatID != senderUserID {
		sendConfirmationToUser(bot, senderUserID, botLogger)
	}
}

// forwardPhoto пересылает фотографию с подписью в указанный chat ID.
func forwardPhoto(bot *tgbotapi.BotAPI, chatID int64, fileID string, caption string, botLogger *logger.Logger, senderUserID int64) {
	msg := tgbotapi.NewPhotoShare(chatID, fileID)
	msg.Caption = caption
	_, err := bot.Send(msg)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка пересылки сообщения в чат ID %d: %v", chatID, err))
	} else {
		botLogger.Log(fmt.Sprintf("Сообщение успешно переслано в чат ID %d", chatID))
	}

	// Отправить подтверждающее сообщение только пользователю, который отправил фотографию
	if chatID != senderUserID {
		sendConfirmationToUser(bot, senderUserID, botLogger)
	}
}

// sendConfirmationToUser отправляет подтверждающее сообщение пользователю о том, что его сообщение было успешно отправлено.
func sendConfirmationToUser(bot *tgbotapi.BotAPI, chatID int64, botLogger *logger.Logger) {
	text := "Ваше сообщение отправлено, спасибо!"
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		botLogger.Log(fmt.Sprintf("Ошибка отправки подтверждения сообщения в чат ID %d: %v", chatID, err))
	} else {
		botLogger.Log(fmt.Sprintf("Подтверждение сообщения успешно отправлено в чат ID %d", chatID))
	}
}
