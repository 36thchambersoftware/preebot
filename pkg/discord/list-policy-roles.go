package discord

import (
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var LIST_POLICY_ROLES_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "list-policy-roles",
	Description:              "See the configured policy roles",
	DefaultMemberPermissions: &ADMIN,
}

var LIST_POLICY_ROLES_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := preeb.LoadConfig(i.GuildID)

	sentence := "## Policy Roles\n"
	for role, bounds := range config.PolicyRoles {
		p := message.NewPrinter(language.English)
		sentence = sentence + p.Sprintf(" <@&%s>\t %v - %v \n", role, bounds.Min, bounds.Max)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: sentence,
			Title:   "List Policy Roles",
		},
	})
}
