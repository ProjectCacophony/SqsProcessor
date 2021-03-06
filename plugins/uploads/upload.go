package uploads

import (
	"sync"

	"gitlab.com/Cacophony/go-kit/events"
)

func (p *Plugin) handleUpload(event *events.Event) {
	if len(event.MessageCreate.Attachments) <= 0 && len(event.Fields()) <= 0 {
		event.Respond("uploads.upload.too-few")
		return
	}

	messages, _ := event.Respond("common.uploading-file")
	var cleanup sync.Once
	cleanupFunc := func() {
		for _, message := range messages {
			event.Discord().Client.ChannelMessageDelete(event.ChannelID, message.ID)
		}
	}
	defer cleanup.Do(cleanupFunc)

	uploads := make([]Upload, 0, len(event.MessageCreate.Attachments))
	for _, attachment := range event.MessageCreate.Attachments {
		file, err := event.AddAttachement(attachment)
		if err != nil {
			event.Except(err)
			return
		}

		err = addUpload(p.db, file, event.UserID)
		if err != nil {
			event.Except(err)
			return
		}

		uploads = append(uploads, Upload{
			FileInfo: *file,
			UserID:   event.UserID,
		})
	}
	if len(event.Fields()) > 0 {
		for _, field := range append(event.Fields()[1:], event.OriginalCommand()) {
			if !xurlsStrict.MatchString(field) {
				continue
			}

			file, err := event.AddFileFromURL(field, "")
			if err != nil {
				event.Except(err)
				return
			}

			err = addUpload(p.db, file, event.UserID)
			if err != nil {
				event.Except(err)
				return
			}

			uploads = append(uploads, Upload{
				FileInfo: *file,
				UserID:   event.UserID,
			})
		}
	}

	if len(uploads) <= 0 {
		event.Respond("uploads.upload.too-few")
		return
	}

	cleanup.Do(cleanupFunc)

	_, err := event.Respond("uploads.upload.content", "uploads", uploads)
	event.Except(err)
}
