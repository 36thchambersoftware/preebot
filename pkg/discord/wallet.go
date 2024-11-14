package discord

import (
	"bytes"
	"context"
	"fmt"
	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"
	"strconv"
	"text/template"
	"time"

	"github.com/bwmarrin/discordgo"
)

func CheckDelegation(i *discordgo.InteractionCreate) {
	user := preebot.LoadUser(i.Member.User.ID)
	config := preebot.LoadConfig(i.GuildID)

	if len(user.Wallets) == 0 {
		content := "You need to link your wallet first. Please use /link-wallet."
		S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	totalAda := blockfrost.GetTotalStake(ctx, config.PoolIDs, user.Wallets)
	roleIDs, err := AssignDelegatorRole(i.GuildID, i.Member.User.ID, int(totalAda))
	if err != nil {
		S.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})

		return
	}

	walletWord := "wallet"
	if len(user.Wallets) > 1 {
		walletWord = "wallets"
	}

	walletList := ""
	n := 0
	for _, stakeAddress := range user.Wallets {
		for _, addr := range stakeAddress {
			n++
			walletList = walletList + strconv.Itoa(n) + ". -# " + string(addr) + "\n"
		}
	}

	var b bytes.Buffer
	sentence := "After looking at your {{ .walletCount }} {{ .walletWord }}\n{{ .walletList }}"

	if roleIDs != nil {
		sentence = sentence + "You have been assigned the following!\n"
		for _, roleID := range roleIDs {
			sentence = sentence + "<@&" + roleID + ">\n"
		}
	} else {
		sentence = sentence + "You don't qualify for any roles."
	}

	partial := template.Must(template.New("check-delegation-template").Parse(sentence))
	partial.Execute(&b, map[string]interface{}{
		"walletCount": len(user.Wallets),
		"walletWord":  walletWord,
		"walletList":  walletList,
	})

	content := b.String()

	S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}

func CheckAssets(i *discordgo.InteractionCreate) {
	user := preebot.LoadUser(i.Member.User.ID)
	config := preebot.LoadConfig(i.GuildID)

	if len(user.Wallets) == 0 {
		content := "You need to link your wallet first. Please use /link-wallet."
		S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	if len(config.PolicyIDs) == 0 {
		content := "The administrator needs to set the policy ID first."
		S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	allAddresses, err := blockfrost.GetAllUserAddresses(ctx, user.Wallets)
	if err != nil {
		// TODO
	}

	assetCount := blockfrost.CountUserAssetsByPolicy(config.PolicyIDs, allAddresses)

	content := fmt.Sprintf("You have %d nfts associated with the policy IDs of this server", assetCount)
	S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	roleIDs, err := AssignPolicyRole(i.GuildID, user.ID, assetCount)
	if err != nil {
		S.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})

		return
	}

	walletWord := "wallet"
	if len(user.Wallets) > 1 {
		walletWord = "wallets"
	}

	walletList := ""
	n := 0
	for _, stakeAddress := range user.Wallets {
		for _, addr := range stakeAddress {
			n++
			walletList = walletList + strconv.Itoa(n) + ". -# " + string(addr) + "\n"
		}
	}

	var b bytes.Buffer
	sentence := "After looking at your {{ .walletCount }} {{ .walletWord }}\n{{ .walletList }}"

	if roleIDs != nil {
		sentence = sentence + "You have been assigned the following!\n"
		for _, roleID := range roleIDs {
			sentence = sentence + "<@&" + roleID + ">\n"
		}
	} else {
		sentence = sentence + "You don't qualify for any roles."
	}

	partial := template.Must(template.New("check-policy-template").Parse(sentence))
	partial.Execute(&b, map[string]interface{}{
		"walletCount": len(user.Wallets),
		"walletWord":  walletWord,
		"walletList":  walletList,
	})

	content = b.String()

	S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}