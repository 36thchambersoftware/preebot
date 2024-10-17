package discord

import (
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

	CheckDelegation(s, i)
}
