package discord

import (
	"github.com/bwmarrin/discordgo"
)

var CHECK_ASSETS_COMMAND = discordgo.ApplicationCommand{
	Version:     "0.01",
	Name:        "check-assets",
	Description: "Get the custodian wallet for your Tireless Workers",
}

var CHECK_ASSETS_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Link the custodian wallet for your Tireless Workers",
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Link Custodian Wallet",
		},
	})

	CheckAssets(s, i)

	// custodianStakeAddrs := []string{}
	// for _, address := range user.Wallets {
	// 	custodianStakeAddr, err := mongodb.GetCustodianByNecro(address)
	// 	if err != nil {
	// 		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
	// 			Content: fmt.Sprintf("Something went wrong! Maybe open a #support-ticket: %v", err),
	// 			Flags:   discordgo.MessageFlagsEphemeral,
	// 		})
	// 	}

	// 	custodianStakeAddrs = append(custodianStakeAddrs, custodianStakeAddr)
	// }
}
