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

func compareWebhooksDiff(diff *events.DiffWebhooks) (new []*discordgo.Webhook, updated [][]*discordgo.Webhook, deleted []*discordgo.Webhook) {
	for _, oldWebhok := range diff.Old {
		newWebhook := webhookSliceFindWebhook(oldWebhok.ID, diff.New)
		if newWebhook != nil {
			if !webhookEqual(oldWebhok, newWebhook) {
				updated = append(updated, []*discordgo.Webhook{oldWebhok, newWebhook})
			}
			continue
		}

		deleted = append(deleted, oldWebhok)
	}

	for _, newWebhook := range diff.New {
		if webhookSliceFindWebhook(newWebhook.ID, diff.Old) == nil {
			new = append(new, newWebhook)
		}
	}

	return new, updated, deleted
}

func webhookSliceFindWebhook(id string, list []*discordgo.Webhook) *discordgo.Webhook {
	for _, item := range list {
		if item.ID == id {
			return item
		}
	}

	return nil
}

func webhookEqual(a, b *discordgo.Webhook) bool {
	if a.ID != b.ID {
		return false
	}

	if a.ChannelID != b.ChannelID {
		return false
	}

	if a.Name != b.Name {
		return false
	}

	if a.Avatar != b.Avatar {
		return false
	}

	return true
}

func compareInvitesDiff(diff *events.DiffInvites) (new []*discordgo.Invite, updated [][]*discordgo.Invite, deleted []*discordgo.Invite) {
	for _, oldInvite := range diff.Old {
		newInvite := inviteSliceFindInvite(oldInvite.Code, diff.New)
		if newInvite != nil {
			if !inviteEqual(oldInvite, newInvite) {
				updated = append(updated, []*discordgo.Invite{oldInvite, newInvite})
			}
			continue
		}

		deleted = append(deleted, oldInvite)
	}

	for _, newInvite := range diff.New {
		if inviteSliceFindInvite(newInvite.Code, diff.Old) == nil {
			new = append(new, newInvite)
		}
	}

	return new, updated, deleted
}

func inviteSliceFindInvite(code string, list []*discordgo.Invite) *discordgo.Invite {
	for _, invite := range list {
		if invite.Code == code {
			return invite
		}
	}

	return nil
}

func inviteEqual(a, b *discordgo.Invite) bool {
	if a.Code != b.Code {
		return false
	}

	if a.Revoked != b.Revoked {
		return false
	}

	return true
}
