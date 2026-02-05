package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/faxryzen/pr-updater/internal/cfgs"
	"github.com/faxryzen/pr-updater/internal/dds"
)

type Bot struct {
	API   *tgbotapi.BotAPI
	Token string
}

func Init() *Bot {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	return &Bot{
		Token: token,
	}
}

func (b *Bot) Run() {
	newBot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		log.Fatal(err)
	}
	b.API = newBot

	log.Printf("Authorized as %s", b.API.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		//inline button callback
		if update.CallbackQuery != nil {
			handleCallback(b.API, update)
			continue
		}
		//if we got msg
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch text {
		case "Update PR":
			showRepoButtons(b.API, chatID)
			continue

			/*
				case "Add Repo":
					bot.Send(tgbotapi.NewMessage(
						chatID,
						"Используй:\n/addrepo owner name",
					))
					continue
			*/
		}

		switch update.Message.Command() {
		case "start":
			sendMainMenu(b.API, chatID)
			/*
				case "addrepo":
					handleAddRepo(bot, update)
			*/
		default:
			b.API.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
		}
	}
}

func sendErr(bot *tgbotapi.BotAPI, chatID int64, err error) {
	log.Println(err)
	bot.Send(tgbotapi.NewMessage(chatID, "Что-то пошло не так.."))
}

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Update PR"),
		),
		/*
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Add Repo"),
			),
		*/
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
		/*
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"Нет репозиториев.\nДобавь:\n/addrepo owner name",
			))
		*/
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

/*
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

	bot.Send(tgbotapi.NewMessage(chatID, "Репозиторий сохранён"))
}
*/

func runGhQuery(query string) ([]dds.PullRequest, error) {
	jqFilter, err := os.ReadFile("configs/pr_filter.jq")
	if err != nil {
		return nil, fmt.Errorf("failed to read jq filter: %w", err)
	}

	cmd := exec.Command(
		"gh", "api", "graphql",
		"-f", "query="+query,
		"--jq", string(jqFilter))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, out)
	}

	if len(out) == 0 {
		return []dds.PullRequest{}, nil
	}

	var pullRequests []dds.PullRequest
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {

		if line == "" {
			continue
		}

		var pr dds.PullRequest

		err = json.Unmarshal([]byte(line), &pr)
		if err != nil {
			panic("what da hell")
		}

		if pr.LabID == "" {
			continue
		}

		pr.Times["created"] = dds.ToMoscow(pr.Times["created"])

		if pr.Times["fined"].IsZero() {
			delete(pr.Times, "fined")
		} else {
			pr.Times["fined"] = dds.ToMoscow(pr.Times["fined"])
		}
		if pr.Times["merged"].IsZero() {
			delete(pr.Times, "merged")
		} else {
			pr.Times["merged"] = dds.ToMoscow(pr.Times["merged"])
		}

		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	if after, ok := strings.CutPrefix(data, "repo:"); ok {
		repoStr := after
		parts := strings.Split(repoStr, "/")

		repo := cfgs.Repo{
			Owner: parts[0],
			Name:  parts[1],
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			"Считаю баллы",
		)
		bot.Send(callback)

		go handleUpdate(bot, chatID, repo)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, chatID int64, repo cfgs.Repo) {
	bot.Send(tgbotapi.NewMessage(
		chatID,
		"Загружаю PR для "+repo.Owner+"/"+repo.Name,
	))

	labs, err := dds.LoadDeadlines()
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	querys := cfgs.GetQuerys([]string{
		repo.Name,
		repo.Owner,
	})

	mergedPRs, err := runGhQuery(querys[0])
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	openedPRs, err := runGhQuery(querys[1])
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	tmpFile, err := os.CreateTemp("", "pr_*.json")
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	for i := range mergedPRs {
		dds.CalculateScore(&mergedPRs[i], labs[mergedPRs[i].LabID])
	}

	j, err := json.MarshalIndent(append(mergedPRs, openedPRs...), "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	tmpFile.Write(j)

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(tmpFile.Name()))
	doc.Caption = "Готово: " + repo.Owner + "/" + repo.Name

	bot.Send(doc)
}
