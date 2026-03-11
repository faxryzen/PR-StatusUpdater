package main

import (
	"log"
	"os"
	"strings"
	"strconv"

	"github.com/faxryzen/pr-updater/internal/bot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	whitelist := os.Getenv("SETTINGS_WHITELIST")

	adminsStr := strings.Split(whitelist, ",")
	admins := []int64{}
	for i := range adminsStr {
		admin, err := strconv.ParseInt(adminsStr[i], 10, 64)
		if err != nil {
			log.Fatal("Error parse admins from .env file")
		}
		admins = append(admins, admin)
	}
	

	bot := bot.Init(token, admins)
	bot.Run()
}
