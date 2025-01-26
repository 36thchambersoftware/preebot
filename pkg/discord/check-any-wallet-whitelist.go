package discord

import (
	"context"
	"fmt"
	"preebot/pkg/blockfrost"
	"preebot/pkg/preeb"
	"time"

	"github.com/bwmarrin/discordgo"
)

var CHECK_ANY_WALLET_WHITELIST_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "check-whitelist",
	Description:              "Check if the supplied wallet is white-listed ",
	Options: []*discordgo.ApplicationCommandOption{{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "address",
		Description: "Your wallet address",
		Required:    true,
		MaxLength:   255,
	}},
}

var CHECK_ANY_WALLET_WHITELIST_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)
	address := options["address"].StringValue()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Checking address against whitelist\n%s", address),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Check wallet for WL",
		},
	})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	account := blockfrost.GetAccountByAddress(ctx, address)
	content := ""

	err := preeb.CheckAddress(account.StakeAddress)
	if err != nil {
		content = fmt.Sprintf("Wallet not found!\n%s",address)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	content = fmt.Sprintf("You're in friend! Your wallet is whitelisted!\n%s",address)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
