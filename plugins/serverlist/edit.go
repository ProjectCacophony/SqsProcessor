package serverlist

import (
	"strings"

	"gitlab.com/Cacophony/go-kit/discord"

	"gitlab.com/Cacophony/go-kit/state"

	"gitlab.com/Cacophony/go-kit/permissions"

	"github.com/lib/pq"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/regexp"
)

// nolint: gocyclo
func (p *Plugin) handleEdit(event *events.Event) {
	if len(event.Fields()) < 3 {
		event.Respond("serverlist.edit.too-few-args-lt3") // nolint: errcheck
		return
	}
	if len(event.Fields()) < 4 {
		event.Respond("serverlist.edit.too-few-args-lt4") // nolint: errcheck
		return
	}
	if len(event.Fields()) < 5 {
		event.Respond("serverlist.edit.too-few-args-lt5") // nolint: errcheck
		return
	}

	values := event.Fields()[4:]

	server := extractExistingServerFromArg(p.redis, p.db, event.Discord(), event.Fields()[2])
	if server == nil {
		event.Respond("serverlist.edit.no-server") // nolint: errcheck
		return
	}

	if !stringSliceContains(event.UserID, server.EditorUserIDs) &&
		!event.Has(permissions.BotOwner) {
		event.Respond("serverlist.edit.no-editor") // nolint: errcheck
		return
	}

	if server.State == StateCensored {
		event.Respond("serverlist.edit.wrong-status") // nolint: errcheck
		return
	}

	var err error
	var changes ServerChange

	switch event.Fields()[3] {
	case "invite":
		var invite *discordgo.Invite
		parts := regexp.DiscordInviteRegexp.FindStringSubmatch(values[0])
		if len(parts) >= 6 {
			invite, err = discord.Invite(p.redis, event.Discord(), parts[5])
			if err != nil {
				event.Except(err)
				return
			}
		} else {
			invite, err = discord.Invite(p.redis, event.Discord(), values[0])
			if err != nil {
				event.Except(err)
				return
			}
		}

		if invite == nil || invite.Guild == nil || invite.Code == "" || invite.Guild.ID == "" {
			event.Respond("serverlist.edit.invalid-invite") // nolint: errcheck
			return
		}

		if invite.Guild.ID != server.GuildID {
			event.Respond("serverlist.edit.invalid-invite-guild") // nolint: errcheck
			return
		}

		err = server.Update(p, Server{
			InviteCode: invite.Code,
		})
		if err != nil {
			event.Except(err)
			return
		}

		_, err = event.Respond("serverlist.edit.invite-success")
		event.Except(err)
		return

	case "name", "names":
		var names []string
		for _, value := range values {
			for _, newName := range strings.Split(value, ";") {
				names = append(names, strings.TrimSpace(newName))
			}
		}

		if len(names) == 0 {
			event.Respond("serverlist.edit.no-name") // nolint: errcheck
			return
		}

		changes.Names = names

	case "description":

		if len(values[0]) > descriptionCharacterLimit {
			event.Respond("serverlist.edit.description-too-long", // nolint: errcheck
				"limit", descriptionCharacterLimit,
			)
			return
		}

		changes.Description = values[0]

	case "category", "categories":

		allCategories, err := categoriesFindMany(p.db, "bot_id = ?", event.BotUserID)
		if err != nil {
			event.Except(err)
			return
		}

		var categoryIDs pq.Int64Array
		for _, value := range values {
			for _, categoryName := range strings.Split(value, ";") {
				categoryName = strings.ToLower(strings.TrimSpace(categoryName))

				for _, allCategory := range allCategories {
					for _, keyword := range allCategory.Keywords {
						if keyword != categoryName {
							continue
						}

						if int64SliceContains(int64(allCategory.ID), categoryIDs) {
							continue
						}

						categoryIDs = append(categoryIDs, int64(allCategory.ID))
					}
				}
			}

			result := state.ChannelRegex.FindStringSubmatch(value)
			if len(result) != 4 {
				continue
			}

			channel, err := p.state.Channel(result[2])
			if err != nil {
				continue
			}

			category, err := categoryFind(p.db, "channel_id = ?", channel.ID)
			if err != nil {
				category, err = categoryFind(p.db, "channel_id = ?", channel.ParentID)
				if err != nil {
					continue
				}
			}

			if int64SliceContains(int64(category.ID), categoryIDs) {
				continue
			}

			categoryIDs = append(categoryIDs, int64(category.ID))
		}

		if len(categoryIDs) == 0 {
			event.Respond("serverlist.edit.no-categories") // nolint: errcheck
			return
		}

		changes.Categories = categoryIDs

	case "editor":

		editorChange, err := p.state.UserFromMention(values[0])
		if err != nil {
			event.Except(err)
			return
		}

		var removed bool

		if stringSliceContains(editorChange.ID, server.EditorUserIDs) {
			if len(server.EditorUserIDs) <= 1 {
				event.Respond("serverlist.edit.remove-editors-too-few") // nolint: errcheck
				return
			}

			var newEditorUserIDs []string
			for _, editorUserID := range server.EditorUserIDs {
				if editorUserID == editorChange.ID {
					continue
				}

				newEditorUserIDs = append(newEditorUserIDs, editorUserID)
			}

			server.EditorUserIDs = newEditorUserIDs
			removed = true
		} else {
			server.EditorUserIDs = append(server.EditorUserIDs, editorChange.ID)
		}

		if len(server.EditorUserIDs) == 0 {
			event.Respond("serverlist.edit.remove-editors-too-few") // nolint: errcheck
			return
		}

		err = server.Update(p, Server{
			EditorUserIDs: server.EditorUserIDs,
		})
		if err != nil {
			event.Except(err)
			return
		}

		_, err = event.Respond("serverlist.edit.editors-success",
			"removed", removed, "editor", editorChange, "server", server,
		)
		event.Except(err)
		return

	default:

		if len(event.Fields()) < 4 {
			event.Respond("serverlist.edit.too-few-args-lt4") // nolint: errcheck
			return
		}

	}

	// TODO: how to edit editors?

	err = server.Edit(p, changes)
	if err != nil {
		event.Except(err)
		return
	}

	_, err = event.Respond("serverlist.edit.queued")
	event.Except(err)
}