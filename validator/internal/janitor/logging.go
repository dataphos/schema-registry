package janitor

import (
	"github.com/pkg/errors"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/errcodes"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-streamproc/pkg/streamproc"
)

// RouterFlags defines the logging level of each call to the LoggingRouter.
//
// Intended to be used in LoggingRouter.
type RouterFlags struct {
	MissingSchema bool
	Valid         bool
	Invalid       bool
	Deadletter    bool
}

// LoggingRouter wraps the given LoggingRouter with logging middleware.
func LoggingRouter(log logger.Log, routerFlags RouterFlags, next Router) Router {
	return RoutingFunc(func(result Result, message Message) string {
		switch {
		case routerFlags.MissingSchema && result == MissingSchema:
			log.Warnw("message is missing the schema", logger.F{
				"status": "missing schema",
				"id":     message.ID,
				"format": message.Format,
			})
		case routerFlags.Valid && result == Valid:
			log.Infow("message is classified as valid", logger.F{
				"status":         "valid",
				"id":             message.ID,
				"schema_id":      message.SchemaID,
				"schema_version": message.Version,
				"format":         message.Format,
			})
		case routerFlags.Invalid && result == Invalid:
			log.Errorw("message is classified as invalid", errcodes.InvalidMessage, logger.F{
				"status":         "invalid",
				"id":             message.ID,
				"schema_id":      message.SchemaID,
				"schema_version": message.Version,
				"format":         message.Format,
			})
		case routerFlags.Deadletter && result == Deadletter:
			log.Errorw("message is classified as Deadletter", errcodes.DeadletterMessage, logger.F{
				"status":         "Deadletter",
				"id":             message.ID,
				"schema_id":      message.SchemaID,
				"schema_version": message.Version,
				"format":         message.Format,
			})
		}

		return next.Route(result, message)
	})
}

type ShouldReturnFlowControl struct {
	OnPullErr          streamproc.FlowControl
	OnProcessErr       streamproc.FlowControl
	OnUnrecoverable    streamproc.FlowControl
	OnThresholdReached streamproc.FlowControl
}

// LoggingCallbacks returns a slice of streamproc.RunOptions, configuring streamproc.RunOptions to log all events with the agreed error codes.
func LoggingCallbacks(log logger.Log, control ShouldReturnFlowControl) []streamproc.RunOption {
	onPullErr := func(err error) streamproc.FlowControl {
		log.Error(err.Error(), errcodes.PullingFailure)
		return control.OnPullErr
	}

	OnProcessErr := func(err error) streamproc.FlowControl {
		code := errcodes.Miscellaneous
		opError := &OpError{}
		if errors.As(err, &opError) {
			code = opError.Code
		}
		log.Error(err.Error(), uint64(code))

		return control.OnProcessErr
	}

	onUnrecoverable := func(err error) streamproc.FlowControl {
		log.Error(errors.Wrap(err, "unrecoverable error encountered").Error(), errcodes.UnrecoverableErrorEncountered)
		return control.OnUnrecoverable
	}

	onThresholdReached := func(err error, count, threshold int64) streamproc.FlowControl {
		log.Error(errors.Errorf("error threshold reached (%d >= %d)", count, threshold).Error(), errcodes.ErrorThresholdReached)
		return control.OnThresholdReached
	}

	return []streamproc.RunOption{
		streamproc.OnPullErr(onPullErr),
		streamproc.OnProcessErr(OnProcessErr),
		streamproc.OnUnrecoverable(onUnrecoverable),
		streamproc.OnThresholdReached(onThresholdReached),
	}
}
