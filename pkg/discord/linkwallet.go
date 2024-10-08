package discord

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"

	"github.com/bwmarrin/discordgo"
)

var (
	LINK_WALLET_AMOUNT  = "3141590"
	LINK_WALLET_COMMAND = discordgo.ApplicationCommand{
		Version:     "0.01",
		Name:        "link-wallet",
		Description: "Link your wallet to receive delegator roles",
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "address",
			Description: "The address you want to link to your discord",
			Required:    true,
			MaxLength:   255,
		}},
	}
)

var LINK_WALLET_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	floatAmount, _ := strconv.ParseFloat(LINK_WALLET_AMOUNT, 64)
	LINK_WALLET_AMOUNT_DISPLAY := fmt.Sprintf("%.5f", floatAmount/blockfrost.LOVELACE)
	options := GetOptions(i)
	address := options["address"].StringValue()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Wallet to link: " + address + "\n\nSend " + string(LINK_WALLET_AMOUNT_DISPLAY) + " ada to yourself. Go ahead, I'll wait a couple minutes. :D",
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Wallet Linker",
		},
	})

	user := preebot.LoadUser(i.Member.User.ID)
	for _, wallet := range user.Wallets {
		if address == wallet {
			content := "Your wallet has already been linked! But feel free to link another."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Give the user a moment to send the tx before checking for it.
	time.Sleep(120 * time.Second)

	msg, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "I'll check to see if your transaction is on the blockchain now.",
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	txDetails := blockfrost.GetLastTransaction(ctx, address)
	walletLinked := false

	for _, output := range txDetails.Outputs {
		for _, amount := range output.Amount {
			if amount.Unit == "lovelace" && amount.Quantity == LINK_WALLET_AMOUNT {
				// Link successful
				walletLinked = true
				content := "Great! Your wallet has been linked!"
				s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
					Content: &content,
				})
				break
			}
		}
	}

	if walletLinked {
		accountDetails := blockfrost.GetStakeInfo(ctx, address)
		role, err := AssignDelegatorRole(s, i, accountDetails)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong! Try again, or maybe open a #support-ticket ",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
		}

		content := "You have been assigned a role! <@&" + role.ID + ">"
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong! Try again, or maybe open a #support-ticket ",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
		}

		user.Wallets = append(user.Wallets, address)

		if user.ID == "" {
			user.ID = i.Member.User.ID
		}

		if user.DisplayName == "" {
			user.DisplayName = i.Member.User.GlobalName
		}

		preebot.SaveUser(user)
	}
}
