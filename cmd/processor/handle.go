package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"gitlab.com/Cacophony/Processor/plugins"
	"gitlab.com/Cacophony/go-kit/events"
	"gitlab.com/Cacophony/go-kit/featureflag"
	"gitlab.com/Cacophony/go-kit/paginator"
	"gitlab.com/Cacophony/go-kit/state"
	"go.uber.org/zap"
)

func handle(
	logger *zap.Logger,
	db *gorm.DB,
	stateClient *state.State,
	discordTokens map[string]string,
	featureFlagger *featureflag.FeatureFlagger,
	redis *redis.Client,
	paginator *paginator.Paginator,
	httpClient *http.Client,
	processingDeadline time.Duration,
	questionnaire *events.Questionnaire,
	storage *events.Storage,
) func(event *events.Event) error {
	l := logger.With(zap.String("service", "processor"))

	return func(event *events.Event) error { // nolint: unparam
		var err error

		ctx, cancel := context.WithTimeout(context.Background(), processingDeadline)
		defer cancel()

		event.WithContext(ctx)
		event.WithTokens(discordTokens)
		event.WithLocalizations(plugins.LocalizationsList)
		event.WithState(stateClient)
		event.WithPaginator(paginator)
		event.WithLogger(l)
		event.WithStorage(storage)
		event.WithRedis(redis)
		event.WithDB(db)
		event.WithHTTPClient(httpClient)
		event.WithQuestionnaire(questionnaire)
		event.WithFeatureFlagger(featureFlagger)

		event.Parse()

		switch event.Type {
		case events.MessageCreateType:
			err = paginator.HandleMessageCreate(event.MessageCreate)
			if err != nil {
				event.ExceptSilent(err)
			}
		case events.MessageReactionAddType:
			err = paginator.HandleMessageReactionAdd(event.MessageReactionAdd)
			if err != nil {
				event.ExceptSilent(err)
			}
		}

		var wg sync.WaitGroup
		var handled bool

		handled, err = questionnaire.Do(event.Context(), event)
		if err != nil {
			return errors.Wrap(err, "questionnaire unable to handle event")
		}
		if handled {
			return nil
		}

		for _, plugin := range plugins.PluginList {
			if !event.IsEnabled(featureFlagPluginKey(plugin.Name()), true) {
				l.Debug("skipping plugin as it is disabled by feature flags",
					zap.String("plugin_name", plugin.Name()),
					zap.String("user_id", event.UserID),
					zap.String("event_id", event.ID),
				)
				continue
			}

			event.WithLogger(l.With(zap.String("plugin", plugin.Name())))

			if plugin.Passthrough() {
				// if passthrough, continue with next plugin asap

				wg.Add(1)

				go func(pl plugins.Plugin) {
					defer wg.Done()

					executePlugin(l, pl, event)
				}(plugin)
				continue
			}

			// else, wait for result, exit if handled
			handled = executePlugin(l, plugin, event)
			if handled {
				return nil
			}
		}

		wg.Wait()

		return nil
	}
}

func executePlugin(logger *zap.Logger, plugin plugins.Plugin, event *events.Event) bool {
	defer func() {
		err := recover()
		if err != nil {
			if _, ok := err.(error); ok {
				event.ExceptSilent(err.(error))
			}

			logger.Error("plugin failed to handle event",
				zap.String("plugin", plugin.Name()),
				zap.String("event_type", string(event.Type)),
				zap.Any("error", err),
			)
		}
	}()

	// check if help command and redirect to help plugin
	if len(event.Fields()) > 1 && event.Fields()[1] == "help" {
		for _, p := range plugins.PluginList {
			if p.Name() == "help" {
				event.Fields()[1] = event.Fields()[0]
				event.Fields()[0] = p.Name()
				return p.Action(event)
			}
		}
	}

	return plugin.Action(event)
}

func featureFlagPluginKey(pluginName string) string {
	return "plugin-" + pluginName
}
