package common

import (
	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/featureflag"
	"gitlab.com/Cacophony/go-kit/interfaces"
	"gitlab.com/Cacophony/go-kit/state"
	"go.uber.org/zap"
)

type StartParameters struct {
	Logger         *zap.Logger
	DB             *gorm.DB
	Redis          *redis.Client
	Tokens         map[string]string
	State          *state.State
	FeatureFlagger *featureflag.FeatureFlagger
	PluginHelpList []*PluginHelp
	Localizations  []interfaces.Localization
	Publisher      *events.Publisher
	HTTPMux        *chi.Mux
}

type StopParameters struct {
	Logger         *zap.Logger
	DB             *gorm.DB
	Redis          *redis.Client
	Tokens         map[string]string
	State          *state.State
	FeatureFlagger *featureflag.FeatureFlagger
	Localizations  []interfaces.Localization
	Publisher      *events.Publisher
}
