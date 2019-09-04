package help

import (
	"gitlab.com/Cacophony/Processor/plugins/common"
	"gitlab.com/Cacophony/go-kit/events"
	"go.uber.org/zap"
)

type Plugin struct {
	logger         *zap.Logger
	pluginHelpList []*common.PluginHelp
}

func (p *Plugin) Names() []string {
	return []string{"help"}
}

func (p *Plugin) Start(params common.StartParameters) error {
	p.logger = params.Logger
	p.pluginHelpList = params.PluginHelpList

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
		Description: "help.help.description",
		Commands: []common.Command{{
			Name: "List All Modules",
			Params: []common.CommandParam{
				{Name: "public", Type: common.Flag, Optional: true},
			},
		}, {
			Name: "List Module Commands",
			Params: []common.CommandParam{
				{Name: "module name", Type: common.Text},
				{Name: "public", Type: common.Flag, Optional: true},
			},
		}},
	}
}

func (p *Plugin) Action(event *events.Event) bool {
	if !event.Command() {
		return false
	}

	if len(event.Fields()) < 1 || event.Fields()[0] != "help" {
		return false
	}

	var displayInChannel bool

	for _, field := range event.Fields() {
		if field == "public" {
			displayInChannel = true
		}
	}

	if len(event.Fields()) == 1 {
		listCommands(event, p.pluginHelpList, displayInChannel)
		return true
	}

	if len(event.Fields()) > 1 {
		if event.Fields()[1] == "public" {
			listCommands(event, p.pluginHelpList, displayInChannel)
			return true
		}

		// check if second param is a plugin name
		for _, help := range p.pluginHelpList {
			for _, name := range help.Names {
				if name == event.Fields()[1] {
					displayPluginCommands(event, help, displayInChannel)
					return true
				}
			}
		}
	}

	return false
}
