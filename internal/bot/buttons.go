package bot

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Run fetch"),
			tgbotapi.NewKeyboardButton("Settings"),
		),
	)
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, "Выбери действие:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showRepoButtons(b *Bot, chatID int64) {
	if len(b.Repos) == 0 {
		b.API.Send(tgbotapi.NewMessage(
			chatID,
			"Похоже, что репозиториев нет..",
		))
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

var settingsButtons = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Repositories", "settings:repo:init",),
		tgbotapi.NewInlineKeyboardButtonData("Output CSV Format", "settings:form:init",),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Labs config", "settings:labs:init",),
		tgbotapi.NewInlineKeyboardButtonData("Output CSV Format", "settings:form",),
	),
)

func showSettingsButtons(b *Bot, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Доступные настройки:")
	msg.ReplyMarkup = settingsButtons

	b.API.Send(msg)
}

var settingsRepoButtons = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Add", "settings:repo:add",),
		tgbotapi.NewInlineKeyboardButtonData("Remove", "settings:repo:del",),
	),
)

func showSettingsRepoButtons(b *Bot, chatID int64) {
	var text string

	if len(b.Repos) == 0 {
		text = "Пока что здесь пусто..."
	} else {
		text = "Сохраненные репозитории:<code>"
		for i, r := range b.Repos {
			text += "\n[" + strconv.Itoa(i + 1) + "] " + r.Auth + "/" + r.Name
		}
		text += "</code>\n"
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = settingsRepoButtons

	b.API.Send(msg)
}
