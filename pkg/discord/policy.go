package discord

import (
	"context"
	"fmt"
	"preebot/pkg/blockfrost"
	"preebot/pkg/logger"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// func automaticPolicyBuyNotifier(ctx context.Context) {
// 	configs := preeb.LoadConfigs()
// 	policyTxs := make(map[string]string)

// 	for _, config := range configs {
// 		for policyID, policy := range config.PolicyIDs {
// 			if policy.ChannelID != "" {
// 				transactions, err := koios.GetPolicyTxs(ctx, policyID)
// 				if err != nil {
// 					logger.Record.Warn("Could not get policy transactions", "policyID", policyID, "ERROR", err)
// 				}

// 				if txHash, ok := policyTxs[policyID]; ok {
// 					if txHash != "" && txHash != transactions[0].TxHash.String() {
// 						for _, transaction := range transactions {

// 						}

// 						// Set the latest transaction hash for the policy.
// 						policyTxs[policyID] = txHash
// 					}
// 				}

// 			}
// 		}
// 	}
// }

func AutomaticLaunchBuyNotifier(ctx context.Context, lastUpdateHash string) (string) {
	address := "addr1qxzdfv2aalyz0nltf3r4rk9ukzlupg57k04v25mlrrl5a2uj5dysang6xcyp62r6dwdm7pnv3nsdwwn7jzzhr03ur6tqpsnml0"
	var embedFields []*discordgo.MessageEmbedField

	txs, err := blockfrost.GetAddressTransactions(ctx, address)
	if err != nil {
		logger.Record.Warn("Could not get address transactions", "address", address, "ERROR", err)
	}

	for _, tx := range txs {
		if lastUpdateHash == "" {
			lastUpdateHash = tx.TxHash
			logger.Record.Warn("Notice", "LAST UPDATE HASH", lastUpdateHash)
			return lastUpdateHash
		}

		if tx.TxHash != lastUpdateHash {
			txInfo, err := blockfrost.GetTransaction(ctx, tx.TxHash)
			if err != nil {
				logger.Record.Warn("Could not get transaction details", "hash", tx.TxHash, "ERROR", err)
			}

			logger.Record.Warn("Notice", "CURRENT HASH", tx.TxHash)
			for _, output := range txInfo.Outputs {
				if output.Address == address {
					for _, amount := range output.Amount {
						if amount.Unit == "lovelace" {
							lovelace, err := strconv.Atoi(amount.Quantity)
							if err != nil {
								logger.Record.Warn("Could not get convert lovelace to ada", "lovelace", amount.Quantity, "ERROR", err)
							}
							embedField := discordgo.MessageEmbedField{
								Name:   "UPDATE",
								Value: fmt.Sprintf("**NEW PURCHASE!**\n\nAmount: %d ada\n\nBuyer: %s\n\nTxHash: %s", lovelace / blockfrost.LOVELACE, txInfo.Inputs[0].Address, tx.TxHash),	
								Inline: false,
							}

							embedFields = append(embedFields, &embedField)

							embed := discordgo.MessageEmbed{
								Title: "SKULLY",
								Color: 0xd269ff,
								Footer:      &discordgo.MessageEmbedFooter{Text: "PREEB thanks you for delegating!"},
								Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: "https://preeb.cloud/wp-content/uploads/2025/03/skullylfg.jpeg", Height: 50, Width: 50},
								Provider:    &discordgo.MessageEmbedProvider{Name: "PREEB"},
								Fields:      embedFields,
							}

							_, err = S.ChannelMessageSendEmbed("1191857041728360508", &embed)
							if err != nil {
								logger.Record.Error("could not send message embed", "ERROR", err)
							}
						}
					}

				}
			}
		} else {
			lastUpdateHash = txs[0].TxHash
			logger.Record.Warn("Notice", "LAST UPDATE HASH", lastUpdateHash)
			return lastUpdateHash
		}
	}

	return lastUpdateHash
}