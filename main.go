package main

import (
	"log"
	"strconv"

	_ "github.com/lib/pq"

	"todolist_bot/supabase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// type Task struct {
// 	Text     string
// 	IsDone bool
// }

var awaiting = make(map[int64]string) // "add", "delete" или ""
// var tasks = make(map[int64][]Task)

func main() {
	bot, err := tgbotapi.NewBotAPI("7650724062:AAFgaH0xtdW_rlgGtMqPduehkOb9E7R3_Hs")
	if err != nil { // обрабатываем ошибку
		log.Panic(err)
	}

	// настройка обновлений от Telegram
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u) // получаем канал

	for update := range updates { // читаем канал пока он не закроется
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text
		sc := supabase.NewSupabaseClient()

		switch awaiting[chatID] {
		case "add":
			// tasks[chatID] = append(tasks[chatID], text)
			// tasks[chatID] = append(tasks[chatID], Task{Text: text, IsDone: false})
			err := sc.AddTask(chatID, text)
			if err != nil {
				log.Fatal("Ошибка при добавлении задачи:", err)
			}
			awaiting[chatID] = ""
			send(bot, chatID, "✅ Задача добавлена: "+text)
			send(bot, chatID, "🤖 Доступные команды: /add /list /delete /done")
			continue

			// case "delete":
			// 	index, err := strconv.Atoi(text)
			// 	if err != nil || index <= 0 || index > len(tasks[chatID]) {
			// 		send(bot, chatID, "❌ Неверный номер задачи.")
			// 	} else {
			// 		index--
			// 		removed := tasks[chatID][index]
			// 		tasks[chatID] = append(tasks[chatID][:index], tasks[chatID][index+1:]...)
			// 		send(bot, chatID, "🗑 Удалена: "+removed.Text)
			// 		send(bot, chatID, "🤖 Доступные команды: /add /list /delete /done")
			// 	}
			// 	awaiting[chatID] = ""
			// 	continue
			// case "done":
			// 	index, err := strconv.Atoi(text)
			// 	if err != nil || index <= 0 || index > len(tasks[chatID]) {
			// 		send(bot, chatID, "❌ Неверный номер задачи.")
			// 	} else {
			// 		tasks[chatID][index-1].IsDone = true
			// 		send(bot, chatID, "🎉 Задача отмечена как выполненная!")
			// 		send(bot, chatID, "🤖 Доступные команды: /add /list /delete /done")
			// 	}
			// 	awaiting[chatID] = ""
			// 	continue

		}

		switch text {
		case "/add":
			awaiting[chatID] = "add"
			send(bot, chatID, "✏️ Напишите задачу:")
		// case "/delete":
		// 	if len(tasks[chatID]) == 0 {
		// 		send(bot, chatID, "📭 У вас нет задач для удаления.")
		// 	} else {
		// 		awaiting[chatID] = "delete"
		// 		send(bot, chatID, "❓ Введите номер задачи для удаления:")
		// 	}
		case "/list":
			tasks, err := sc.ListTasks(chatID)
			if err != nil {
				log.Fatal("Ошибка при получении задач:", err)
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
		// case "/done":
		// 	if len(tasks[chatID]) == 0 {
		// 		send(bot, chatID, "📭 У вас нет задач.")
		// 	} else {
		// 		awaiting[chatID] = "done"
		// 		send(bot, chatID, "☑️ Введите номер задачи, которую вы завершили:")
		// 	}
		default:
			send(bot, chatID, "🤖 Доступные команды: /add /list /delete /done")
		}

	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

// func justTest(){
// 	m := [...]int{}
// 	fmt.Println(m)

// 	m2 := make([]int, 0, 0)
// 	fmt.Println(m2)

// 	for index, v := range m2 {
// 		fmt.Println(index, v)
// 	}

// 	map1 := make(map[string]int)
// 	map1["task 1"] = 1

// 	ch := make(chan int, 3)
// }
