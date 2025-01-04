package discord

import (
	"fmt"

	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
)

var CONFIGURE_POLICY_ROLE_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "configure-policy-role",
	Description:              "Set roles based on the configured pools",
	DefaultMemberPermissions: &ADMIN,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionRole,
			Name:        "role",
			Description: "The role to associate with the attached policy range",
			Required:    true,
			MaxLength:   255,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "min",
			Description: "The minimum quantity to qualify for the attached role",
			Required:    true,
			MaxLength:   255,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "max",
			Description: "The max quantity to qualify for the attached role",
			Required:    false,
		},
	},
}

var CONFIGURE_POLICY_ROLE_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)

	role := options["role"].RoleValue(s, i.GuildID)
	min := options["min"].IntValue()
	maybemax, ok := options["max"]
	var max int64 = 999_999_999_999
	if ok {
		max = maybemax.IntValue()
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Adding policy role <@&%v>: Min %d - Max %d\n", role.ID, min, max),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Configure Policy ID",
		},
	})

	config := preeb.LoadConfig(i.GuildID)

	order := int64(len(config.PolicyRoles) + 1)

	bounds := preeb.RoleBounds{Min: preeb.Bound(min), Max: preeb.Bound(max), Order: order}

	if bounds.IsValid() {
		config.PolicyRoles[role.ID] = bounds
		config.Save()
		return
	}
}
