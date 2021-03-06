package lastfm

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jinzhu/gorm"
	lastfm_client "gitlab.com/Cacophony/Processor/plugins/lastfm/lastfm-client"
	"gitlab.com/Cacophony/go-kit/discord"
	"gitlab.com/Cacophony/go-kit/events"
)

func extractUsername(event *events.Event, db *gorm.DB, args []string, pos int) string {
	var username string
	// try any mentions in the command
	if len(event.MessageCreate.Mentions) > 0 {
		username = getLastFmUsername(db, event.MessageCreate.Mentions[0].ID)

		if username != "" {
			return username
		}
	}
	// try field at pos
	if pos < 0 {
		pos = len(args) - 1
		if pos < 0 {
			pos = 0
		}
	}
	if len(args) > pos {
		username = args[pos]

		if username != "" {
			return username
		}
	}

	username = getLastFmUsername(db, event.MessageCreate.Author.ID)
	return username
}

// getLastfmUserBaseEmbed gets a discordgo embed base for a last.fm user
func getLastfmUserBaseEmbed(userInfo lastfm_client.UserData) *discordgo.MessageEmbed {
	var embed discordgo.MessageEmbed

	// set embed author
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL: "lastfm.embed.user.author.url",
	}
	if userInfo.Icon != "" {
		embed.Author.IconURL = userInfo.Icon
	}
	// set embed footer
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    "lastfm.embed.user.footer.text",
		IconURL: "lastfm.embed.footer.iconurl",
	}
	// set embed colour
	embed.Color = discord.HexToColorCode("#d51007")

	return &embed
}

// // getLastfmGuildBaseEmbed gets a discordgo embed base for a discord guild
// func getLastfmGuildBaseEmbed(guild *discordgo.Guild, listeners int) (embed discordgo.MessageEmbed) {
// 	// set embed author
// 	embed.Author = &discordgo.MessageEmbedAuthor{
// 		Name: guild.Name,
// 	}
// 	if guild.Icon != "" {
// 		embed.Author.IconURL = discordgo.EndpointGuildIcon(guild.ID, guild.Icon)
// 	}
// 	// set embed footer
// 	embed.Footer = &discordgo.MessageEmbedFooter{
// 		Text:    "powered by last.fm",
// 		IconURL: "https://i.imgur.com/p8wijg4.png",
// 	}
// 	if listeners > 0 {
// 		embed.Footer.Text += " | Total Listeners " + strconv.Itoa(listeners)
// 	}
// 	// set embed colour
// 	embed.Color = discord.HexToColorCode("#d51007")
//
// 	return embed
// }

// isCollageRequest returns true if the args contain an arg making it a collage request
func isCollageRequest(args []string) (bool, []string) {
	for i, arg := range args {
		arg = strings.ToLower(arg)

		if arg != "collage" {
			continue
		}

		return true, append(args[:i], args[i+1:]...)
	}
	return false, args

}
