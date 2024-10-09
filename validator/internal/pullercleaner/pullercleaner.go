// Package pullercleaner houses a pipeline over a message stream which attempts to generate and register new
// schemas and send the updated messages to a destination topic.
package pullercleaner

import (
	"context"

	"github.com/pkg/errors"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errtemplates"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/janitor"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry"
	"github.com/dataphos/lib-brokers/pkg/broker"
	"github.com/dataphos/lib-logger/logger"
)

// PullerCleaner models the pullercleaner process.
type PullerCleaner struct {
	Publisher        broker.Publisher
	TopicIDs         Topics
	CleanerRouter    janitor.CleanerRouter
	cleanerRouterSem chan struct{}
	topics           map[string]broker.Topic
	log              logger.Log
}

// Topics defines the standard destination topics, based on possible validation results.
type Topics struct {
	Valid      string
	Deadletter string
}

// RouterFlags defines logging levels for logging each routing decision.
type RouterFlags struct {
	Valid      bool
	Deadletter bool
}

const (
	csvFormat  = "csv"
	jsonFormat = "json"
)

// New returns a new instance of PullerCleaner.
func New(schemaGenerators janitor.SchemaGenerators, registry registry.SchemaRegistry, validators janitor.Validators, publisher broker.Publisher, topicIds Topics, numCleaners int, log logger.Log, routerFlags RouterFlags) (*PullerCleaner, error) {
	topics, err := setupTopics(topicIds, publisher)
	if err != nil {
		return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.Deadletter))
	}

	return &PullerCleaner{
		Publisher: publisher,
		TopicIDs:  topicIds,
		CleanerRouter: janitor.CleanerRouter{
			Cleaner: janitor.NewCachingCleaner(schemaGenerators, validators, registry),
			Router:  setupRoutingFunc(topicIds, routerFlags, log),
		},
		cleanerRouterSem: make(chan struct{}, numCleaners),
		topics:           topics,
		log:              log,
	}, nil
}

// setupTopics maps Topics into instances of broker.Topic.
func setupTopics(topicIds Topics, publisher broker.Publisher) (map[string]broker.Topic, error) {
	topics := make(map[string]broker.Topic)

	if topicIds.Valid != "" {
		topic, err := publisher.Topic(topicIds.Valid)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.Valid))
		}
		topics[topicIds.Valid] = topic
	}

	if topicIds.Deadletter != "" {
		topic, err := publisher.Topic(topicIds.Deadletter)
		if err != nil {
			return nil, errors.Wrap(err, errtemplates.CreatingTopicInstanceFailed(topicIds.Deadletter))
		}
		topics[topicIds.Deadletter] = topic
	}
	return topics, nil
}

// setupRoutingFunc sets up the janitor.LoggingRouter, by first checking if there's a need for logging any of the routing
// decisions (if any logging level flag is set). If none of the flags are set, standard IntoRoutingFunc is used,
// wrapping it with logging middleware otherwise.
func setupRoutingFunc(topicIDs Topics, routerFlags RouterFlags, log logger.Log) janitor.Router {
	next := IntoRoutingFunc(topicIDs)

	if routerFlags.Valid || routerFlags.Deadletter {
		return janitor.LoggingRouter(
			log,
			janitor.RouterFlags{
				Valid:      routerFlags.Valid,
				Deadletter: routerFlags.Deadletter,
			},
			next,
		)
	}

	return next
}

// IntoRoutingFunc maps the given Topics into a janitor.LoggingRouter.
//
// If the janitor.Result is janitor.Deadletter, the message are routed to Topics.Deadletter.
// All valid messages are sent to Topics.Valid.
func IntoRoutingFunc(topics Topics) janitor.Router {
	return janitor.RoutingFunc(func(result janitor.Result, _ janitor.Message) string {
		switch result {
		case janitor.Valid:
			return topics.Valid
		case janitor.Invalid, janitor.Deadletter, janitor.MissingSchema:
			return topics.Deadletter
		default:
			return topics.Deadletter
		}
	})
}

func (pc *PullerCleaner) AsProcessor() *janitor.Processor {
	return janitor.NewProcessor(pc, pc.topics, pc.TopicIDs.Deadletter, pc.log)
}

func (pc *PullerCleaner) Handle(ctx context.Context, message janitor.Message) (janitor.MessageTopicPair, error) {
	acquireIfSet(pc.cleanerRouterSem)
	messageTopicPair, err := pc.CleanerRouter.CleanAndReroute(ctx, message)
	if err != nil {
		releaseIfSet(pc.cleanerRouterSem)
		return janitor.MessageTopicPair{}, err
	}
	releaseIfSet(pc.cleanerRouterSem)

	return messageTopicPair, nil
}

func acquireIfSet(sem chan struct{}) {
	if sem != nil {
		sem <- struct{}{}
	}
}

func releaseIfSet(sem chan struct{}) {
	if sem != nil {
		<-sem
	}
}
