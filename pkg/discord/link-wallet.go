package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/logger"
	"preebot/pkg/preeb"

	bfg "github.com/blockfrost/blockfrost-go"
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

	rand.Seed(uint64(time.Now().UnixNano()))
	amount := rand.Intn(1000000)
	LINK_WALLET_AMOUNT = fmt.Sprintf("1%s",strconv.Itoa(amount))
	LINK_WALLET_AMOUNT_DISPLAY = fmt.Sprintf("1.%s", strconv.Itoa(amount)) // strconv.FormatFloat(1.0 + (float64(amount) / float64(1000000)), 'f', -1, 64)
	logger.Record.Info("Linking wallet", "AMOUNT", LINK_WALLET_AMOUNT_DISPLAY, "LINK_WALLET_AMOUNT", LINK_WALLET_AMOUNT)
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

	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	foundTx := false
	consecutiveErrors := 0
	const maxErrors = 4
	
	loop:
	for {
		var txDetails bfg.TransactionUTXOs
		var txErr error
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			txDetails, txErr = blockfrost.GetLastTransaction(ctx, address)
			if txErr != nil {
				consecutiveErrors++
				logger.Record.Warn("Error fetching last transaction", "error", txErr, "attempt", consecutiveErrors)

				if consecutiveErrors == 1 {
					msg := "Still checking... encountered a small hiccup reaching the blockchain. Trying again!"
					s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: msg,
						Flags:   discordgo.MessageFlagsEphemeral,
					})
				}

				if consecutiveErrors >= maxErrors {
					content := "There was a problem checking your transaction after multiple tries. Please try `/link-wallet` again in a few minutes or open a #support-ticket."
					s.FollowupMessageEdit(i.Interaction, linkAmtMsg.ID, &discordgo.WebhookEdit{
						Content: &content,
					})
					break loop
				}

				continue
			}

			// Reset error count after a successful call
			consecutiveErrors = 0

			for _, output := range txDetails.Outputs {
				for _, amount := range output.Amount {
					logger.Record.Info("Checking amount", "UNIT", amount.Unit, "QUANTITY", amount.Quantity, "LINK_WALLET_AMOUNT", LINK_WALLET_AMOUNT)
					if amount.Unit == "lovelace" && amount.Quantity[:4] == LINK_WALLET_AMOUNT[:4] {
						logger.Record.Info("Found transaction with correct amount", "ADDRESS", address, "AMOUNT", amount.Quantity)
						foundTx = true
						break
					}
				}
				if foundTx {
					break loop
				}
			}

			if foundTx {
				logger.Record.Info("Transaction matched, exiting loop")
				break loop
			}
		}
	}


	if !foundTx {
		content := "I couldn't verify your address within 5 minutes. Make sure you sent the right amount and try again. If you're stuck, open a #support-ticket."
		s.FollowupMessageEdit(i.Interaction, linkAmtMsg.ID, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	// Link successful
	content := "Great! Your wallet has been linked!"
	s.FollowupMessageEdit(i.Interaction, linkAmtMsg.ID, &discordgo.WebhookEdit{
		Content: &content,
	})


	account := blockfrost.GetAccountByAddress(ctx, address)
	// if user.Wallets == nil {
	// 	user.Wallets = make(preeb.Wallets)
	// }
	user.Wallets[preeb.StakeAddress(account.StakeAddress)] = preeb.Address(address)

	if user.ID == "" {
		user.ID = i.Member.User.ID
	}

	// preeb.SaveUser(user)
	user.Save()

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
