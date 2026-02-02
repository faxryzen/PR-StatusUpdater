package main

import (
	"github.com/faxryzen/pr-updater/internal/bot"
)

func main() {
	bot := bot.Init()
	bot.Run()
}
