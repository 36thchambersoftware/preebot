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

var CONFIGURE_POLICY_ID_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "configure-policy-id",
	Description:              "Set a policy ID to work with your server.",
	DefaultMemberPermissions: &ADMIN,
	Options: []*discordgo.ApplicationCommandOption{{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "policyid",
		Description: "The a policy ID to link to your discord",
		Required:    true,
		MaxLength:   255,
	}},
}

var CONFIGURE_POLICY_ID_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)
	policyID := options["policyid"].StringValue()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Saving %s as your policy ID", policyID),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Configure Policy ID",
		},
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	assets, err := blockfrost.GetPolicyAssets(ctx, policyID)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Something went wrong! Maybe open a #support-ticket: %v", err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	if assets == nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "That looks like a bad policy ID. It has no assets!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	var b bytes.Buffer
	sentence := "POLICY: {{ .id }}"
	partial := template.Must(template.New("configure-policy-id-template").Parse(sentence))
	partial.Execute(&b, map[string]interface{}{
		"id": policyID,
	})

	content := b.String()

	config := preebot.LoadConfig(i.GuildID)
	config.PolicyIDs[policyID] = true
	preebot.SaveConfig(config)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}