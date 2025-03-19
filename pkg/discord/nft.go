package discord

import (
	"context"
	"fmt"
	"net/url"
	"preebot/pkg/logger"
	"preebot/pkg/preeb"
	"preebot/pkg/taptools"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func AutomaticNFTBuyNotifier(ctx context.Context) {
	logger.Record.Info("getting nft trades")
	trades, err := taptools.GetNFTTrades(ctx)
	if err != nil {
		logger.Record.Warn("Could not get nft trades", "ERROR", err)
		return
	}
	configs := preeb.LoadConfigs()
	logger.Record.Info("getting configs")
	for _, config := range configs {
		for policyID, policy := range config.PolicyIDs {
			if policy.NFT && policy.Notify {
				var buys taptools.NftTrades
				logger.Record.Info("checking nft trades")
				if LAST_UPDATE_TIME[policyID] == 0 {
					now := time.Now()
					LAST_UPDATE_TIME[policyID] = int(now.Unix())
					logger.Record.Info("Notice", "POLICY", policyID, "LAST UPDATE TIME", LAST_UPDATE_TIME[policyID])
				}

				for _, trade := range trades {
					if trade.Policy != policyID {
						continue
					}

					buys = append(buys, trade)
					logger.Record.Info("time", "NEW", trade.Time > LAST_UPDATE_TIME[policyID])
					if (trade.Time > LAST_UPDATE_TIME[policyID]) {
						logger.Record.Info("building embed")
						p := message.NewPrinter(language.English)
						var embedFields []*discordgo.MessageEmbedField

						embedFields = append(embedFields, &discordgo.MessageEmbedField{
							Name:   p.Sprintf("%s", trade.CollectionName),
							Value:  p.Sprintf("%s", trade.Name),
							Inline: false,
						})

						embedFields = append(embedFields, &discordgo.MessageEmbedField{
							Name:   p.Sprintf("Amount: %d ada", trade.Price),
							Value: fmt.Sprintf("-# [Tx](https://cardanoscan.io/transaction/%s 'View Transaction')", trade.Hash),
							Inline: false,
						})

						message := fmt.Sprintf("Check out the latest buy! Get yours at [jpg.store](https://jpg.store/collection/%s) ", trade.Policy)
						alt_channel_id := policy.DefaultChannelID

						for _, n := range policy.BuyNotifications {
							logger.Record.Info("tier check", "min", n.Min, "price", trade.Price)
							if float64(trade.Price) > float64(n.Min) {
								logger.Record.Info("buy matched to tier", "NOTIFICATION", n)
								alt_channel_id = n.ChannelID
								if n.Message != "" {
									message = n.Message
								}
								if err != nil {
									logger.Record.Error("could not parse image url", "ERROR", err)
								}
							}
						}

						// Trade urls: "ipfs://QmboJKkYbfyPrrD7pnvgRUjd4VXPo6kTfv2W7oVF4q3F52"
						// Converted urls: https://ipfs.io/ipfs/QmboJKkYbfyPrrD7pnvgRUjd4VXPo6kTfv2W7oVF4q3F52
						image := strings.Replace(trade.Image, "///", "//", 1)
						tokenURI, err := url.Parse(trade.Image)
						if err == nil {
							tokenURI.Scheme = "https"
							tokenURI.Path = fmt.Sprintf("ipfs/%s%s", tokenURI.Host, tokenURI.Path)
							tokenURI.Host = "ipfs.io"
							image = tokenURI.String()
							logger.Record.Info("ipfs conversion", "url", image)
						} else {
							logger.Record.Info("could not convert ipfs to standard", "ERROR", err)
						}

						embed := discordgo.MessageEmbed{
							Title:       "New Collection Buy!",
							Description: message,
							Color: 		 0xd269ff,
							Image:       &discordgo.MessageEmbedImage{URL: image},
							Footer:      &discordgo.MessageEmbedFooter{Text: "PREEB thanks you for delegating!"},
							Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
							Fields:      embedFields,
						}

						_, err = S.ChannelMessageSendEmbed(policy.DefaultChannelID, &embed)
						if err != nil {
							logger.Record.Error("could not send message embed", "ERROR", err)
						}

						if alt_channel_id != policy.DefaultChannelID {
							_, err = S.ChannelMessageSendEmbed(alt_channel_id, &embed)
							if err != nil {
								logger.Record.Error("could not send message embed", "ERROR", err)
							}
						}
					}
				}

				if buys != nil {
					LAST_UPDATE_TIME[policyID] = buys[0].Time
					logger.Record.Info("Notice", "POLICY", policyID, "LAST UPDATE HASH", LAST_UPDATE_TIME)
				}
			}
		}
	}
}
