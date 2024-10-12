package discord

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"

	"github.com/bwmarrin/discordgo"
)

var ADMIN int64 = discordgo.PermissionAdministrator

var CONFIGURE_POOL_ID_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "configure-pool-id",
	Description:              "Set a pool ID to work with your server.",
	DefaultMemberPermissions: &ADMIN,
	Options: []*discordgo.ApplicationCommandOption{{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "poolid",
		Description: "The pool ID to link to your discord (i.e. pool19peeq2czwunkwe3s70yuvwpsrqcyndlqnxvt67usz98px57z7fk)",
		Required:    true,
		MaxLength:   255,
	}},
}

var CONFIGURE_POOL_ID_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)
	poolID := options["poolid"].StringValue()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Saving %s as your pool ID", poolID),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Configure Pool ID",
		},
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	m, err := blockfrost.GetPoolMetaData(ctx, poolID)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Something went wrong! Maybe open a #support-ticket: %v", err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	var b bytes.Buffer
	sentence := "POOL: {{ .name }} [{{ .ticker }}]\nDESCRIPTION: {{ .desc }}\nPOOL ID: {{ .id }}\n"
	partial := template.Must(template.New("configure-pool-id-template").Parse(sentence))
	partial.Execute(&b, map[string]interface{}{
		"name":   m.Name,
		"ticker": m.Ticker,
		"desc":   m.Description,
		"id":     m.PoolID,
	})

	content := b.String()

	config := preebot.LoadConfig()
	config.PoolID = poolID
	preebot.SaveConfig(config)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
