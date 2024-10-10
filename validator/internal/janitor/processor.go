// Copyright 2024 Syntio Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package janitor

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/dataphos/schema-registry-validator/internal/errcodes"
	"github.com/dataphos/lib-batchproc/pkg/batchproc"
	"github.com/dataphos/lib-brokers/pkg/broker"
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-streamproc/pkg/streamproc"
)

type Processor struct {
	Handler    Handler
	Topics     map[string]broker.Topic
	Deadletter string
	log        logger.Log
}

type Handler interface {
	Handle(context.Context, Message) (MessageTopicPair, error)
}

func NewProcessor(handler Handler, topics map[string]broker.Topic, deadletter string, log logger.Log) *Processor {
	return &Processor{
		Handler:    handler,
		Topics:     topics,
		Deadletter: deadletter,
		log:        log,
	}
}

// HandleMessage processes the given streamproc.Message by first attempting to parse it, and then calling the
// underlying Handler.
func (p *Processor) HandleMessage(ctx context.Context, message streamproc.Message) error {
	// The received ctx isn't used because the whole process is assumed to take up very little time.
	// Because of this, it's preferred to exit "cleanly" instead of stopping mid-process, which might
	// have side effects.
	ctx = context.Background() //nolint:staticcheck (ignored the rule SA4009 in .golangci.yaml)

	parsed, ok, err := p.parseOrSendToDeadletter(ctx, message)
	if err != nil {
		UpdateFailureMetrics(parsed)
		return err
	}
	if ok {
		messageTopicPair, err := p.Handler.Handle(ctx, parsed)
		if err != nil {
			UpdateFailureMetrics(parsed)
			return err
		}
		if err = PublishToTopic(ctx, messageTopicPair.Message, p.Topics[messageTopicPair.Topic]); err != nil {
			UpdateFailureMetrics(messageTopicPair.Message)
			return err
		}
		// if message is invalid (sent to DL), update DL metrics
		if messageTopicPair.Topic == p.Deadletter {
			UpdateSuccessDLMetrics(parsed)
		}
	}
	UpdateSuccessMetrics(parsed)
	return nil
}

func (p *Processor) parseOrSendToDeadletter(ctx context.Context, message streamproc.Message) (Message, bool, error) {
	parsed, err := ParseMessage(message)
	if err != nil {
		p.log.Errorw(err.Error(), errcodes.ParsingMessage, logger.F{
			"id": message.ID,
		})
		parsed.RawAttributes["deadLetterErrorCategory"] = "Parsing error"
		parsed.RawAttributes["deadLetterErrorReason"] = err.Error()
		if err = PublishToTopic(ctx, parsed, p.Topics[p.Deadletter]); err != nil {
			return Message{}, false, err
		}
		UpdateSuccessDLMetrics(parsed)
		return Message{}, false, nil
	}
	return parsed, true, nil
}

// HandleBatch processes the given slice of streamproc.Message instances, by calling the underlying Handler,
// on each streamproc.Message concurrently.
func (p *Processor) HandleBatch(ctx context.Context, batch []streamproc.Message) error {
	// The received ctx isn't used because the whole process is assumed to take up very little time.
	// Because of this, it's preferred to exit "cleanly" instead of stopping mid-process, which might
	// have side effects.
	ctx = context.Background() //nolint:staticcheck

	batchSize := len(batch)
	messageTopicPairs := make([]*MessageTopicPair, batchSize)
	failed := make([]bool, batchSize)

	// The batch is processed in chunks (and not one per goroutine), since the Handler is assumed to be mostly CPU bound.
	if err := batchproc.Parallel(ctx, batchSize, func(ctx context.Context, lb int, ub int) error {
		for i := lb; i < ub; i++ {
			parsed, ok, err := p.parseOrSendToDeadletter(ctx, batch[i])
			if err != nil {
				UpdateFailureMetrics(parsed)
				failed[i] = true
				return err
			}
			if ok {
				messageTopicPair, err := p.Handler.Handle(ctx, parsed)
				if err != nil {
					UpdateFailureMetrics(parsed)
					failed[i] = true
					return err
				}
				messageTopicPairs[i] = &messageTopicPair
			}
		}
		return nil
	}); err != nil {
		return &streamproc.PartiallyProcessedBatchError{
			Failed: indicesWhereTrue(failed),
			Err:    err,
		}
	}

	// Publish order needs to be preserved if messages have the same key, so we partition the processed messages before publishing them.
	// That way, we can still utilize concurrency in the general case (it's unlikely for multiple messages in a batch to actually share the key),
	// but still ensure ordering on the target topics when it matters.
	partitions := groupByKey(messageTopicPairs)
	// The empty string key implies no key is defined, so we don't care about order for this partition.
	keyless := partitions[""]
	delete(partitions, "")

	eg, ctx := errgroup.WithContext(ctx)

	// Required extra check for the number of keyless messages to avoid division by zero in batchproc.Process().
	if numKeyless := len(keyless); numKeyless != 0 {
		eg.Go(func() error {
			return batchproc.Process(ctx, numKeyless, numKeyless, func(ctx context.Context, i int, _ int) error {
				messageTopicPair := messageTopicPairs[keyless[i]]
				if err := PublishToTopic(ctx, messageTopicPair.Message, p.Topics[messageTopicPair.Topic]); err != nil {
					UpdateFailureMetrics(messageTopicPair.Message)
					failed[keyless[i]] = true
					return err
				}
				// if message is invalid (sent to DL), update DL metrics
				if messageTopicPair.Topic == p.Deadletter {
					UpdateSuccessDLMetrics(messageTopicPair.Message)
				}
				UpdateSuccessMetrics(messageTopicPair.Message)
				return nil
			})
		})
	}

	for _, partition := range partitions {
		partition := partition
		eg.Go(func() error {
			// Publishing needs to be sequential on a per-partition basis.
			for _, index := range partition {
				messageTopicPair := messageTopicPairs[index]
				if err := PublishToTopic(ctx, messageTopicPair.Message, p.Topics[messageTopicPair.Topic]); err != nil {
					UpdateFailureMetrics(messageTopicPair.Message)
					failed[index] = true
					return err
				}
				// if message is invalid (sent to DL), update DL metrics
				if messageTopicPair.Topic == p.Deadletter {
					UpdateSuccessDLMetrics(messageTopicPair.Message)
				}
				UpdateSuccessMetrics(messageTopicPair.Message)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return &streamproc.PartiallyProcessedBatchError{
			Failed: indicesWhereTrue(failed),
			Err:    err,
		}
	}
	return nil
}

// indicesWhereTrue collects the indices of those elements in the bool slice which are set to true.
func indicesWhereTrue(bools []bool) []int {
	var indices []int
	for i, isFailed := range bools {
		if isFailed {
			indices = append(indices, i)
		}
	}
	return indices
}

// groupByKey groups the given slice of *MessageTopicPair by the contents of the Message.Key, returning
// a map which for every unique key, holds the indices of all elements which share that key.
func groupByKey(messageTopicPairs []*MessageTopicPair) map[string][]int {
	groups := make(map[string][]int)
	for i, messageTopicPair := range messageTopicPairs {
		if messageTopicPair == nil {
			continue
		}
		groups[messageTopicPair.Message.Key] = append(groups[messageTopicPair.Message.Key], i)
	}
	return groups
}

// AsReceiverExecutor returns a new streamproc.ReceiverExecutor, referencing this instance.
func (p *Processor) AsReceiverExecutor() *streamproc.ReceiverExecutor {
	return streamproc.NewReceiverExecutor(p)
}

// AsBatchedReceiverExecutor returns a new streamproc.BatchedReceiverExecutor, referencing this instance.
func (p *Processor) AsBatchedReceiverExecutor() *streamproc.BatchedReceiverExecutor {
	return streamproc.NewBatchedReceiverExecutor(p)
}

// AsRecordExecutor returns a new streamproc.RecordExecutor, referencing this instance.
func (p *Processor) AsRecordExecutor() *streamproc.RecordExecutor {
	return streamproc.NewRecordExecutor(p)
}

// AsBatchExecutor returns a new streamproc.BatchExecutor, referencing this instance.
func (p *Processor) AsBatchExecutor() *streamproc.BatchExecutor {
	return streamproc.NewBatchExecutor(p)
}
