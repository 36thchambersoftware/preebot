package discord

import (
	"preebot/pkg/blockfrost"
	"preebot/pkg/preebot"

	bfg "github.com/blockfrost/blockfrost-go"

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
	user_has_role := UserHasRole(i.Member.Roles, *role)
	if !user_has_role {
		err := s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, role.ID)
		if err != nil {
			return err
		}
	} else {
		err := s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, role.ID)
		if err != nil {
			return err
		}
	}

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

func AssignDelegatorRole(s *discordgo.Session, i *discordgo.InteractionCreate, accountDetails bfg.Account) (*discordgo.Role, error) {
	if *accountDetails.PoolID == blockfrost.PREEB_POOL_ID {
		roleName := preebot.GetTier(accountDetails.ActiveEpoch, accountDetails.ControlledAmount)

		if roleName != "" {
			role, err := AssignRoleByRoleName(s, i, roleName)
			if err != nil {
				return role, err
			}

			return role, nil
		}
	}

	return nil, nil
}
