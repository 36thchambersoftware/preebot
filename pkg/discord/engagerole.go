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
		Description: "Receive the twitter-liason role",
	}
)

var ENGAGE_ROLE_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var (
		response       string
		user_has_role  bool
		twitterliaison *discordgo.Role
	)

	perms, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Fatalf("Could not get roles: %v", err)
	}

	for _, role := range perms {
		if role.Name == ENGAGE_ROLE_NAME {
			twitterliaison = role
		}
	}

	if !user_has_role {
		err = s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, twitterliaison.ID)
		if err != nil {
			log.Fatalf("Could not assign role: %v", err)
		}

		response = "Role added: " + twitterliaison.Name
	} else {
		err = s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, twitterliaison.ID)
		if err != nil {
			log.Fatalf("Could not remove role: %v", err)
		}

		response = "Role removed: " + twitterliaison.Name
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
