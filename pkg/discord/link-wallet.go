package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/rand"
)

var (
	LINK_WALLET_AMOUNT         = "3141590"
	LINK_WALLET_AMOUNT_DISPLAY = "3.14159"
	LINK_WALLET_COMMAND        = discordgo.ApplicationCommand{
		Version:     "0.01",
		Name:        "link-wallet",
		Description: "Link your wallet to receive delegator roles.",
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
	options := GetOptions(i)
	address := options["address"].StringValue()
	slog.Info(LINK_WALLET_COMMAND.Name, "USER_ID", i.Member.User.ID, "ADDRESS", address)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	address, err := blockfrost.HandleAddress(ctx, address)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Uh oh! %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
				Title:   "Wallet Linker",
			},
		})

		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Wallet to link: " + address + "\n\nSend the following ada to yourself. Go ahead, I'll wait a couple minutes. :D",
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Wallet Linker",
		},
	})

	user := preeb.LoadUser(i.Member.User.ID)
	for _, addr := range user.Wallets {
		if address == addr.String() {
			content := "Your wallet has already been linked! But feel free to link another."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}
	}

	amount := rand.Intn(1000000)
	LINK_WALLET_AMOUNT = fmt.Sprintf("1%s",strconv.Itoa(amount))
	LINK_WALLET_AMOUNT_DISPLAY = fmt.Sprintf("1.%s", strconv.Itoa(amount)) // strconv.FormatFloat(1.0 + (float64(amount) / float64(1000000)), 'f', -1, 64)

	linkAmtMsg, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: LINK_WALLET_AMOUNT_DISPLAY,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	// Give the user a moment to send the tx before checking for it.
	time.Sleep(120 * time.Second)

	content := "I'll check to see if your transaction is on the blockchain now."
	msg, err := s.FollowupMessageEdit(i.Interaction, linkAmtMsg.ID, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	txDetails, err := blockfrost.GetLastTransaction(ctx, address)
	if err != nil {
		content := fmt.Sprintf("Something went wrong! Maybe open a #support-ticket: %v", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}
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
		account := blockfrost.GetAccountByAddress(ctx, address)
		user.Wallets[preeb.StakeAddress(account.StakeAddress)] = preeb.Address(address)

		if user.ID == "" {
			user.ID = i.Member.User.ID
		}

		// preeb.SaveUser(user)
		user.Save()
	} else {
		content := "I couldn't verify your address. Maybe the transaction isn't on the blockchain yet. Try the /link-wallet command again when your transaction is complete. If it still doesn't work, open a ticket and we'll figure it out."
		s.FollowupMessageEdit(i.Interaction, msg.ID, &discordgo.WebhookEdit{
			Content: &content,
		})
	}

	config := preeb.LoadConfig(i.GuildID)

	if config.PoolIDs != nil || config.PolicyIDs != nil {
		newRoleIDs, err := AssignQualifiedRoles(i.GuildID, user)
		if err != nil {
			S.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong! Maybe open a #support-ticket ",
				Flags:   discordgo.MessageFlagsEphemeral,
			})

			return
		}

		content := FormatNewRolesMessage(user, newRoleIDs)
		S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}
}
