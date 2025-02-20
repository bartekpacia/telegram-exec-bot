package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WhitelistedUsers returns a map of allowed Telegram usernames
func loadWhitelistedUsers() map[string]bool {
	// Read from environment variable for simplicity
	// Format: username1,username2,username3
	whitelistStr := os.Getenv("TELEGRAM_WHITELIST")
	whitelist := make(map[string]bool)

	users := strings.Split(whitelistStr, ",")
	for _, user := range users {
		if trimmed := strings.TrimSpace(user); trimmed != "" {
			whitelist[trimmed] = true
		}
	}

	return whitelist
}

func executeCommand(command string) (string, error) {
	expandedCommand := os.ExpandEnv(command)
	parts := strings.Fields(expandedCommand)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing command: %v", err)
	}

	return string(output), nil
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	whitelist := loadWhitelistedUsers()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !whitelist[update.Message.From.UserName] {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "wypierdalaj")
			bot.Send(msg)
			continue
		}

		output, err := executeCommand(update.Message.Text)
		response := output
		if err != nil {
			response = fmt.Sprintf("Error: %v", err)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		bot.Send(msg)
	}
}
