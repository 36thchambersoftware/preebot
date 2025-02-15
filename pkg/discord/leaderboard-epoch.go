package discord

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"preebot/pkg/blockfrost"
	"preebot/pkg/koios"
	"preebot/pkg/preeb"
	"slices"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

var LEADERBOARD_EPOCH_COMMAND = discordgo.ApplicationCommand{
	Version:                  "0.01",
	Name:                     "leaderboard-epoch",
	Description:              "Check the Epoch Leaderboard",
	DefaultMemberPermissions: &ADMIN,
}

var LEADERBOARD_EPOCH_HANDLER = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	config := preeb.LoadConfig(i.GuildID)
	var poolID string
	for k, _ := range config.PoolIDs {
		poolID = k
		break
	}

	if len(config.PoolIDs) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No pools configured! Run `/configure-pool-id`",
				Title:   "Epoch Leaderboard",
			},
		})
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	tip, err := koios.Tip(ctx)
	if err != nil {
		msg := "koios is having a tantrum - try again later"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
	}
	currentEpoch := int(tip.EpochNo)

	m, err := blockfrost.GetPoolMetaData(ctx, poolID)

	users := preeb.LoadUsers()

	type Leader struct {
		ID string
		ActiveEpoch int
	}
	leaderboard := []Leader{}

	for _, user := range users {
		leader := Leader{
			ID: user.ID,
		}

		stakeAddresses := slices.Collect(maps.Keys(user.Wallets))
		var epochs []int
		for _, stakeAddress := range stakeAddresses {
			epoch, err := blockfrost.EpochsDelegatedToPool(ctx, string(stakeAddress), poolID)
			if err != nil {
				slog.Error("unable to get history", "stake", stakeAddress, "ERROR", err)
				continue
			}

			if epoch != nil {
				epochs = append(epochs, *epoch)
			}
		}

		if len(epochs) > 0 {
			slices.Sort(epochs)
			leader.ActiveEpoch = epochs[0]
			if leader.ActiveEpoch > 0 {
				leaderboard = append(leaderboard, leader)
			}
		}
	}

	sort.Slice(leaderboard, func(i int, j int) bool{
		return leaderboard[i].ActiveEpoch < leaderboard[j].ActiveEpoch
	})

	// sentence := "## Epoch Leaderboard\n"
	// for _, leader := range leaderboard {
	// 	p := message.NewPrinter(language.English)
	// 	sentence = sentence + p.Sprintf(" <@%s>\t Active Epoch: %d\t Total Epochs: %d\n", leader.ID, leader.ActiveEpoch - 2, currentEpoch - leader.ActiveEpoch + 2)
	// }

	var names, activeEpoch, totalsEpochs string

	for _, leader := range leaderboard {
		names += fmt.Sprintf("<@%s>\n", leader.ID)
		activeEpoch += fmt.Sprintf("%d\n", leader.ActiveEpoch)
		totalsEpochs += fmt.Sprintf("%d\n", currentEpoch - leader.ActiveEpoch)
	}

	embed := discordgo.MessageEmbed{
		Title: "Epoch Leaderboard",
		Description: fmt.Sprintf("Length in epochs staked to %s", *m.Name),
		Color: 0x7289DA,
		Footer:      &discordgo.MessageEmbedFooter{Text: "PREEBOT thanks you for delegating!"},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: "https://preeb.cloud/wp-content/uploads/2024/06/Transparent-png.png", Height: 50, Width: 50},
		Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
		Fields:      []*discordgo.MessageEmbedField{
			{
				Name:   "Member",
				Value: names,
				Inline: true,
			},
			{
				Name:   "Active Epoch",
				Value:  activeEpoch,
				Inline: true,
			},
			{
				Name:   "Total Epochs",
				Value:  totalsEpochs,
				Inline: true,
			},
		},
	}

	_, err = s.ChannelMessageSendEmbed(i.ChannelID, &embed)
	if err != nil {
		slog.Error("could not send message embed", "ERROR", err)
	}
}