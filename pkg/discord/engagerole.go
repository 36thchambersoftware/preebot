package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	ENGAGE_ROLE_NAME    = "Twitter Liaison"
	ENGAGE_ROLE_COMMAND = discordgo.ApplicationCommand{
		Version:     "0.01",
		Name:        "tweet-me-baby-one-more-time",
		Description: "Add/Remove the twitter-liaison role",
	}
)

var ENGAGE_ROLE_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var (
		response       string
		twitterLiaison *discordgo.Role
	)

	twitterLiaison, err := FindRoleByName(s, i, ENGAGE_ROLE_NAME)
	if err != nil {
		log.Fatalf("Could not find role: %v", err)
	}

	err = ToggleRole(s, i, twitterLiaison)
	if err != nil {
		log.Fatalf("Could not toggle role: %v", err)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
