package discord

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preeb"

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
	},{
		Type:        discordgo.ApplicationCommandOptionChannel,
		Name:        "channel",
		Description: "The channel you want block updates to be sent to (i.e. general chat or bot commands)",
		Required:    true,
		MaxLength:   255,
		ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	}},
}

var CONFIGURE_POOL_ID_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)
	channelID := options["channel"].ChannelValue(nil).ID
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
	sentence := "POOL: {{ .name }} [{{ .ticker }}]\nDESCRIPTION: {{ .desc }}\nPOOL ID: {{ .id }}\nCHANNEL: {{ .channel }}\n"
	partial := template.Must(template.New("configure-pool-id-template").Parse(sentence))
	partial.Execute(&b, map[string]interface{}{
		"name":   m.Name,
		"ticker": m.Ticker,
		"desc":   m.Description,
		"id":     m.PoolID,
		"channel": channelID,
	})

	content := b.String()

	config := preeb.LoadConfig(i.GuildID)
	config.PoolChannelID = channelID
	config.PoolIDs[poolID] = true
	config.Save()

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
