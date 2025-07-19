package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"text/template"
	"time"

	"preebot/pkg/blockfrost"
	"preebot/pkg/preeb"

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

func AssignQualifiedRoles(guildID string, user preeb.User) ([]string, error) {
	config := preeb.LoadConfig(guildID)

	// Get existing roles
	member, err := S.GuildMember(guildID, user.ID)
	if err != nil {
		if member == nil || member.GuildID == "" || member.GuildID != guildID {
			// User does not exist in this server, skip them.
			return nil, nil
		}
		return nil, err
	}

	slog.Info("Checking roles", "USER", user.ID, "GUILD", guildID)
	// Get qualified delegator roles
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	totalStake := blockfrost.GetTotalStake(ctx, config.PoolIDs, user.Wallets)
	delegatorRoleIDs := preeb.GetDelegatorRolesByStake(int(totalStake), config.DelegatorRoles)
	sort.Strings(delegatorRoleIDs)

	// Get qualified policy holder roles
	allAssets, err := blockfrost.GetAllUserAssets(ctx, user.Wallets)
	if err != nil {
		slog.Error("could not get user assets", "user", user.ID, "error", err)
		return nil, err
	}
	assetCount := blockfrost.CountUserAssetsByPolicy(ctx, config.PolicyIDs, allAssets)
	policyRoleIDs := preeb.GetPolicyRoles(assetCount, config.PolicyIDs)
	sort.Strings(policyRoleIDs)

	allQualifiedRoles := append(delegatorRoleIDs, policyRoleIDs...)

	uniqueRoles := make(map[string]bool)
	var rolesToAdd []string
	for _, roleID := range allQualifiedRoles {
		if !uniqueRoles[roleID] {
			rolesToAdd = append(rolesToAdd, roleID)
			uniqueRoles[roleID] = true
		}
	}



	// Get current roles
	var currentRoles []string
	for _, currentRoleID := range member.Roles {
		// if it's a delegator role ...
		if _, ok := config.DelegatorRoles[currentRoleID]; ok {
			currentRoles = append(currentRoles, currentRoleID)
			// Check if the user should have this role ...
			if (!slices.Contains(rolesToAdd, currentRoleID)) {
				// and remove it if not
				err := S.GuildMemberRoleRemove(guildID, user.ID, currentRoleID)
				if err != nil {
					slog.Error("could not remove role", "role", currentRoleID, "error", err)
					continue
				}
			}
		}

		// if it's a policy role ...
		for _, policy := range config.PolicyIDs {
			if _, ok := policy.Roles[currentRoleID]; ok {
				currentRoles = append(currentRoles, currentRoleID)
				// Check if the user should have this role ...
				if (!slices.Contains(rolesToAdd, currentRoleID)) {
					// and remove it if not
					slog.Info("Removing role", "GUILD", guildID, "ROLE", currentRoleID, "USER", user.ID)
					err := S.GuildMemberRoleRemove(guildID, user.ID, currentRoleID)
					if err != nil {
						slog.Error("could not remove role", "role", currentRoleID, "error", err)
						continue
					}
				}
			}
		}
	}

	sort.Strings(currentRoles)

	if reflect.DeepEqual(currentRoles, allQualifiedRoles) {
		return currentRoles, nil
	}

	var assignedRoles []string

	for _, roleID := range allQualifiedRoles {
		slog.Info("Assigning role", "GUILD", guildID, "ROLE", roleID, "USER", user.ID)
		role, err := AssignRoleByID(guildID, user.ID, roleID)
		if err != nil {
			return nil, err
		}

		assignedRoles = append(assignedRoles, role.ID)
	}


	return assignedRoles, nil
}

// Automatic Role Checker
func AutomaticRoleChecker() {
	// Get guild id
	configs := preeb.LoadConfigs()

	// Get verified guild members
	users := preeb.LoadUsers()

	for _, config := range configs {
		slog.Info("Checking roles for guild", "GUILD", config.GuildID)

		// Get guild member linked wallets
		for _, user := range users {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			// Check roles
			_, err := AssignQualifiedRoles(config.GuildID, user)
			if err != nil {
				slog.Error("could not assign roles", "user", user.ID, "error", err)
			}
		}
	}
}

func FormatNewRolesMessage(user preeb.User, roleIDs []string) (string) {
	walletWord := "wallet"
	if len(user.Wallets) > 1 {
		walletWord = "wallets"
	}

	walletList := ""
	n := 0
	for _, addr := range user.Wallets {
		n++
		walletList = walletList + strconv.Itoa(n) + ". -# " + string(addr) + "\n"
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
	return content
}

func loadCustodianData(c []preeb.Custodian) {
	for _, custodian := range c {
		response, err := http.Get(custodian.Url.String())
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			slog.Error("invalid response body", "error", err)
		}

		var data []map[string]interface{}
		err = json.Unmarshal([]byte(responseData), &data)
		if err != nil {
			slog.Error("could not unmarshal response body", "error", err)
		}

		custodian_addresses := make(map[string]string)
		for _, pair := range data {
			if _, ok := pair[custodian.UserAddress]; !ok {
				break
			}
	
			if _, ok := pair[custodian.CustodianAddress]; !ok {
				break
			}

			custodian_addresses[custodian.UserAddress] = custodian.CustodianAddress
		}
		CUSTODIAN_ADDRESSES = custodian_addresses
	}
}