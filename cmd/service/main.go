package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/faxryzen/pr-updater/internal/cfgs"
)

const fileFormat = ".csv"

func handleUpdate(bot *tgbotapi.BotAPI, chatID int64, repo cfgs.Repo) {
	msg := tgbotapi.NewMessage(chatID,
		"⏳ Собираю Pull Requests для "+repo.Owner+"/"+repo.Name+"...")
	bot.Send(msg)

	tmpDir, err := os.MkdirTemp("", "prupdater-")
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "prlist_*.csv")
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}
	defer tmpFile.Close()

	querys := cfgs.GetQuerys([]string{
		repo.Name,
		repo.Owner,
	})

	prMerged, err := runGhQuery(querys[0])
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	prOpened, err := runGhQuery(querys[1])
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	allData := cfgs.TransformPenaltyData(
		prMerged+prOpened,
		2,
		3,
	)

	if _, err := tmpFile.WriteString(allData); err != nil {
		sendErr(bot, chatID, err)
		return
	}

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(tmpFile.Name()))
	doc.Caption = "✅ Готово! CSV для " + repo.Owner + "/" + repo.Name

	if _, err := bot.Send(doc); err != nil {
		log.Println("telegram send error:", err)
	}
}


func runGhQuery(query string) (string, error) {
	cmd := exec.Command(
		"gh", "api", "graphql",
		"-f", fmt.Sprintf("query=%s", query),
		"--jq", `.data.repository.pullRequests.nodes[] | [
			.number,
			.author.login,
			(.title | split("/") | .[1]),
			.createdAt,
			(if any(.timelineItems.nodes[]; .label.name == "fine")
			then last(.timelineItems.nodes[] | select(.label.name == "fine").createdAt)
			else "null" end)
		] | @csv`,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}

	res := strings.ReplaceAll(string(out), "\"", "")
	res = strings.ReplaceAll(res, ",", ";")

	return res, nil
}

func sendErr(bot *tgbotapi.BotAPI, chatID int64, err error) {
	log.Println(err)
	msg := tgbotapi.NewMessage(chatID, "❌ Ошибка:\n"+err.Error())
	bot.Send(msg)
}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	if strings.HasPrefix(data, "repo:") {
		repoStr := strings.TrimPrefix(data, "repo:")
		parts := strings.Split(repoStr, "/")

		repo := cfgs.Repo{
			Owner: parts[0],
			Name:  parts[1],
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			"⏳ Обновляю PR",
		)
		bot.Send(callback)

		/*
		bot.Send(tgbotapi.NewMessage(
			chatID,
			"Запускаю обновление для "+repo.Owner+"/"+repo.Name,
		))
		*/
		go handleUpdate(bot, chatID, repo)
	}
}

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔄 Update PR"),
			tgbotapi.NewKeyboardButton("➕ Add Repo"),
		),
	)
	kb.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, "Выбери действие:")
	msg.ReplyMarkup = kb

	bot.Send(msg)
}

func showRepoButtons(bot *tgbotapi.BotAPI, chatID int64) {
	repos, err := cfgs.GetRepos()
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	if len(repos) == 0 {
		bot.Send(tgbotapi.NewMessage(
			chatID,
			"Нет сохранённых репозиториев.\nДобавь:\n/addrepo owner name",
		))
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, r := range repos {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			r.Owner+"/"+r.Name,
			"repo:"+r.Owner+"/"+r.Name,
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, "Выбери репозиторий:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	bot.Send(msg)
}


func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized as %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		// 1️⃣ CALLBACK-КНОПКИ
		if update.CallbackQuery != nil {
			handleCallback(bot, update)
			continue
		}

		// 2️⃣ СООБЩЕНИЯ
		if update.Message == nil {
			continue
		}

		text := update.Message.Text

		switch text {
		case "🔄 Update PR":
			showRepoButtons(bot, update.Message.Chat.ID)
			continue

		case "➕ Add Repo":
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Добавить репозиторий:\n/addrepo owner name",
			))
			continue
		}

		chatID := update.Message.Chat.ID
		switch update.Message.Command() {
		case "start":
			sendMainMenu(bot, chatID)
			
		case "update":
			repos, err := cfgs.GetRepos()
			if err != nil {
				sendErr(bot, chatID, err)
				return
			}

			if len(repos) == 0 {
				bot.Send(tgbotapi.NewMessage(chatID,
					"Нет сохранённых репозиториев.\nИспользуй:\n/addrepo owner name"))
				return
			}

			var rows [][]tgbotapi.InlineKeyboardButton

			for _, r := range repos {
				data := fmt.Sprintf("repo:%s/%s", r.Owner, r.Name)
				btn := tgbotapi.NewInlineKeyboardButtonData(
					r.Owner+"/"+r.Name,
					data,
				)
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
			}

			msg := tgbotapi.NewMessage(chatID, "Выбери репозиторий:")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

			bot.Send(msg)
		case "addrepo":
			args := strings.Fields(update.Message.CommandArguments())
			if len(args) != 2 {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"Используй:\n/addrepo owner name",
				))
				return
			}

			repo := cfgs.Repo{
				Owner: args[0],
				Name:  args[1],
			}

			if err := cfgs.SaveRepo(repo); err != nil {
				sendErr(bot, chatID, err)
				return
			}

			bot.Send(tgbotapi.NewMessage(chatID, "✅ Репозиторий сохранён"))

		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Неизвестная команда")
			bot.Send(msg)
		}
	}
}
