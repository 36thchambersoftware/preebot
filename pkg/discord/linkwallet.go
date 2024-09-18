package discord

import (
	"github.com/bwmarrin/discordgo"
)

var LinkWalletCommand = discordgo.ApplicationCommand{
	Version:     "0.01",
	Name:        "link-wallet",
	Description: "Link your wallet to receive delegator roles",
}

var LinkWalletHandler = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
}
