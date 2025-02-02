package discord

import (
	"flag"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
)

var (
	S                   *discordgo.Session
	DISCORD_WEBHOOK_URL string
	CUSTODIAN_ADDRESSES map[string]string
)

func init() {
	initDiscord()
	initWebhook()
}

func initDiscord() {
	token, ok := os.LookupEnv("PREEBOT_TOKEN")
	if !ok {
		log.Fatalf("Missing token")
	}
	var err error
	S, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	go automaticRoleChecker()
}

func automaticRoleChecker() {
	PREEBOT_ROLE_CHECK_INTERVAL, ok := os.LookupEnv("PREEBOT_ROLE_CHECK_INTERVAL")
	if !ok {
		slog.Error("Interval not set. Roles will not be updated.", "PREEBOT_ROLE_CHECK_INTERVAL", PREEBOT_ROLE_CHECK_INTERVAL)
		return
	}

	interval, err := strconv.Atoi(PREEBOT_ROLE_CHECK_INTERVAL)
	if err != nil {
		slog.Error("Could not read interval. Roles will not be updated.", "PREEBOT_ROLE_CHECK_INTERVAL", PREEBOT_ROLE_CHECK_INTERVAL)
		return
	}

	for {
		slog.Info("Checking roles...")
		AutomaticRoleChecker()
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func initWebhook() {
	// DISCORD_WEBHOOK_URL
	webhook, ok := os.LookupEnv("DISCORD_WEBHOOK_URL")
	if !ok {
		log.Fatalf("Could not get DISCORD_WEBHOOK_URL")
	}

	webhookURL, err := url.Parse(webhook)
	if err != nil {
		log.Fatalf("Invalid webhook url %v", err)
	}

	DISCORD_WEBHOOK_URL = webhookURL.String()
}