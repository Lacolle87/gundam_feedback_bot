package bot

import (
	"fmt"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gundam_feedback_bot/loader"
)

type Handler struct {
	Bot          *tgbotapi.BotAPI
	SenderChatID int64
}

// NewBotHandler Инициализация бота
func NewBotHandler() (*Handler, error) {
	// Загрузка .env файла
	if err := loader.LoadEnv(); err != nil {
		return nil, err
	}

	// Создаем экземпляр бота
	bot, err := tgbotapi.NewBotAPI(loader.BotToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %v", err)
	}

	bot.Debug = false
	loader.BotLogger.Log(fmt.Sprintf("Успешная авторизация на аккаунте %s", bot.Self.UserName))

	// Загрузка ответов из файла JSON
	if err := loader.LoadResponsesFromFile("responses/responses.json"); err != nil {
		return nil, fmt.Errorf("ошибка при загрузке ответов из файла JSON: %v", err)
	}

	return &Handler{
		Bot:          bot,
		SenderChatID: 0, // Изменить на chat ID отправителя, если это значение известно заранее
	}, nil
}

// HandleUpdates Обработчик входящих обновлений
func (bh *Handler) HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Пересылка сообщения каждому chat ID из списка adminIDs
		for _, id := range loader.AdminIDs {
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
func (bh *Handler) handleCommand(msg *tgbotapi.Message) {
	command := msg.Command()
	chatID := msg.Chat.ID

	switch command {
	case "start":
		response := loader.Responses[strings.ToLower(command)]
		bh.sendMessage(chatID, response)
	case "info":
		response := loader.Responses[strings.ToLower(command)]
		bh.sendMessage(chatID, response)
	default:
		bh.sendMessage(chatID, "Неизвестная команда.")
	}
}

// Обработчик входящих текстовых сообщений от пользователей
func (bh *Handler) handleText(msg *tgbotapi.Message, chatID int64) {
	if msg.Text != "" {
		text := fmt.Sprintf("Текст от @%s\n\n%s", msg.From.UserName, msg.Text)
		bh.sendMessage(chatID, text)

		// Отправка подтверждающего сообщения только пользователю, который отправил текстовое сообщение
		bh.sendConfirmationToUser(int64(msg.From.ID))
	}
}

// Обработчик входящих фотографий от пользователей
func (bh *Handler) handlePhoto(msg *tgbotapi.Message, chatID int64) {
	if msg.Photo != nil && len(*msg.Photo) > 0 {
		photoID := (*msg.Photo)[len(*msg.Photo)-1].FileID
		caption := fmt.Sprintf("Картинка от @%s\n\n%s", msg.From.UserName, msg.Caption)
		bh.forwardPhoto(chatID, photoID, caption)

		// Отправка подтверждающего сообщения только пользователю, который отправил фотографию
		bh.sendConfirmationToUser(int64(msg.From.ID))
	}
}

// Отправка сообщения пользователю с указанным chat ID
func (bh *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bh.Bot.Send(msg)
	if err != nil {
		loader.BotLogger.Log(fmt.Sprintf("Ошибка отправки сообщения в чат ID %d: %v", chatID, err))
	} else {
		loader.BotLogger.Log(fmt.Sprintf("Сообщение успешно отправлено в чат ID %d", chatID))
	}
}

// Пересылка фотографии с подписью в указанный chat ID
func (bh *Handler) forwardPhoto(chatID int64, fileID string, caption string) {
	msg := tgbotapi.NewPhotoShare(chatID, fileID)
	msg.Caption = caption
	_, err := bh.Bot.Send(msg)
	if err != nil {
		loader.BotLogger.Log(fmt.Sprintf("Ошибка пересылки сообщения в чат ID %d: %v", chatID, err))
	} else {
		loader.BotLogger.Log(fmt.Sprintf("Сообщение успешно переслано в чат ID %d", chatID))
	}
}

// Отправка подтверждающего сообщения пользователю о том, что его сообщение было успешно отправлено
func (bh *Handler) sendConfirmationToUser(chatID int64) {
	if chatID != bh.SenderChatID {
		text := "Ваше сообщение отправлено, спасибо!"
		msg := tgbotapi.NewMessage(chatID, text)
		_, err := bh.Bot.Send(msg)
		if err != nil {
			loader.BotLogger.Log(fmt.Sprintf("Ошибка отправки подтверждения сообщения в chat ID %d: %v", chatID, err))
		} else {
			loader.BotLogger.Log(fmt.Sprintf("Подтверждение сообщения успешно отправлено в chat ID %d", chatID))
		}
	}
}
