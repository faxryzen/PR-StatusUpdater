package main

import (
	"os"
	"log"

	"github.com/joho/godotenv"
	"github.com/faxryzen/pr-updater/internal/bot"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	bot := bot.Init(token, -1003870316764)
	bot.Run()
}
