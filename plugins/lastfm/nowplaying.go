package lastfm

import (
	"strings"

	"github.com/Seklfreak/lastfm-go/lastfm"
	"github.com/bwmarrin/discordgo"
	lastfm_client "gitlab.com/Cacophony/Processor/plugins/lastfm/lastfm-client"
	"gitlab.com/Cacophony/go-kit/events"
)

func (p *Plugin) handleNowPlaying(event *events.Event, lastfmClient *lastfm.Api, offset int) {
	lastfmUsername := extractUsername(event, p.db, event.Fields(), offset)
	if lastfmUsername == "" {
		event.Respond("lastfm.no-user")
		return
	}

	// lookup user
	userInfo, err := lastfm_client.GetUserinfo(lastfmClient, lastfmUsername)
	if err != nil {
		if strings.Contains(err.Error(), "User not found") {
			event.Respond("lastfm.user-not-found", "username", lastfmUsername)
			return
		}
		event.Except(err)
		return
	}

	// get basic embed for user
	embed := getLastfmUserBaseEmbed(userInfo)

	// get recent tracks
	var tracks []lastfm_client.TrackData
	tracks, err = lastfm_client.GetRecentTracks(lastfmClient, userInfo.Username, 2)
	if err != nil {
		event.Except(err)
		return
	}

	// if no tracks found, post error and stop
	if len(tracks) < 1 {
		event.Respond("lastfm.no-scrobbles", "username", lastfmUsername)
		return
	}

	embed.Author.Name = "lastfm.np.embed.title"
	embed.Description = event.Translate("lastfm.track.long", "track", tracks[0])

	// add album information if possible
	if tracks[0].Album != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "lastfm.album.title",
			Value:  event.Translate("lastfm.album.value", "album", tracks[0].Album),
			Inline: false,
		})
	}

	// add image if possible
	if tracks[0].ImageURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: tracks[0].ImageURL,
		}
	}

	// add previous track if possible
	if len(tracks) >= 2 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "lastfm.np.embed.listened-to-before",
			Value:  event.Translate("lastfm.track.long", "track", tracks[1]),
			Inline: false,
		})
	}

	_, err = event.RespondComplex(
		&discordgo.MessageSend{
			Embed: embed,
		},
		"userData", userInfo, "tracks", tracks,
	)
	if err != nil {
		event.Except(err)
		return
	}
}
