package serverlist

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/Cacophony/go-kit/discord"

	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
)

func (p *Plugin) getLogApprovedEmbed(server *Server) *discordgo.MessageEmbed {
	baseEmbed := p.getBaseServerEmbed(server, true)
	baseEmbed.Title = "serverlist.log.approved.embed.title"

	return baseEmbed
}

// func getLogRejectedEmbed(server *Server) *discordgo.MessageEmbed {
// 	baseEmbed := getBaseServerEmbed(server, false)
// 	baseEmbed.Title = "serverlist.log.rejected.embed.title"
//
// 	return baseEmbed
// }
//
// func getLogRemovedEmbed(server *Server) *discordgo.MessageEmbed {
// 	baseEmbed := getBaseServerEmbed(server, false)
// 	baseEmbed.Title = "serverlist.log.removed.embed.title"
//
// 	return baseEmbed
// }
//
// func getLogHiddenEmbed(server *Server) *discordgo.MessageEmbed {
// 	baseEmbed := getBaseServerEmbed(server, false)
// 	baseEmbed.Title = "serverlist.log.hidden.embed.title"
//
// 	return baseEmbed
// }
//
// func getLogUnhiddenEmbed(server *Server) *discordgo.MessageEmbed {
// 	baseEmbed := getBaseServerEmbed(server, true)
// 	baseEmbed.Title = "serverlist.log.unhidden.embed.title"
//
// 	return baseEmbed
// }

func (p *Plugin) getBaseServerEmbed(server *Server, invite bool) *discordgo.MessageEmbed {
	var categoryText string
	for _, category := range server.Categories {
		channel, err := p.state.Channel(category.Category.ChannelID)
		if err == nil && channel.ParentID != "" {
			categoryText += "<#" + channel.ParentID + "> / "
		}

		categoryText += "<#" + category.Category.ChannelID + ">, "
	}
	categoryText = strings.TrimRight(categoryText, ", ")

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "🏷 Name",
			Value:  strings.Join(server.Names, "; "),
			Inline: true,
		},
	}
	if invite {
		fields = append(fields,
			&discordgo.MessageEmbedField{
				Name:   "🚩 Invite",
				Value:  fmt.Sprintf("https://discord.gg/%s", server.InviteCode),
				Inline: true,
			},
		)
	}
	fields = append(fields,
		&discordgo.MessageEmbedField{
			Name:   "📖 Description",
			Value:  server.Description,
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "🗃 Category",
			Value:  categoryText,
			Inline: false,
		},
	)

	return &discordgo.MessageEmbed{
		Fields: fields,
	}
}

func (p *Plugin) getQueueMessageEmbed(session *discord.Session, server *Server, total int) *discordgo.MessageEmbed {
	if server == nil {
		return &discordgo.MessageEmbed{
			Title:       "⌛ Serverlist Queue",
			Description: "Queue empty!",
		}
	}

	titleText := "**⭐ New Server**"

	if server.Change.State != "" && server.Change.State != StateQueued {
		titleText = "**🔄 Server Update**"
	}

	var categoryText string
	for _, category := range server.Categories {
		channel, err := p.state.Channel(category.Category.ChannelID)
		if err == nil && channel.ParentID != "" {
			categoryText += "<#" + channel.ParentID + "> / "
		}

		categoryText += "<#" + category.Category.ChannelID + ">, "
	}
	categoryText = strings.TrimRight(categoryText, ", ")

	var nameChange, inviteChange, descriptionChange, categoryChange string
	if len(server.Change.Names) > 0 {
		nameChange = "\n➡\n" + strings.Join(server.Change.Names, "; ")
	}
	if len(server.Change.InviteCode) > 0 {
		inviteChange = "\n➡\n" + fmt.Sprintf("https://discord.gg/%s", server.Change.InviteCode)
	}
	if len(server.Change.Description) > 0 {
		descriptionChange = "\n➡\n" + server.Change.Description
	}
	if len(server.Change.Categories) > 0 {
		categoryChange = "\n➡\n"
		for _, categoryID := range server.Change.Categories {
			category, err := categoryFind(p.db, "id = ?", categoryID)
			if err != nil {
				continue
			}

			channel, err := p.state.Channel(category.ChannelID)
			if err == nil && channel.ParentID != "" {
				categoryChange += "<#" + channel.ParentID + "> / "
			}

			categoryChange += "<#" + category.ChannelID + ">, "
		}
		categoryChange = strings.TrimRight(categoryChange, ", ")
	}
	embed := &discordgo.MessageEmbed{
		Title:       "⌛ Serverlist Queue: " + titleText,
		Description: "serverlist.queue.embed.description",
		Timestamp:   server.CreatedAt.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf(
				"there are %d Servers queued in total • added", total,
			),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🏷 Name(s)",
				Value:  strings.Join(server.Names, "; ") + nameChange,
				Inline: true,
			},
			{
				Name: "🔢 Server ID",
				Value: fmt.Sprintf("`#%s`",
					server.GuildID,
				),
				Inline: true,
			},
			{
				Name: "👥 Editor(s)",
				Value: fmt.Sprintf("<@%s>",
					strings.Join(server.EditorUserIDs, "> <@"),
				),
				Inline: true,
			},
			{
				Name:   "🚩 Invite",
				Value:  fmt.Sprintf("https://discord.gg/%s", server.InviteCode) + inviteChange,
				Inline: true,
			},
			{
				Name:   "📈 Members",
				Value:  humanize.Comma(int64(server.TotalMembers)),
				Inline: true,
			},
			{
				Name:   "📖 Description",
				Value:  server.Description + descriptionChange,
				Inline: false,
			},
			{
				Name:   "🗃 Category",
				Value:  categoryText + categoryChange,
				Inline: false,
			},
		},
	}

	invite, err := discord.Invite(p.redis, session, server.InviteCode)
	if err == nil && invite != nil && invite.Code == server.InviteCode &&
		invite.Inviter != nil && invite.Inviter.ID != "" {

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🚪 Inviter",
			Value:  fmt.Sprintf("<@%s>", invite.Inviter.ID),
			Inline: true,
		})
	}

	guild, err := p.state.Guild(server.GuildID)
	if err == nil && guild != nil && guild.OwnerID != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "👑 Owner",
			Value:  fmt.Sprintf("<@%s>", guild.OwnerID),
			Inline: true,
		})
	}

	return embed
}