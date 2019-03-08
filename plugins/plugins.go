package plugins

import (
	"sort"

	"gitlab.com/Cacophony/Processor/plugins/chatlog"

	"gitlab.com/Cacophony/Processor/plugins/rss"

	"gitlab.com/Cacophony/Processor/plugins/gall"

	"gitlab.com/Cacophony/Processor/plugins/whitelist"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"gitlab.com/Cacophony/Processor/plugins/automod"
	"gitlab.com/Cacophony/Processor/plugins/color"
	"gitlab.com/Cacophony/Processor/plugins/common"
	"gitlab.com/Cacophony/Processor/plugins/dev"
	"gitlab.com/Cacophony/Processor/plugins/lastfm"
	"gitlab.com/Cacophony/Processor/plugins/ping"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/featureflag"
	"gitlab.com/Cacophony/go-kit/interfaces"
	"gitlab.com/Cacophony/go-kit/state"
	"go.uber.org/zap"
)

type Plugin interface {
	Name() string

	// TODO: add context for deadline
	Start(common.StartParameters) error

	// TODO: add context for deadline
	Stop(common.StopParameters) error

	Priority() int

	Passthrough() bool

	Localisations() []interfaces.Localisation

	Action(event *events.Event) bool
}

// nolint: gochecknoglobals
var (
	PluginList = []Plugin{
		&ping.Plugin{},
		&color.Plugin{},
		&dev.Plugin{},
		&lastfm.Plugin{},
		&automod.Plugin{},
		&whitelist.Plugin{},
		&gall.Plugin{},
		&rss.Plugin{},
		&chatlog.Plugin{},
	}

	LocalisationsList []interfaces.Localisation
)

// nolint: gochecknoinits
func init() {
	// sort plugins by priority
	sort.Sort(ByPriority(PluginList))
}

func StartPlugins(
	logger *zap.Logger,
	db *gorm.DB,
	redis *redis.Client,
	tokens map[string]string,
	state *state.State,
	featureFlagger *featureflag.FeatureFlagger,
) {
	var err error
	for _, plugin := range PluginList {
		err = plugin.Start(common.StartParameters{
			Logger:         logger,
			DB:             db,
			Redis:          redis,
			Tokens:         tokens,
			State:          state,
			FeatureFlagger: featureFlagger,
		})
		if err != nil {
			logger.Error("failed to start plugin",
				zap.Error(err),
			)
		}
		// TODO: do not send events to plugins that failed to start

		LocalisationsList = append(LocalisationsList, plugin.Localisations()...)
	}
}

func StopPlugins(
	logger *zap.Logger,
	db *gorm.DB,
	redis *redis.Client,
	tokens map[string]string,
	state *state.State,
	featureFlagger *featureflag.FeatureFlagger,
) {
	var err error
	for _, plugin := range PluginList {
		err = plugin.Stop(common.StopParameters{
			Logger:         logger,
			DB:             db,
			Redis:          redis,
			Tokens:         tokens,
			State:          state,
			FeatureFlagger: featureFlagger,
		})
		if err != nil {
			logger.Error("failed to stop plugin",
				zap.Error(err),
			)
		}
	}
}

type ByPriority []Plugin

func (p ByPriority) Len() int           { return len(p) }
func (p ByPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByPriority) Less(i, j int) bool { return p[i].Priority() > p[j].Priority() }
