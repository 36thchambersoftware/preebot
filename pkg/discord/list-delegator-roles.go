package discord

import (
	"preebot/pkg/preebot"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var LIST_DELEGATOR_ROLES_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "list-delegator-roles",
	Description:              "See the configured delegator roles",
	DefaultMemberPermissions: &ADMIN,
}

var LIST_DELEGATOR_ROLES_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := preebot.LoadConfig(i.GuildID)

	sentence := "## Delegator Roles\n"
	for role, bounds := range config.DelegatorRoles {
		p := message.NewPrinter(language.English)
		sentence = sentence + p.Sprintf(" <@&%s>\t %v - %v \n", role, bounds.Min, bounds.Max)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: sentence,
			Title:   "List Delegator Roles",
		},
	})
}
