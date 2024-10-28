package discord

import (
	"flag"
	"log"
	"log/slog"
	"net/url"
	"os"
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
)

func init() { flag.Parse() }

func init() {
	token, ok := os.LookupEnv("PREEBOT_TOKEN")
	if !ok {
		log.Fatalf("Missing token")
	}
	var err error
	S, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	go func() {
		for {
			slog.Info("Checking roles...")
			AutomaticRoleChecker()
			time.Sleep(24 * time.Hour)
		}
	}()
}

func init() {
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
