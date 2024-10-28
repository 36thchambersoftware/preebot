package discord

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"text/template"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slog"
)

func FindRoleByName(i *discordgo.InteractionCreate, name string) (*discordgo.Role, error) {
	var desiredRole *discordgo.Role

	perms, err := S.GuildRoles(i.GuildID)
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

func FindRoleByRoleID(guildID, id string) (*discordgo.Role, error) {
	var desiredRole *discordgo.Role

	perms, err := S.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	for _, role := range perms {
		if role.ID == id {
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

func ToggleRole(i *discordgo.InteractionCreate, role *discordgo.Role) error {
	var response string
	user_has_role := UserHasRole(i.Member.Roles, *role)
	if !user_has_role {
		err := S.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, role.ID)
		if err != nil {
			return err
		}
		response = "Role added: <@&" + role.ID + ">"
	} else {
		err := S.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, role.ID)
		if err != nil {
			return err
		}
		response = "Role removed: <@&" + role.ID + ">"
	}

	S.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	return nil
}

func AssignRoleByRoleName(i *discordgo.InteractionCreate, roleName string) (*discordgo.Role, error) {
	role, err := FindRoleByName(i, roleName)
	if err != nil {
		return nil, err
	}

	err = S.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func AssignRoleByID(guildID, userID, roleID string) (*discordgo.Role, error) {
	role, err := FindRoleByRoleID(guildID, roleID)
	if err != nil {
		return nil, err
	}

	err = S.GuildMemberRoleAdd(guildID, userID, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func AssignDelegatorRole(guildID, userID string, totalStake int) ([]string, error) {
	config := preebot.LoadConfig(guildID)

	// Get existing roles
	member, err := S.GuildMember(guildID, userID)
	if err != nil {
		return nil, err
	}

	// Clear existing roles
	var currentRoles []string
	for _, currentRoleID := range member.Roles {
		_, ok := config.DelegatorRoles[currentRoleID]
		if ok {
			currentRoles = append(currentRoles, currentRoleID)
			// err := S.GuildMemberRoleRemove(guildID, userID, currentRoleID)
			// if err != nil {
			// 	slog.Error("could not remove role", "role", currentRoleID, "error", err)
			// 	continue
			// }
		}
	}

	// Add newly accounted for roles
	roleIDs := preebot.GetDelegatorRolesByStake(totalStake, config.DelegatorRoles)

	if reflect.DeepEqual(currentRoles, roleIDs) {
		return currentRoles, nil
	}

	var assignedRoles []string
	if roleIDs != nil {
		for _, roleID := range roleIDs {
			role, err := AssignRoleByID(guildID, userID, roleID)
			if err != nil {
				return nil, err
			}

			assignedRoles = append(assignedRoles, role.ID)
		}
	}

	return assignedRoles, nil
}

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
}

// Automatic Role Checker
func AutomaticRoleChecker() {
	// Get guild id
	configs := preebot.LoadConfigs()
	for _, config := range configs {
		// Get verified guild members
		users := preebot.LoadUsers()

		// Get guild member linked wallets
		for _, user := range users {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			// Check delegation
			totalAda := blockfrost.GetTotalStake(ctx, config.PoolIDs, user.Wallets)
			_, err := AssignDelegatorRole(config.GuildID, user.ID, int(totalAda))
			if err != nil {
				slog.Error("could not assign roles", "user", user.ID, "error", err)
			}
		}
	}
}
