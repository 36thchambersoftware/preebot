package discord

import (
	"context"
	"fmt"
	"preebot/pkg/blockfrost"
	"preebot/pkg/preeb"
	"time"

	"github.com/bwmarrin/discordgo"
)

var WHITELIST_ADDRESS_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "whitelist-address",
	Description:              "Add a wallet to the whitelist",
	Options: []*discordgo.ApplicationCommandOption{{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "address",
		Description: "Your wallet address",
		Required:    true,
		MaxLength:   255,
	}},
}

var WHITELIST_ADDRESS_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)
	address := options["address"].StringValue()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Adding to whitelist...\n%s", address),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Add wallet to WL",
		},
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	address, err := blockfrost.HandleAddress(ctx, address)
	if err != nil {
		content := fmt.Sprintf("I couldn't resolve your handle!\n%s",address)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	addr := blockfrost.GetAddress(ctx, address)

	whitelist := preeb.Whitelist{
		UserID:      	i.Member.User.ID,
		StakeAddress: 	preeb.StakeAddress(*addr.StakeAddress),
		Address:     	preeb.Address(address),
		GuildID:     	i.GuildID,
	}

	err = whitelist.AddAddress()
	if err != nil {
		content := fmt.Sprintf("Wallet not found!\n%s",address)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	content := fmt.Sprintf("You're in friend! Your wallet is whitelisted!\n%s",address)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
