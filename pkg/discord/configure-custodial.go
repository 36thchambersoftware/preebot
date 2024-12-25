package discord

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
)

var CONFIGURE_CUSTODIAL_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "configure-custodial",
	Description:              "Add an API endpoint for custodial staking addresses.",
	DefaultMemberPermissions: &ADMIN,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "url",
			Description: "The URL to the API",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "user_address",
			Description: "The name of the field that contains the user's address, i.e. `address`",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "custodian_address",
			Description: "The name of the field that contains the custodian address, i.e. `stakeAddress`",
			Required:    true,
		},
	},
}

var CONFIGURE_CUSTODIAL_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := GetOptions(i)

	url, err := url.Parse(options["url"].StringValue())
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Invalid URL:\n%s", err.Error()),
				Flags:   discordgo.MessageFlagsEphemeral,
				Title:   "Invalid URL",
			},
		})
	}

	response, err := http.Get(url.String())
	responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        slog.Error("invalid response body", "error", err)
    }

	user_address := options["user_address"].StringValue()
	custodian_address := options["custodian_address"].StringValue()

	var data []map[string]interface{}
	err = json.Unmarshal([]byte(responseData), &data)
	if err != nil {
		slog.Error("could not unmarshal response body", "error", err)
	}

	var foundAddress bool
	var foundCustodianAddress bool
	for _, pair := range data {
		if _, ok := pair[user_address]; ok {
			foundAddress = true
		}

		if _, ok := pair[custodian_address]; ok {
			foundCustodianAddress = true
		}
	}

	// slog.Info("success", "address", data[0]["address"].(string))
	// slog.Info("success", "stakingAddress", data[0]["stakingAddress"].(string))

	if (!foundAddress || !foundCustodianAddress) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("The provided user_address %s and/or custodian_address %s do not exist at the given endpoint. Please review and retry.", user_address, custodian_address),
				Flags:   discordgo.MessageFlagsEphemeral,
				Title:   "Creating Links",
			},
		})

		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Adding custodian endpoint:\nurl: %s\nuser_address: %s\ncustodian_address: %s", url, user_address, custodian_address),
			Flags:   discordgo.MessageFlagsEphemeral,
			Title:   "Creating Links",
		},
	})

	custodian := preeb.Custodian{
		Url:              *url,
		UserAddress:      user_address,
		CustodianAddress: custodian_address,
	}

	config := preeb.LoadConfig(i.GuildID)
	config.Custodians = append(config.Custodians, custodian)

	config.Save()
}
