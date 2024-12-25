package discord

import (
	"fmt"
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
)

var CONFIGURE_ENGAGE_ROLE_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "configure-engage-role",
	Description:              "Set which role is associated with engage",
	DefaultMemberPermissions: &ADMIN,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionRole,
			Name:        "role",
			Description: "The role to associate with engage",
			Required:    true,
			MaxLength:   255,
		},
	},
}

var CONFIGURE_ENGAGE_ROLE_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := preeb.LoadConfig(i.GuildID)
	options := GetOptions(i)
	role := options["role"].RoleValue(s, i.GuildID)

	config.EngageRole = role.ID

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Saving <@&%s> as your Engage Role", role.ID),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Configure Policy ID",
		},
	})

	config.Save()
}
