package discord

import (
	"log"

	"preebot/pkg/preeb"

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
	var engageRole *discordgo.Role
	config := preeb.LoadConfig(i.GuildID)

	engageRole, err := FindRoleByRoleID(i.GuildID, config.EngageRole)
	if err != nil {
		log.Fatalf("Could not find role: %v", err)
	}

	err = ToggleRole(i, engageRole)
	if err != nil {
		log.Fatalf("Could not toggle role: %v", err)
	}
}
