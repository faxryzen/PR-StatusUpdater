package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/faxryzen/pr-updater/internal/labs"
	ft "github.com/faxryzen/pr-updater/internal/fetcher"
)

type Bot struct {
	API   *tgbotapi.BotAPI
	Token string
	Admin map[int64]bool
	Repos []ft.Repository
}

func Init(token string, admins []int64) *Bot {
	repos, err := ft.LoadRepositories()
	if err != nil {
		log.Fatal(err)
	}
	ads := make(map[int64]bool)
	for _, ad := range admins {
		ads[ad] = true
	}

	return &Bot{
		Token: token,
		Admin: ads,
		Repos: repos,
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
			handleCallback(b, update)
			continue
		}
		//ignore any non msg updates
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch text {
		case "Run fetch":
			showRepoButtons(b, chatID)
			continue
		case "Settings":
			if b.Admin[chatID] {
				showSettingsButtons(b, chatID)
				continue
			}
			b.API.Send(tgbotapi.NewMessage(chatID, "У вас нет таких прав доступа"))
			continue
		}

		switch update.Message.Command() {
		case "start":
			sendMainMenu(b.API, chatID)
			log.Println(chatID)
		default:
			b.API.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
		}
	}
}

func sendErr(bot *tgbotapi.BotAPI, chatID int64, err error) {
	log.Println(err)
	bot.Send(tgbotapi.NewMessage(chatID, "Что-то пошло не так.."))
}

func handleCallback(bot *Bot, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	//fetch repo
	if after, ok := strings.CutPrefix(data, "repo:"); ok {

		repoIndex, err := strconv.Atoi(after)
		if err != nil {
			panic("it cant be str in repo: callback")
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			"Считаю баллы",
		)
		bot.API.Send(callback)

		go handleFetch(bot, chatID, repoIndex)
	}
	//upload gist
	if after, ok := strings.CutPrefix(data, "forward:"); ok {

		repoIndex, err := strconv.Atoi(after)
		if err != nil {
			panic("it cant be str in forward: callback")
		}

		err = ft.UploadCSVtoGist(bot.Repos[repoIndex])
		if err != nil {
			sendErr(bot.API, chatID, err)
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			"Загружено",
		)
		bot.API.Send(callback)
	}
	//settings
	if after, ok := strings.CutPrefix(data, "settings:"); ok {
		//setup repos
		if after, ok := strings.CutPrefix(after, "repo:"); ok {
			switch after {
			case "init":
				showSettingsRepoButtons(bot, chatID)
			case "add":
				fmt.Println("adding repo")
			case "del":
				fmt.Println("deleting repo")
			}
			
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "",)
			bot.API.Send(callback)
		}
	}
}

func handleFetch(bot *Bot, chatID int64, repoIndex int) {
	repo := bot.Repos[repoIndex]

	bot.API.Send(tgbotapi.NewMessage(
		chatID,
		"Загружаю PR для " + repo.Auth + "/" + repo.Name,
	))

	if err := os.MkdirAll("output", 0755); err != nil {
		sendErr(bot.API, chatID, err)
		return
	}

	file, err := os.Create("output/" + repo.Name + ".json")
	if err != nil {
		sendErr(bot.API, chatID, err)
		return
	}
	defer file.Close()

	j, err := labs.FetchLabsToJSON(repo)
	if err != nil {
		sendErr(bot.API, chatID, err)
		return
	}

	if _, err := file.Write(j); err != nil {
		sendErr(bot.API, chatID, err)
		return
	}

	file.Close()

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(file.Name()))
	doc.Caption = "Готово: " + repo.Auth + "/" + repo.Name

	var rows [][]tgbotapi.InlineKeyboardButton
	btn := tgbotapi.NewInlineKeyboardButtonData("Загрузить в таблицы", "forward:" + strconv.Itoa(repoIndex))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	doc.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	sentMsg, err := bot.API.Send(doc)
	if err != nil {
		sendErr(bot.API, chatID, err)
		return
	}

	bot.API.Buffer = sentMsg.MessageID
}
