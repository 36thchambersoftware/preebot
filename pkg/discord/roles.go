package discord

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"

	"github.com/bwmarrin/discordgo"
)

func FindRoleByName(s *discordgo.Session, i *discordgo.InteractionCreate, name string) (*discordgo.Role, error) {
	var desiredRole *discordgo.Role

	perms, err := s.GuildRoles(i.GuildID)
	if err != nil {
		return nil, err
	}

	for _, role := range perms {
		if role.Name == name {
			desiredRole = role
		}
	}

	return desiredRole, nil
}

func UserHasRole(memberRoles []string, role discordgo.Role) bool {
	var user_has_role bool
	for _, r := range memberRoles {
		if r == role.ID {
			user_has_role = true
			break
		}
	}

	return user_has_role
}

func ToggleRole(s *discordgo.Session, i *discordgo.InteractionCreate, role *discordgo.Role) error {
	var response string
	user_has_role := UserHasRole(i.Member.Roles, *role)
	if !user_has_role {
		err := s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, role.ID)
		if err != nil {
			return err
		}
		response = "Role added: <@&" + role.ID + ">"
	} else {
		err := s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, role.ID)
		if err != nil {
			return err
		}
		response = "Role removed: <@&" + role.ID + ">"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	return nil
}

func AssignRoleByRoleName(s *discordgo.Session, i *discordgo.InteractionCreate, roleName string) (*discordgo.Role, error) {
	role, err := FindRoleByName(s, i, roleName)
	if err != nil {
		return nil, err
	}

	err = s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func AssignDelegatorRole(s *discordgo.Session, i *discordgo.InteractionCreate, totalStake int) (*discordgo.Role, error) {
	roleName := preebot.GetTier(totalStake)

	if roleName != "" {
		role, err := AssignRoleByRoleName(s, i, roleName)
		if err != nil {
			return role, err
		}

		return role, nil

	}

	return nil, nil
}

func CheckDelegation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := preebot.LoadUser(i.Member.User.ID)
	config := preebot.LoadConfig(i.GuildID)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if len(user.Wallets) == 0 {
		content := "You need to link your wallet first. Please use /link-wallet."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	totalStake := blockfrost.GetTotalStake(ctx, config.PoolIDs, user.Wallets)
	role, err := AssignDelegatorRole(s, i, totalStake)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong! Maybe open a #support-ticket ",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	} else if role != nil {
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
		sentence := "After looking at your {{ .walletCount }} {{ .walletWord }}\n{{ .walletList }}You have been assigned a role! <@&{{ .role }}>"
		partial := template.Must(template.New("check-delegation-template").Parse(sentence))
		partial.Execute(&b, map[string]interface{}{
			"walletCount": len(user.Wallets),
			"walletWord":  walletWord,
			"walletList":  walletList,
			"role":        role.ID,
		})

		content := b.String()
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	} else {
		content := fmt.Sprintf("%+v", role)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}
}

func CheckAssets(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := preebot.LoadUser(i.Member.User.ID)
	config := preebot.LoadConfig(i.GuildID)

	if len(user.Wallets) == 0 {
		content := "You need to link your wallet first. Please use /link-wallet."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	if len(config.PolicyIDs) == 0 {
		content := "The administrator needs to set the policy ID first."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
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
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
