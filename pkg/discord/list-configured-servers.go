package discord

import (
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var LIST_CONFIGURED_SERVERS_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "list-configured-servers",
	Description:              "See the configured servers",
	DefaultMemberPermissions: &ADMIN,
}

var LIST_CONFIGURED_SERVERS_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	configs := preeb.LoadConfigs()
	p := message.NewPrinter(language.English)

	sentence := "## Configured Servers\n"
	for _, config := range configs {
		guild, err := s.Guild(config.GuildID)
		if err != nil {
			sentence = sentence + p.Sprintf("Error %s: %v\n", config.GuildID, err)
			continue
		}

		sentence = sentence + p.Sprintf("**%s (%s)**\n", guild.Name, config.GuildID)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: sentence,
			Title:   "List Configured Servers",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
