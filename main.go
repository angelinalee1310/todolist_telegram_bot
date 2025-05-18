package main

import (
	"context"
	"log"
	"strconv"
	"time"

	db "todolist_bot/mongodb"
	"todolist_bot/tasks"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Подключаемся к MongoDB
	client := db.Connect("mongodb+srv://angelinali1310:RRMg8Fxl9uIo2mp6@todolistbotgo.hz0tmef.mongodb.net/?retryWrites=true&w=majority&appName=todolistbotgo")
	collection := client.Database("todolistbotgo").Collection("tasks")

	taskService := tasks.NewTaskService(collection)

	bot, err := tgbotapi.NewBotAPI("7650724062:AAFgaH0xtdW_rlgGtMqPduehkOb9E7R3_Hs")
	if err != nil {
		log.Panic(err)
	}

	// Удаляем webhook, чтобы использовать polling
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Fatalf("Не удалось удалить webhook: %v", err)
	}
	log.Println("Webhook удалён успешно, запускаем polling...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	awaiting := make(map[int64]string)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		switch awaiting[chatID] {
		case "add":
			err := taskService.AddTask(ctx, chatID, text)
			if err != nil {
				log.Println("Ошибка при добавлении задачи:", err)
				send(bot, chatID, "❌ Ошибка при добавлении задачи.")
			} else {
				send(bot, chatID, "✅ Задача добавлена: "+text)
			}
			awaiting[chatID] = ""
			send(bot, chatID, "🤖 Доступные команды: /add /list")
			cancel()
			continue
		}

		cancel()

		switch text {
		case "/add":
			awaiting[chatID] = "add"
			send(bot, chatID, "✏️ Напишите задачу:")
		case "/list":
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			tasks, err := taskService.ListTasks(ctx, chatID)
			cancel()
			if err != nil {
				log.Println("Ошибка при получении задач:", err)
				send(bot, chatID, "❌ Ошибка при получении задач.")
				continue
			}
			if len(tasks) == 0 {
				send(bot, chatID, "📭 Список задач пуст.")
			} else {
				msg := "📋 Ваши задачи:\n"
				for i, t := range tasks {
					status := "❌"
					if t.IsDone {
						status = "✅"
					}
					msg += strconv.Itoa(i+1) + ". " + t.Text + " " + status + "\n"
				}
				send(bot, chatID, msg)
			}
		default:
			send(bot, chatID, "🤖 Доступные команды: /add /list")
		}
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
