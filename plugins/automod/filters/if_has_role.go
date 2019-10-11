package filters

import (
	"errors"
	"strings"

	"gitlab.com/Cacophony/Processor/plugins/automod/models"

	"gitlab.com/Cacophony/Processor/plugins/automod/interfaces"
)

type Role struct {
}

func (f Role) Name() string {
	return "if_has_role"
}

func (f Role) Args() int {
	return 1
}

func (f Role) Deprecated() bool {
	return false
}

func (f Role) NewItem(env *models.Env, args []string) (interfaces.FilterItemInterface, error) {
	if len(args) < 1 {
		return nil, errors.New("too few arguments")
	}

	var roleIDs []string

	for _, roleMention := range strings.Split(args[0], ",") {
		role, err := env.State.RoleFromMention(env.GuildID, roleMention)
		if err != nil {
			continue
		}

		roleIDs = append(roleIDs, role.ID)
	}

	if len(roleIDs) == 0 {
		return nil, errors.New("no Roles found")
	}

	return &RoleItem{
		RoleIDs: roleIDs,
	}, nil
}

func (f Role) Description() string {
	return "automod.filters.if_has_role"
}

type RoleItem struct {
	RoleIDs []string
}

func (f *RoleItem) Match(env *models.Env) bool {
	if env.Event == nil {
		return false
	}

	for _, userID := range env.UserID {
		member, err := env.State.Member(env.GuildID, userID)
		if err != nil {
			return false
		}

		for _, roleID := range f.RoleIDs {
			if matchChannels(member.Roles, roleID) {
				continue
			}

			return false
		}
	}

	return true
}
