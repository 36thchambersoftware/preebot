package discord

import (
	"bytes"
	"preebot/pkg/preeb"
	"strconv"
	"text/template"

	"github.com/bwmarrin/discordgo"
)

func CheckUserWallets(i *discordgo.InteractionCreate) {
	user := preeb.LoadUser(i.Member.User.ID)

	if len(user.Wallets) == 0 {
		content := "You need to link your wallet first. Please use /link-wallet."
		S.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	roleIDs, err := AssignQualifiedRoles(i.GuildID, i.Member.User.ID)
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