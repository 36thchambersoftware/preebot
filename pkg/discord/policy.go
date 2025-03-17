package discord

import (
	"context"
	"fmt"
	"net/url"
	"preebot/pkg/logger"
	"preebot/pkg/preeb"
	"preebot/pkg/taptools"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func AutomaticBuyNotifier(ctx context.Context) {
	configs := preeb.LoadConfigs()
	logger.Record.Info("getting configs")
	for _, config := range configs {
		for policyID, policy := range config.PolicyIDs {
			if policy.Notify {
				logger.Record.Info("getting trades")
				trades, err := taptools.GetTrades(ctx, policyID + policy.HexName)
				if err != nil {
					logger.Record.Warn("Could not get trades", "policyID", policyID, "ERROR", err)
					continue
				}

				logger.Record.Info("getting buys")
				var buys taptools.Trades
				for _, trade := range trades {
					if (trade.Action == "buy") {
						buys = append(buys, trade)
						if LAST_UPDATE_TIME[policyID] == 0 {
							LAST_UPDATE_TIME[policyID] = trade.Time
							logger.Record.Info("Notice", "POLICY", policyID, "LAST UPDATE TIME", LAST_UPDATE_TIME[policyID])
						}

						logger.Record.Info("time", "NEW", trade.Time > LAST_UPDATE_TIME[policyID])
						if (trade.Time > LAST_UPDATE_TIME[policyID]) {
							logger.Record.Info("building embed")
							p := message.NewPrinter(language.English)
							embedField := discordgo.MessageEmbedField{
								Name:   p.Sprintf("Amount: %.0f %s / %.0f %s", trade.TokenAAmount, trade.TokenAName, trade.TokenBAmount, trade.TokenBName),
								Value: fmt.Sprintf("-# [Tx](https://cardanoscan.io/transaction/%s 'View Transaction')", trade.Hash),
								Inline: false,
							}

							var embedFields []*discordgo.MessageEmbedField
							embedFields = append(embedFields, &embedField)

							var image *url.URL
							var message string
							var label string
							alt_channel_id := policy.DefaultChannelID
							logger.Record.Info("buy matched to tier", "BUY NOTIS", policy.BuyNotifications)
							for _, n := range policy.BuyNotifications {
								logger.Record.Info("tier check", "min", n.Min, "max", n.Max, "amount", trade.TokenAAmount)
								if trade.TokenAAmount > float64(n.Min) && trade.TokenAAmount < float64(n.Max) {
									logger.Record.Info("buy matched to tier", "NOTIFICATION", n)
									alt_channel_id = n.ChannelID
									message = n.Message
									label = n.Label
									image, err = url.Parse(n.Image)
									if err != nil {
										logger.Record.Error("could not parse image url", "ERROR", err)
									}
								}
							}

							embed := discordgo.MessageEmbed{
								Title:       fmt.Sprintf("New %s $%s Buy!", label, trade.TokenAName),
								Description: message,
								Color: 		 0xd269ff,
								Footer:      &discordgo.MessageEmbedFooter{Text: "PREEB thanks you for delegating!"},
								Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
								Fields:      embedFields,
							}

							if image.String() != "" {
								logger.Record.Info("image", "URL", image.String())
								embed.Image = &discordgo.MessageEmbedImage{URL: image.String()}
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
				}

				if buys != nil {
					LAST_UPDATE_TIME[policyID] = buys[0].Time
					logger.Record.Info("Notice", "POLICY", policyID, "LAST UPDATE HASH", LAST_UPDATE_TIME)
				}
			}
		}
	}
}

// func AutomaticBuyNotifier(ctx context.Context) {
// 	configs := preeb.LoadConfigs()

// 	for _, config := range configs {
// 		for policyID, policy := range config.PolicyIDs {
// 			if policy.ChannelID != "" {
// 				policyInfo, err := blockfrost.AssetInfo(ctx, )
// 				txs, err := koios.GetPolicyTxs(ctx, policyID)
// 				if err != nil {
// 					logger.Record.Warn("Could not get policy transactions", "policyID", policyID, "ERROR", err)
// 				}

// 				if LAST_UPDATE_HASH[policyID] == "" {
// 					LAST_UPDATE_HASH[policyID] = string(txs[0].TxHash)
// 					logger.Record.Warn("Notice", "POLICY", policyID, "LAST UPDATE HASH", LAST_UPDATE_HASH)
// 				}

// 				if txHash, ok := LAST_UPDATE_HASH[policyID]; ok {
// 					for _, tx := range txs {
// 						if (txHash != tx.TxHash.String()) {
// 							txInfo, err := blockfrost.GetTransaction(ctx, string(tx.TxHash))
// 							if err != nil {
// 								logger.Record.Warn("Could not get transaction details", "hash", tx.TxHash, "ERROR", err)
// 							}

// 							logger.Record.Warn("Notice", "CURRENT HASH", tx.TxHash)
// 							for _, output := range txInfo.Outputs {
// 								if output.Address == address {
// 									for _, amount := range output.Amount {
// 										if amount.Unit == "lovelace" {
// 											lovelace, err := strconv.Atoi(amount.Quantity)
// 											if err != nil {
// 												logger.Record.Warn("Could not get convert lovelace to ada", "lovelace", amount.Quantity, "ERROR", err)
// 											}
// 											embedField := discordgo.MessageEmbedField{
// 												Name:   "UPDATE",
// 												Value: fmt.Sprintf("**NEW PURCHASE!**\n\nAmount: %d ada\n\nBuyer: %s\n\nTxHash: %s", lovelace / blockfrost.LOVELACE, txInfo.Inputs[0].Address, tx.TxHash),
// 												Inline: false,
// 											}

// 											embedFields = append(embedFields, &embedField)

// 											embed := discordgo.MessageEmbed{
// 												Title: "SKULLY",
// 												Color: 0xd269ff,
// 												Footer:      &discordgo.MessageEmbedFooter{Text: "PREEB thanks you for delegating!"},
// 												Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: "https://preeb.cloud/wp-content/uploads/2025/03/skullylfg.jpeg", Height: 50, Width: 50},
// 												Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
// 												Fields:      embedFields,
// 											}

// 											_, err = S.ChannelMessageSendEmbed("1191857041728360508", &embed)
// 											if err != nil {
// 												logger.Record.Error("could not send message embed", "ERROR", err)
// 											}
// 										}
// 									}

// 								}
// 							}
// 						}
// 					}

// 					// Set the latest transaction hash for the policy.
// 					LAST_UPDATE_HASH[policyID] = txHash
// 				}

// 			}
// 		}
// 	}
// }

// func AutomaticLaunchBuyNotifier(ctx context.Context, lastUpdateHash string) (string) {
// 	address := "addr1qxzdfv2aalyz0nltf3r4rk9ukzlupg57k04v25mlrrl5a2uj5dysang6xcyp62r6dwdm7pnv3nsdwwn7jzzhr03ur6tqpsnml0"
// 	var embedFields []*discordgo.MessageEmbedField

// 	txs, err := blockfrost.GetAddressTransactions(ctx, address)
// 	if err != nil {
// 		logger.Record.Warn("Could not get address transactions", "address", address, "ERROR", err)
// 	}

// 	for _, tx := range txs {
// 		if lastUpdateHash == "" {
// 			lastUpdateHash = tx.TxHash
// 			logger.Record.Warn("Notice", "LAST UPDATE HASH", lastUpdateHash)
// 			return lastUpdateHash
// 		}

// 		if tx.TxHash != lastUpdateHash {
// 			txInfo, err := blockfrost.GetTransaction(ctx, tx.TxHash)
// 			if err != nil {
// 				logger.Record.Warn("Could not get transaction details", "hash", tx.TxHash, "ERROR", err)
// 			}

// 			logger.Record.Warn("Notice", "CURRENT HASH", tx.TxHash)
// 			for _, output := range txInfo.Outputs {
// 				if output.Address == address {
// 					for _, amount := range output.Amount {
// 						if amount.Unit == "lovelace" {
// 							lovelace, err := strconv.Atoi(amount.Quantity)
// 							if err != nil {
// 								logger.Record.Warn("Could not get convert lovelace to ada", "lovelace", amount.Quantity, "ERROR", err)
// 							}
// 							embedField := discordgo.MessageEmbedField{
// 								Name:   "UPDATE",
// 								Value: fmt.Sprintf("**NEW PURCHASE!**\n\nAmount: %d ada\n\nBuyer: %s\n\nTxHash: %s", lovelace / blockfrost.LOVELACE, txInfo.Inputs[0].Address, tx.TxHash),
// 								Inline: false,
// 							}

// 							embedFields = append(embedFields, &embedField)

// 							embed := discordgo.MessageEmbed{
// 								Title: "SKULLY",
// 								Color: 0xd269ff,
// 								Footer:      &discordgo.MessageEmbedFooter{Text: "PREEB thanks you for delegating!"},
// 								Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: "https://preeb.cloud/wp-content/uploads/2025/03/skullylfg.jpeg", Height: 50, Width: 50},
// 								Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
// 								Fields:      embedFields,
// 							}

// 							_, err = S.ChannelMessageSendEmbed("1191857041728360508", &embed)
// 							if err != nil {
// 								logger.Record.Error("could not send message embed", "ERROR", err)
// 							}
// 						}
// 					}

// 				}
// 			}
// 		} else {
// 			lastUpdateHash = txs[0].TxHash
// 			logger.Record.Warn("Notice", "LAST UPDATE HASH", lastUpdateHash)
// 			return lastUpdateHash
// 		}
// 	}

// 	return lastUpdateHash
// }