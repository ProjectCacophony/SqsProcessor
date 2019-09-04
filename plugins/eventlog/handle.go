package eventlog

import (
	"strings"

	"github.com/jinzhu/gorm"

	"gitlab.com/Cacophony/Processor/plugins/common"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/interfaces"
	"gitlab.com/Cacophony/go-kit/permissions"
	"gitlab.com/Cacophony/go-kit/state"
	"go.uber.org/zap"
)

type Plugin struct {
	logger *zap.Logger
	state  *state.State
	db     *gorm.DB
}

func (p *Plugin) Names() []string {
	return []string{"eventlog"}
}

func (p *Plugin) Start(params common.StartParameters) error {
	p.logger = params.Logger
	p.state = params.State
	p.db = params.DB
	return nil
}

func (p *Plugin) Stop(params common.StopParameters) error {
	return nil
}

func (p *Plugin) Priority() int {
	return 0
}

func (p *Plugin) Passthrough() bool {
	return false
}

func (p *Plugin) Help() *common.PluginHelp {
	return &common.PluginHelp{
		Names:       p.Names(),
		Description: "eventlog.help.description",
		PermissionsRequired: []interfaces.Permission{
			permissions.BotAdmin,
			permissions.Patron,
		},
		Commands: []common.Command{
			{
				Name:        "eventlog.help.status.name",
				Description: "eventlog.help.status.description",
			},
			{
				Name:        "eventlog.help.enable.name",
				Description: "eventlog.help.enable.description",
				Params: []common.CommandParam{
					{Name: "enable", Type: common.Flag},
				},
			},
			{
				Name:        "eventlog.help.disable.name",
				Description: "eventlog.help.disable.description",
				Params: []common.CommandParam{
					{Name: "disable", Type: common.Flag},
				},
			},
		},
	}
}

func (p *Plugin) Action(event *events.Event) bool {
	if !event.Command() {
		return false
	}

	if event.Fields()[0] != "eventlog" {
		return false
	}

	// has to be Server Admin, and Patron/Staff
	event.Require(
		func() {
			p.handleCommand(event)
		},
		permissions.DiscordAdministrator,
		permissions.BotAdmin,
		// TODO
		// permissions.Or(
		// 	permissions.BotAdmin,
		// 	permissions.Patron,
		// ),
		permissions.Not(
			permissions.DiscordChannelDM,
		),
	)

	return true
}

func (p *Plugin) handleCommand(event *events.Event) {

	if len(event.Fields()) >= 2 {
		switch strings.ToLower(event.Fields()[1]) {

		case "enable":

			p.handleEnable(event)
			return
		case "disable":

			p.handleDisable(event)
			return
		}
	}

	p.handleStatus(event)
}
