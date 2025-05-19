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

// func sendWithKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string) {
// 	keyboard := tgbotapi.NewReplyKeyboard(
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton("/add"),
// 			tgbotapi.NewKeyboardButton("/list"),
// 		),
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton("/done"),
// 			tgbotapi.NewKeyboardButton("/delete"),
// 		),
// 	)

// 	msg := tgbotapi.NewMessage(chatID, text)
// 	msg.ReplyMarkup = keyboard
// 	bot.Send(msg)
// }

func sendWithInlineButtons(bot *tgbotapi.BotAPI, chatID int64, text string) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить задачу", "add"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Список задач", "list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Завершить задачу", "done"),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить задачу", "delete"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func main() {
	// Подключаемся к MongoDB
	client := db.Connect("mongodb+srv://angelinali1310:RRMg8Fxl9uIo2mp6@todolistbotgo.hz0tmef.mongodb.net/?retryWrites=true&w=majority&appName=todolistbotgo")
	collection := client.Database("todolistbotgo").Collection("tasks") // автоматически создает tasks

	taskService := tasks.NewTaskService(collection)

	bot, err := tgbotapi.NewBotAPI("7650724062:AAFgaH0xtdW_rlgGtMqPduehkOb9E7R3_Hs")
	if err != nil { // обрабатываем ошибку
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
	updates := bot.GetUpdatesChan(u) // получаем канал

	awaiting := make(map[int64]string) // "add", "delete" или ""

	for update := range updates { // читаем канал пока он не закроется
		// 1) обработка нажатий на Inline-кнопки
		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data // "add", "list", "done", "delete"

			// Обязательно «отвечаем» на Callback, чтобы убрать часы
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
			switch data {
			case "add":
				awaiting[chatID] = "add"
				send(bot, chatID, "✏️ Напишите задачу:")
			case "list":
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				tasks, err := taskService.ListTasks(ctx, chatID)
				cancel()
				if err != nil {
					send(bot, chatID, "❌ Ошибка при получении задач.")
					continue
				}
				if len(tasks) == 0 {
					send(bot, chatID, "📭 Список задач пуст.")
				} else {
					var msg string = "📋 Ваши задачи:\n"
					for i, t := range tasks {
						status := "❌"
						if t.IsDone {
							status = "✅"
						}
						msg += strconv.Itoa(i+1) + ". " + t.Text + " " + status + "\n"
					}
					send(bot, chatID, msg)
				}
				sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
			case "done":
				awaiting[chatID] = "done"
				send(bot, chatID, "☑️ Введите номер задачи, которую вы завершили:")
			case "delete":
				awaiting[chatID] = "delete"
				send(bot, chatID, "❓ Введите номер задачи для удаления:")
			default:
				sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
			}

			// после обработки callback возвращаемся к чтению канала
			continue
		}

		// 2) проверяем текстовые сообщения
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		// Если пользователь новый — покажем список задач сразу
		if _, ok := awaiting[chatID]; !ok {
			awaiting[chatID] = "" // инициализируем

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			tasks, err := taskService.ListTasks(ctx, chatID)
			cancel()

			if err != nil {
				log.Println("Ошибка при получении задач:", err)
				send(bot, chatID, "❌ Ошибка при получении задач.")
			} else if len(tasks) == 0 {
				send(bot, chatID, "👋 Добро пожаловать! У вас пока нет задач.")
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
			sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
		}

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
			sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")

			cancel()
			continue
		case "done":
			tasks, err := taskService.ListTasks(ctx, chatID)
			if err != nil {
				log.Println("Ошибка при done", err)
			}
			index, err := strconv.Atoi(text)
			if err != nil || index <= 0 || index > len(tasks) {
				send(bot, chatID, "❌ Неверный номер задачи.")
			} else {
				taskToUpdate := tasks[index-1]
				err = taskService.MarkTaskDone(ctx, taskToUpdate.ID)
				if err != nil {
					// handle error
				} else {
					send(bot, chatID, "🎉 Задача отмечена как выполненная!")
				}
				sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
			}
			awaiting[chatID] = ""
			continue
		case "delete":
			tasks, err := taskService.ListTasks(ctx, chatID)
			if err != nil {
				log.Println("Ошибка при delete", err)
			}
			index, err := strconv.Atoi(text)
			if err != nil || index <= 0 || index > len(tasks) {
				send(bot, chatID, "❌ Неверный номер задачи.")
			} else {
				taskTRemoved := tasks[index-1]
				err = taskService.RemoveTask(ctx, taskTRemoved.ID)
				if err != nil {
					// handle error
				} else {
					send(bot, chatID, "🗑 Удалена: "+taskTRemoved.Text)
				}
				sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
			}
			awaiting[chatID] = ""
			continue
		}

		cancel()

		// switch text {
		// case "/add":
		// 	awaiting[chatID] = "add"
		// 	send(bot, chatID, "✏️ Напишите задачу:")
		// case "/list":
		// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// 	tasks, err := taskService.ListTasks(ctx, chatID)
		// 	cancel()
		// 	if err != nil {
		// 		log.Println("Ошибка при получении задач:", err)
		// 		send(bot, chatID, "❌ Ошибка при получении задач.")
		// 		continue
		// 	}
		// 	if len(tasks) == 0 {
		// 		send(bot, chatID, "📭 Список задач пуст.")
		// 	} else {
		// 		msg := "📋 Ваши задачи:\n"
		// 		for i, t := range tasks {
		// 			status := "❌"
		// 			if t.IsDone {
		// 				status = "✅"
		// 			}
		// 			msg += strconv.Itoa(i+1) + ". " + t.Text + " " + status + "\n"
		// 		}
		// 		send(bot, chatID, msg)
		// 		sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
		// 	}
		// case "/done":
		// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// 	tasks, err := taskService.ListTasks(ctx, chatID)
		// 	cancel()
		// 	if err != nil {
		// 		log.Println("Ошибка при /done:", err)
		// 		send(bot, chatID, "❌ Ошибка при /done.")
		// 		continue
		// 	}
		// 	if len(tasks) == 0 {
		// 		send(bot, chatID, "📭 У вас нет задач.")
		// 	} else {
		// 		awaiting[chatID] = "done"
		// 		send(bot, chatID, "☑️ Введите номер задачи, которую вы завершили:")
		// 	}
		// case "/delete":
		// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// 	tasks, err := taskService.ListTasks(ctx, chatID)
		// 	cancel()
		// 	if err != nil {
		// 		log.Println("Ошибка при /delete:", err)
		// 		send(bot, chatID, "❌ Ошибка при /delete.")
		// 		continue
		// 	}
		// 	if len(tasks) == 0 {
		// 		send(bot, chatID, "📭 У вас нет задач для удаления.")
		// 	} else {
		// 		awaiting[chatID] = "delete"
		// 		send(bot, chatID, "❓ Введите номер задачи для удаления:")
		// 	}
		// default:
		// 	sendWithInlineButtons(bot, chatID, "🤖 Доступные действия:")
		// }
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
