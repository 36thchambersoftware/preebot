package discord

import (
	"log/slog"
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
)

var CHECK_MY_WALLETS_COMMAND = discordgo.ApplicationCommand{
	Version:     "0.01",
	Name:        "check-my-wallets",
	Description: "Check which pool has your delegation.",
}

var CHECK_MY_WALLETS_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Let's take a look at your current delegation!",
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Check Delegation",
		},
	})

	user := preeb.LoadUser(i.Member.User.ID)
	slog.Info(CHECK_MY_WALLETS_COMMAND.Name, "USER_ID", i.Member.User.ID)

	if len(user.Wallets) == 0 {
		content := "You need to link your wallet first. Please use /link-wallet."
		S.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	newRoleIDs, err := AssignQualifiedRoles(i.GuildID, user)
	if err != nil {
		slog.Error("could not assign roles", "ERROR", err)
		S.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return
	}

	slog.Info(CHECK_MY_WALLETS_COMMAND.Name, "ROLES", newRoleIDs)
	content := FormatNewRolesMessage(user, newRoleIDs)
	_, err = S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	if err != nil {
		slog.Error("could not respond", "ERROR", err)
	}
	// S.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
	// 	Content: content,
	// 	Flags:   discordgo.MessageFlagsEphemeral,
	// })
}