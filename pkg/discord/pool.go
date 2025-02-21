package discord

import (
	"context"
	"fmt"
	"preebot/pkg/blockfrost"
	"preebot/pkg/logger"
	"preebot/pkg/preeb"

	"github.com/bwmarrin/discordgo"
)

func AutomaticPoolBlocks(ctx context.Context, lastBlock string) (string) {
	configs := preeb.LoadConfigs()
	var embedFields []*discordgo.MessageEmbedField
	for _, config := range configs {
		for poolID, active := range config.PoolIDs {
			if active {
				blocks, err := blockfrost.PoolBlocks(ctx, poolID)
				if err != nil {
					logger.Record.Warn("Could not get blocks", "poolID", poolID, "ERROR", err)
					continue
				}

				if (len(blocks) > 0 && blocks[0] != lastBlock && lastBlock != "") {
					lastBlock = blocks[0]
					meta, err := blockfrost.PoolMeta(ctx, poolID)
					if err != nil {
						logger.Record.Warn("Could not get pool meta", "poolID", poolID, "ERROR", err)
					}

					info, err := blockfrost.PoolInfo(ctx, poolID)
					if err != nil {
						logger.Record.Warn("Could not get pool info", "poolID", poolID, "ERROR", err)
					}

					embedField := discordgo.MessageEmbedField{
						Name:   *meta.Ticker,
						Value: fmt.Sprintf("**NEW BLOCK!**\n\n%s\n\nTotal Blocks: %d", *meta.Description, info.BlocksMinted),	
						Inline: false,
					}
					embedFields = append(embedFields, &embedField)
				} else if (len(blocks) > 0) {
					lastBlock = blocks[0]
				}
			}
		}

		if len(embedFields) > 0 {
			embed := discordgo.MessageEmbed{
				Title: "Pool Info",
				Description: fmt.Sprintf("Update!"),
				Color: 0x58d68d,
				Footer:      &discordgo.MessageEmbedFooter{Text: "PREEBOT thanks you for delegating!"},
				Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: preeb.Logo, Height: 50, Width: 50},
				Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
				Fields:      embedFields,
			}

			// TODO create configure channel for pool info
			_, err := S.ChannelMessageSendEmbed("992429789774360617", &embed)
			if err != nil {
				logger.Record.Error("could not send message embed", "ERROR", err)
			}
		}
	}

	return lastBlock
}