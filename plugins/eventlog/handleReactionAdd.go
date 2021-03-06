package eventlog

import (
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/permissions"
)

func (p *Plugin) handleReactionAdd(event *events.Event) bool {
	user, err := event.State().User(event.UserID)
	if err != nil {
		event.ExceptSilent(err)
		return false
	}
	if user.Bot {
		return false
	}

	var item *Item
	switch event.MessageReactionAdd.Emoji.Name {
	case reactionEditReason,
		reactionRevert:
		item, err = FindItem(event.DB(),
			"log_message_channel_id = ? AND log_message_message_id = ?",
			event.MessageReactionAdd.ChannelID, event.MessageReactionAdd.MessageID,
		)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			event.ExceptSilent(err)
			return false
		}
	default:
		return false
	}

	if item == nil {
		return false
	}

	switch event.MessageReactionAdd.Emoji.Name {
	case reactionEditReason:
		p.handleReactionEditReason(event, item)
		return true
	case reactionRevert:
		p.handleReactionRevert(event, item)
		return true
	}

	return false
}

func (p *Plugin) handleReactionEditReason(event *events.Event, item *Item) {
	defer func() {
		// remove reaction if possible
		if permissions.DiscordManageMessages.Match(
			event.State(),
			p.db,
			event.BotUserID,
			event.MessageReactionAdd.ChannelID,
			false,
			false,
		) {
			err := event.Discord().Client.MessageReactionRemove(
				event.MessageReactionAdd.ChannelID,
				event.MessageReactionAdd.MessageID,
				event.MessageReactionAdd.Emoji.APIName(),
				event.MessageReactionAdd.UserID,
			)
			if err != nil {
				event.ExceptSilent(err)
				return
			}
		}
	}()

	messages, err := event.Send(
		event.ChannelID,
		"eventlog.edit-reason.ask",
		"userID",
		event.MessageReactionAdd.UserID,
		"item",
		item,
	)
	if err != nil {
		event.ExceptSilent(err)
		return
	}
	if len(messages) <= 0 {
		return
	}

	err = event.Questionnaire().Register(
		questionnaireEditReasonKey,
		events.QuestionnaireFilter{
			GuildID:   event.GuildID,
			ChannelID: event.ChannelID,
			UserID:    event.MessageReactionAdd.UserID,
			Type:      events.MessageCreateType,
		},
		map[string]interface{}{
			"item_id":             item.ID,
			"question_channel_id": messages[0].ChannelID,
			"question_message_id": messages[0].ID,
		},
	)
	if err != nil {
		event.ExceptSilent(err)
	}
}

func (p *Plugin) handleQuestionnaireEditReason(event *events.Event) {
	cleanup := func() {
		// delete questionnaire response message if possible
		if permissions.DiscordManageMessages.Match(
			event.State(),
			p.db,
			event.BotUserID,
			event.ChannelID,
			false,
			false,
		) {
			event.Discord().Client.ChannelMessageDelete(
				event.ChannelID,
				event.MessageID,
			)
		}

		// delete questionnaire question message
		event.Discord().Client.ChannelMessageDelete(
			event.QuestionnaireMatch.Payload["question_channel_id"].(string),
			event.QuestionnaireMatch.Payload["question_message_id"].(string),
		)
	}

	itemID, ok := event.QuestionnaireMatch.Payload["item_id"].(float64)
	if !ok {
		event.ExceptSilent(errors.New("received unexpected item ID"))
		return
	}

	if event.MessageCreate.Content == "" ||
		strings.ToLower(event.MessageCreate.Content) == "cancel" {
		cleanup()
		return
	}

	err := CreateOptionForItem(event.DB(), event.Publisher(), uint(itemID), event.GuildID, &ItemOption{
		AuthorID: event.UserID,
		Key:      "reason",
		NewValue: event.MessageCreate.Content,
		Type:     EntityTypeText,
	})
	if err != nil {
		event.ExceptSilent(err)
		return
	}

	go func() {
		time.Sleep(1 * time.Second)

		cleanup()
	}()
}

func (p *Plugin) handleReactionRevert(event *events.Event, item *Item) {
	defer func() {
		// remove reaction if possible
		if permissions.DiscordManageMessages.Match(
			event.State(),
			p.db,
			event.BotUserID,
			event.MessageReactionAdd.ChannelID,
			false,
			false,
		) {
			err := event.Discord().Client.MessageReactionRemove(
				event.MessageReactionAdd.ChannelID,
				event.MessageReactionAdd.MessageID,
				event.MessageReactionAdd.Emoji.APIName(),
				event.MessageReactionAdd.UserID,
			)
			if err != nil {
				event.ExceptSilent(err)
				return
			}
		}
	}()

	if !item.ActionType.Revertable() {
		return
	}

	err := item.Revert(event)
	if err != nil {
		event.Send(event.ChannelID, "eventlog.revert.error", "userID", event.UserID, "item", item, "error", err)
		event.ExceptSilent(err)
		return
	}

	event.Discord().Client.MessageReactionRemove(
		event.MessageReactionAdd.ChannelID,
		event.MessageReactionAdd.MessageID,
		event.MessageReactionAdd.Emoji.APIName(),
		event.BotUserID,
	)
}
