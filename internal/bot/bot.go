package bot

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

func runGhQuery(query string) (string, error) {
	cmd := exec.Command(
		"gh", "api", "graphql",
		"-f", "query="+query,
		"--jq", `.data.repository.pullRequests.nodes[] | [
			.number,
			.author.login,
			(.title | split("/") | .[1]),
			.createdAt,
			(if any(.timelineItems.nodes[]; .label.name == "fine") 
			then last(.timelineItems.nodes[] | select(.label.name == "fine").createdAt)
			else (.mergedAt // "null") end)
			] | @csv`)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}

	fmt.Println("===================out:")
	fmt.Println(out)

	res := strings.ReplaceAll(string(out), "\"", "")
	res = strings.ReplaceAll(res, ",", ";")

	fmt.Println("===================res:")
	fmt.Println(res)

	return res, nil
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
		lab, ok := dds.MatchLab(title, labs)
		if !ok {
			continue
		}

		mergedAtUTC, err := time.Parse(time.RFC3339, parts[4])
		if err != nil {
			continue
		}

		mergedAt := dds.ToMoscow(mergedAtUTC)

		score := dds.CalculateScore(lab, mergedAt)

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
	doc.Caption = "Готово: " + repo.Owner + "/" + repo.Name

	bot.Send(doc)
}
