package instagram

import (
	"strings"

	"gitlab.com/Cacophony/go-kit/events"
)

func (p *Plugin) add(event *events.Event) {
	if len(event.Fields()) < 2 {
		event.Respond("instagram.add.too-few")
		return
	}

	fields := event.Fields()[2:]

	channel, fields, err := paramsExtractChannel(event, fields)
	if err != nil {
		event.Except(err)
		return
	}
	if event.DM() {
		channel.ID = event.UserID
	}

	if len(fields) < 1 {
		event.Respond("instagram.add.too-few")
		return
	}

	entries, err := entryFindMany(p.db,
		"((guild_id = ? AND dm = false) OR (channel_or_user_id = ? AND dm = true)) AND dm = ?",
		event.GuildID, event.UserID, event.DM(),
	)
	if err != nil {
		event.Except(err)
		return
	}
	if len(entries) >= feedsPerGuildLimit(event) &&
		feedsPerGuildLimit(event) >= 0 {
		event.Respond("instagram.add.too-many")
		return
	}

	instagramUsername := extractInstagramUsername(fields[0])

	for _, entry := range entries {
		if entry.ChannelOrUserID != channel.ID {
			continue
		}
		if !strings.EqualFold(entry.InstagramUsername, instagramUsername) {
			continue
		}

		event.Respond("instagram.add.duplicate")
		return
	}

	err = entryAdd(
		p.db,
		event.UserID,
		channel.ID,
		event.GuildID,
		instagramUsername,
		event.BotUserID,
		event.DM(),
	)
	if err != nil {
		event.Except(err)
		return
	}

	_, err = event.Respond("instagram.add.success",
		"instagramUsername", instagramUsername,
		"channel", channel,
		"dm", event.DM(),
	)
	event.Except(err)
}
