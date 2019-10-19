package eventlog

import (
	"github.com/bwmarrin/discordgo"
	"gitlab.com/Cacophony/go-kit/events"
)

func compareEmojiDiff(diff *events.DiffEmoji) (new []*discordgo.Emoji, updated [][]*discordgo.Emoji, deleted []*discordgo.Emoji) {
	for _, oldEmoji := range diff.Old {
		newEmoji := emojiSliceFindEmoji(oldEmoji.ID, diff.New)
		if newEmoji != nil {
			if !emojiEqual(oldEmoji, newEmoji) {
				updated = append(updated, []*discordgo.Emoji{oldEmoji, newEmoji})
			}
			continue
		}

		deleted = append(deleted, oldEmoji)
	}

	for _, newEmoji := range diff.New {
		if emojiSliceFindEmoji(newEmoji.ID, diff.Old) == nil {
			new = append(new, newEmoji)
		}
	}

	return new, updated, deleted
}

func emojiSliceFindEmoji(id string, list []*discordgo.Emoji) *discordgo.Emoji {
	for _, emoji := range list {
		if emoji.ID == id {
			return emoji
		}
	}

	return nil
}

func emojiEqual(a, b *discordgo.Emoji) bool {
	if a.ID != b.ID {
		return false
	}

	if a.Name != b.Name {
		return false
	}

	if a.Managed != b.Managed {
		return false
	}

	if !stringSliceMatches(a.Roles, b.Roles) {
		return false
	}

	return true
}
