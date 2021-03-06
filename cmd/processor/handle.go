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
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.uber.org/zap"
)

var b3Prop = b3.B3{}

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
	publisher *events.Publisher,
) func(event *events.Event) error {
	l := logger.With(zap.String("service", "processor"))

	return func(event *events.Event) error { // nolint: unparam
		var err error

		ctx, cancel := context.WithTimeout(context.Background(), processingDeadline)
		defer cancel()

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
		event.WithPublisher(publisher)

		event.Parse()

		ctx, span := global.Tracer("cacophony.dev/processor").Start(
			b3Prop.Extract(ctx, &event.SpanContext),
			"handle.Event",
			trace.WithAttributes(
				events.SpanLabelEventingType.String(string(event.Type)),
				events.SpanLabelEventingIsCommand.Bool(event.Command()),
				events.SpanLabelDiscordBotUserID.String(event.BotUserID),
				events.SpanLabelDiscordGuildID.String(event.GuildID),
				events.SpanLabelDiscordChannelID.String(event.ChannelID),
				events.SpanLabelDiscordUserID.String(event.UserID),
				events.SpanLabelDiscordMessageID.String(event.MessageID),
			),
		)
		defer span.End()
		event.WithContext(ctx)

		if event.Command() {
			span.SetAttributes(events.SpanLabelEventingCommand.String(event.Fields()[0]))
		}

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

		ctx = event.Context()
		for _, plugin := range plugins.PluginList {
			pluginContext, span := global.Tracer("cacophony.dev/processor").Start(ctx, "Plugin."+plugin.Names()[0])
			event.WithContext(pluginContext)

			if !event.IsEnabled(featureFlagPluginKey(plugin.Names()[0]), true) {
				l.Debug("skipping plugin as it is disabled by feature flags",
					zap.String("plugin_name", plugin.Names()[0]),
					zap.String("user_id", event.UserID),
					zap.String("event_id", event.ID),
				)
				span.End()
				continue
			}

			event.WithLogger(l.With(zap.String("plugin", plugin.Names()[0])))

			if plugin.Passthrough() {
				// if passthrough, continue with next plugin asap

				wg.Add(1)

				go func(pl plugins.Plugin) {
					defer wg.Done()

					// TODO: figure out span, linked span?

					executePlugin(l, pl, event)
				}(plugin)
				span.End()
				continue
			}

			// else, wait for result, exit if handled
			handled = executePlugin(l, plugin, event)
			if handled {
				span.End()
				return nil
			}
			span.End()
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
				zap.String("plugin", plugin.Names()[0]),
				zap.String("event_type", string(event.Type)),
				zap.Any("error", err),
			)
		}
	}()

	// check if help command and redirect to help plugin
	if len(event.Fields()) > 1 && event.Fields()[1] == "help" {
		for _, p := range plugins.PluginList {
			if p.Names()[0] == "help" {
				event.Fields()[1] = event.Fields()[0]
				event.Fields()[0] = p.Names()[0]
				return p.Action(event)
			}
		}
	}

	return plugin.Action(event)
}

func featureFlagPluginKey(pluginName string) string {
	return "plugin-" + pluginName
}
