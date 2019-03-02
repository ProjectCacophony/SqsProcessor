package gall

import (
	"time"

	"github.com/lib/pq"

	"github.com/jinzhu/gorm"
)

type Entry struct {
	gorm.Model
	GuildID   string
	ChannelID string
	AddedBy   string

	BoardID      string
	MinorGallery bool
	Recommended  bool
	LastCheck    time.Time
}

func (*Entry) TableName() string {
	return "gall_entries"
}

// Post model maintained by Worker
type Post struct {
	gorm.Model
	EntryID uint

	PostID     string
	MessageIDs pq.StringArray `gorm:"type:varchar[]"`
}

func (*Post) TableName() string {
	return "gall_posts"
}