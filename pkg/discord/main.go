package discord

import (
	"context"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
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

	RefreshCommands()

	ctx := context.Background()
	go automaticRoleChecker(ctx)
	go automaticPoolBlocks(ctx)
}

func RefreshCommands() {
	appId, ok := os.LookupEnv("PREEBOT_APPLICATION_ID")
	if !ok {
		log.Fatalf("Missing application id")
	}
	registeredCommands, err := S.ApplicationCommands(appId, "")
	if err != nil {
		log.Panicf("Cannot retrieve commands:\n%v", err)
	}

	guildID := ""
	_, err = S.ApplicationCommandBulkOverwrite(appId, guildID, registeredCommands)
	if err != nil {
		log.Panicf("Cannot overwrite commands:\n%v", err)
	}
}

func automaticRoleChecker(ctx context.Context) {
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
        select {
        case <-time.After(time.Duration(interval) * time.Minute):
            slog.Info("Checking roles...")
            AutomaticRoleChecker()
        case <-ctx.Done():
			RefreshCommands()
            return
        }
    }
}

func automaticPoolBlocks(ctx context.Context) {
	var lastBlock string
	for {
		select {
		case <-time.After(time.Minute):
			slog.Info("Getting Pool Info")
			lastBlock = AutomaticPoolBlocks(ctx, lastBlock)
		case <-ctx.Done():
			RefreshCommands()
			return
		}
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