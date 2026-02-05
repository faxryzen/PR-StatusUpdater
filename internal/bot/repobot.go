package bot

import (
	"errors"
	"os"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/faxryzen/pr-updater/internal/dds"
)

type Bot struct {
	API   *tgbotapi.BotAPI
	Token string
	Chann int64
	Repos []dds.Repository
}

func Init(token string, channel int64) *Bot {
	repos, err := dds.GetRepositories()
	if err != nil {
		log.Fatal(err)
	}

	return &Bot{
		Token: token,
		Chann: channel,
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
		//if we got msg
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch text {
		case "Update PR":
			showRepoButtons(b, chatID)
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

func showRepoButtons(b *Bot, chatID int64) {
	if len(b.Repos) == 0 {
		sendErr(b.API, chatID, ErrEmptyRepos)
		/*
			bot.Send(tgbotapi.NewMessage(
				chatID,
				"Нет репозиториев.\nДобавь:\n/addrepo owner name",
			))
		*/
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for i, r := range b.Repos {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			r.Auth + "/" + r.Name,
			"repo:" + strconv.Itoa(i),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, "Выбери репозиторий:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	b.API.Send(msg)
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

func handleCallback(bot *Bot, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

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

		go handleUpdate(bot, chatID, repoIndex)
	}
	if after, ok := strings.CutPrefix(data, "forward:"); ok {

		repoIndex, err := strconv.Atoi(after)
		if err != nil {
			panic("it cant be str in forward: callback")
		}

		err = dds.UploadGist(bot.Repos[repoIndex])
		if err != nil {
			sendErr(bot.API, chatID, err)
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			"Загружено",
		)
		bot.API.Send(callback)
	}
}

func handleUpdate(bot *Bot, chatID int64, repoIndex int) {
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

	j, err := dds.UnloadLabs(repo)
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
