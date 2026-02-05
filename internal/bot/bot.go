package bot

import (
	"errors"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

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

var ErrEmptyRepos = errors.New("repos.csv is empty")

func showRepoButtons(bot *tgbotapi.BotAPI, chatID int64) {
	repos, err := dds.GetRepositories()
	if err != nil {
		sendErr(bot, chatID, err)
		return
	}

	if len(repos) == 0 {
		sendErr(bot, chatID, ErrEmptyRepos)
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
			r.Auth + "/" + r.Name,
			"repo:" + r.Auth + "/" + r.Name,
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

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	if after, ok := strings.CutPrefix(data, "repo:"); ok {
		repoStr := after
		parts := strings.Split(repoStr, "/")

		repo := dds.Repository{
			Auth: parts[0],
			Name: parts[1],
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			"Считаю баллы",
		)
		bot.Send(callback)

		go handleUpdate(bot, chatID, repo)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, chatID int64, repo dds.Repository) {
	bot.Send(tgbotapi.NewMessage(
		chatID,
		"Загружаю PR для " + repo.Auth + "/" + repo.Name,
	))

	tmpFile, err := os.CreateTemp("", "pr_*.json")
	if err != nil {
		sendErr(bot, chatID, err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	j, err := dds.UnloadLabs(repo)
	if err != nil {
		sendErr(bot, chatID, err)
	}

	tmpFile.Write(j)

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(tmpFile.Name()))
	doc.Caption = "Готово: " + repo.Auth + "/" + repo.Name

	bot.Send(doc)
}
