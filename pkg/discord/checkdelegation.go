package discord

import (
	"context"
	"fmt"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"

	"github.com/bwmarrin/discordgo"
)

var CHECK_DELEGATION_COMMAND = discordgo.ApplicationCommand{
	Version:     "0.01",
	Name:        "check-delegation",
	Description: "Check which pool has your delegation.",
}

var CHECK_DELEGATION_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Let's take a look at your current delegation!",
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Check Delegation",
		},
	})

	user := preebot.LoadUser(i.Member.User.ID)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if user.Wallets == nil {
		content := "You need to link your wallet first. Please use /link-wallet."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}

	for _, address := range user.Wallets {
		accountDetails := blockfrost.GetStakeInfo(ctx, address)
		if *accountDetails.PoolID == blockfrost.PREEB_POOL_ID && accountDetails.Active {
			role, err := AssignDelegatorRole(s, i, accountDetails)
			if err != nil {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong! Maybe open a #support-ticket ",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
			} else if role != nil {
				content := "You have been assigned a role! <@&" + role.ID + ">"
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
			} else {
				content := fmt.Sprintf("%+v", role)
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
			}
		} else {
			content := "Oh no! It looks like you're not delegating to PREEB!"
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
		}
	}
}
