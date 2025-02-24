package discord

import (
	"context"
	"fmt"
	"maps"
	"preebot/pkg/blockfrost"
	"preebot/pkg/koios"
	"preebot/pkg/logger"
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
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	l := logger.Record
	l = l.WithGroup("LEADERBOARD-EPOCH")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	tip, err := koios.Tip(ctx)
	if err != nil {
		l.Error("could not get tip data", "ERROR", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "koios is having a tantrum - try again later",
				Title:   "Epoch Leaderboard",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	currentEpoch := int(tip.EpochNo)

	m, err := blockfrost.GetPoolMetaData(ctx, poolID)
	if err != nil {
		l.Error("could not get pool data", "ERROR", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "blockfrost is having a tantrum - try again later",
				Title:   "Epoch Leaderboard",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading leaderboard ...",
			Title:   "Epoch Leaderboard",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	users := preeb.LoadUsers()

	type Leader struct {
		ID string
		ActiveEpoch int
	}
	leaderboard := []Leader{}

	members, err := S.GuildMembers(i.GuildID, "0", 1000)
	if err != nil {
		l.Error("could not get members", "GuildID", i.GuildID, "ERROR", err)
	}

	for _, user := range users {
		found := false
		for _, member := range members {
			if member.User.ID == user.ID {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		leader := Leader{
			ID: user.ID,
		}

		stakeAddresses := slices.Collect(maps.Keys(user.Wallets))
		var epochs []int
		for _, stakeAddress := range stakeAddresses {
			l.Info("getting history", "USER", user.ID, "STAKE", string(stakeAddress))
			epoch, err := blockfrost.EpochsDelegatedToPool(ctx, string(stakeAddress), poolID)
			if err != nil {
				l.Error("unable to get history", "stake", stakeAddress, "ERROR", err)
				continue
			}

			if epoch != nil && *epoch != 0 {
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

	var leaderFields []*discordgo.MessageEmbedField
	var leaderCount uint
	for _, leader := range leaderboard {
		if (currentEpoch - leader.ActiveEpoch > 1) && leaderCount <= 20 {
			leaderCount++
			field := discordgo.MessageEmbedField{
				Name: "",
				Value: fmt.Sprintf("%d. <@%s> `Total Epochs: %d`", leaderCount, leader.ID, currentEpoch - leader.ActiveEpoch),
				Inline: false,
			}
			leaderFields = append(leaderFields, &field)
		}
	}

	embed := discordgo.MessageEmbed{
		Title: "Epoch Leaderboard (Top 20)",
		Description: fmt.Sprintf("Length in epochs staked to %s\nNOTE: It take 2 epochs to become an active delegator.", *m.Name),
		Color: 0x58d68d,
		Footer:      &discordgo.MessageEmbedFooter{Text: "PREEBOT thanks you for delegating!"},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: preeb.LogoImage, Height: 50, Width: 50},
		Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
		Fields:      leaderFields,
	}

	_, err = s.ChannelMessageSendEmbed(i.ChannelID, &embed)
	if err != nil {
		l.Error("could not send message embed", "ERROR", err)
	}

	s.InteractionResponseDelete(i.Interaction)
}