package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var BotToken string
var AdminIDs []int64
var Responses map[string]string

// LoadResponsesFromFile Загрузка ответов из файла JSON
func LoadResponsesFromFile(filename string) error {
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &Responses)
	if err != nil {
		return err
	}

	return nil
}

// LoadEnv Загружает значения из файла .env и устанавливает необходимые переменные окружения.
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("ошибка загрузки .env файла: %v", err)
	}

	BotToken = os.Getenv("BOT_TOKEN")
	if BotToken == "" {
		return fmt.Errorf("токен бота не найден в .env файле")
	}

	adminIDStrings := os.Getenv("ADMIN_IDS")
	if adminIDStrings == "" {
		return fmt.Errorf("chat IDs не найдены в .env файле")
	}

	AdminIDs = make([]int64, 0) // Очищаем пакетную переменную
	for _, idStr := range strings.Split(adminIDStrings, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("ошибка при парсинге chat ID: %v", err)
		}
		AdminIDs = append(AdminIDs, id) // Добавляем обработанные ID в пакетную переменную
	}

	return nil
}
