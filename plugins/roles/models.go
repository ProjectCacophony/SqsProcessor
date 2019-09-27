package roles

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Category struct {
	gorm.Model

	Name      string
	GuildID   string
	ChannelID string // the channel this Category will listen to for role assignments
	Roles     []Role `gorm:"foreignkey:CategoryID"`
	Enabled   bool
	Limit     int // 0 = no limit
}

type Role struct {
	gorm.Model

	CategoryID   uint // roles aren't guarenteed to have a category
	GuildID      string
	ServerRoleID string
	PrintName    string
	Enabled      bool
	Aliases      pq.StringArray `gorm:"type:varchar[]"`
	// Reactions []string
}

func (*Category) TableName() string {
	return "role_categories"
}

func (*Role) TableName() string {
	return "roles"
}
