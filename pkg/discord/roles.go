package discord

import (
	"github.com/bwmarrin/discordgo"
)

func FindRoleByName(roles []*discordgo.Role, name string) *discordgo.Role {
	var desiredRole *discordgo.Role

	for _, role := range roles {
		if role.Name == name {
			desiredRole = role
		}
	}

	return desiredRole
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
