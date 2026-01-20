package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/faxryzen/pr-updater/internal/cfgs"
)

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
			"⏳ Считаю баллы",
		)
		bot.Send(callback)

		go handleUpdate(bot, chatID, repo)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, chatID int64, repo cfgs.Repo) {
	bot.Send(tgbotapi.NewMessage(
		chatID,
		"⏳ Загружаю PR для "+repo.Owner+"/"+repo.Name,
	))

	_, labs, err := cfgs.LoadDeadlines()
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	cfgs.NormalizeDeadlines(labs)

	querys := cfgs.GetQuerys([]string{
		repo.Name,
		repo.Owner,
	})

	mergedRaw, err := runGhQuery(querys[0])
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	lines := strings.Split(mergedRaw, "\n")

	tmpFile, err := os.CreateTemp("", "pr_*.csv")
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	tmpFile.WriteString("PR;User;Lab;Score\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ";")
		if len(parts) < 5 {
			continue
		}

		title := parts[2]
		lab, ok := cfgs.MatchLab(title, labs)
		if !ok {
			continue
		}

		mergedAtUTC, err := time.Parse(time.RFC3339, parts[4])
		if err != nil {
			continue
		}

		mergedAt := cfgs.ToMoscow(mergedAtUTC)

		score := cfgs.CalculateScore(lab, mergedAt)

		row := fmt.Sprintf(
			"%s;%s;%s;%d\n",
			parts[0], // PR number
			parts[1], // user
			lab.ID,
			score,
		)

		tmpFile.WriteString(row)
	}

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(tmpFile.Name()))
	doc.Caption = "✅ Готово: " + repo.Owner + "/" + repo.Name

	bot.Send(doc)
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized as %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.CallbackQuery != nil {
			handleCallback(bot, update)
			continue
		}

		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch text {
		case "Update PR":
			showRepoButtons(bot, chatID)
			continue

		case "Add Repo":
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"Используй:\n/addrepo owner name",
			))
			continue
		}

		switch update.Message.Command() {
		case "start":
			sendMainMenu(bot, chatID)

		case "addrepo":
			handleAddRepo(bot, update)

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
		}
	}
}

func sendErr(bot *tgbotapi.BotAPI, chatID int64, err error) {
	log.Println(err)
	bot.Send(tgbotapi.NewMessage(chatID, "❌ "+err.Error()))
}

func runGhQuery(query string) (string, error) {
	cmd := exec.Command(
		"gh", "api", "graphql",
		"-f", "query="+query,
		"--jq", `.data.repository.pullRequests.nodes[] | [
			.number,
			.author.login,
			.title,
			.createdAt,
			.mergedAt
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

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔄 Update PR"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Add Repo"),
		),
	)
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, "Выбери действие:")
	msg.ReplyMarkup = keyboard
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
			"Нет репозиториев.\nДобавь:\n/addrepo owner name",
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

func handleAddRepo(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
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
}
