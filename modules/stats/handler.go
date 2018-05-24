package stats

import (
	"context"

	"gitlab.com/Cacophony/dhelpers"
)

// Module is a struct for the module
type Module struct{}

// GetDestinations defines valid destinations for the module
func (m *Module) GetDestinations() []string {
	return []string{
		"stats",
	}
}

// GetTranslationFiles defines all translation files for the module
func (m *Module) GetTranslationFiles() []string {
	return []string{
		//"stats.en.toml",
	}
}

// Action is the module entry point when event is triggered
func (m *Module) Action(ctx context.Context) {
	event := dhelpers.EventFromContext(ctx)

	switch event.Type {
	case dhelpers.MessageCreateEventType:

		for _, destination := range event.Destinations {

			switch destination.Name {
			case "stats":
				displayStats(ctx)
				return
			}
		}
	}
}

// Init is called on bot startup
func (m *Module) Init() {
}

// Uninit is called on normal bot shutdown
func (m *Module) Uninit() {
}
