package eventlog

import (
	"github.com/bwmarrin/discordgo"
	"gitlab.com/Cacophony/go-kit/config"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/permissions"
)

func (p *Plugin) handleEventlogUpdate(event *events.Event) {
	item, err := GetItem(event.DB(), event.EventlogUpdate.ItemID)
	if err != nil {
		event.ExceptSilent(err)
		return
	}

	channelID, err := config.GuildGetString(p.db, event.EventlogUpdate.GuildID, eventlogChannelKey)
	if err != nil {
		event.ExceptSilent(err)
		return
	}

	botID, err := event.State().BotForChannel(channelID, permissions.DiscordSendMessages, permissions.DiscordEmbedLinks)
	if err != nil {
		event.ExceptSilent(err)
		return
	}
	event.BotUserID = botID

	if item.LogMessage.MessageID != "" && item.LogMessage.ChannelID != "" {
		p.logger.Error("updating eventlog messages not supported")
		// TODO: update existing message
		return
	}

	embed := item.Embed(event.State())

	messages, err := event.SendComplex(
		channelID,
		&discordgo.MessageSend{Embed: embed},
	)
	if err != nil {
		event.ExceptSilent(err)
		return
	}

	err = saveItemMessage(event.DB(), item.ID, messages[0].ID, messages[0].ChannelID)
	if err != nil {
		event.ExceptSilent(err)
		return
	}
}
