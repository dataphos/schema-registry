package producer

import (
	"context"
	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/janitor"
	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
	"log"
	"math"
	"time"

	"github.com/dataphos/aquarium-janitor-standalone-internal/internal/registry"
	"github.com/dataphos/lib-brokers/pkg/broker"

	"golang.org/x/sync/errgroup"
)

type Mode int

const (
	Default Mode = iota
	OneCCPerTopic
)

// Producer models a producer used for schema registry and the publishing of the given dataset.
type Producer struct {
	Registry      registry.SchemaRegistry
	Topic         broker.Topic
	Frequency     int
	Mode          Mode
	EncryptionKey string
}

// New returns a new Producer instance.
func New(registry registry.SchemaRegistry, topic broker.Topic, rate int, mode Mode, encryptionKey string) *Producer {
	return &Producer{
		Registry:      registry,
		Topic:         topic,
		Frequency:     rate,
		Mode:          mode,
		EncryptionKey: encryptionKey,
	}
}

// LoadAndProduce loads the dataset from the given filename, assuming the directory to the dataset is given with baseDir,
// (optionally) registers and publishes the loaded messages exactly n times.
func (p *Producer) LoadAndProduce(ctx context.Context, baseDir, filename string, n int) error {
	log.Printf("loading entries from %s...\n", filename)
	start := time.Now()
	entries, err := LoadEntries(filename)
	if err != nil {
		return err
	}
	log.Println("entries loaded in", time.Since(start))

	log.Println("loading messages and schemas...")
	start = time.Now()
	processedEntries, err := ProcessEntries(baseDir, entries)
	if err != nil {
		return err
	}
	log.Println("messages and schemas loaded in", time.Since(start))

	log.Println("converting into broker messages...")
	start = time.Now()
	messages, err := IntoBrokerMessages(ctx, processedEntries, p.Registry)
	if err != nil {
		return err
	}
	log.Println("converted into broker messages in", time.Since(start))

	log.Printf("publishing...")
	start = time.Now()
	if err = p.Publish(ctx, messages, n); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return err
	}
	log.Println("published", n, "messages in", time.Since(start))
	return nil
}

// Publish publishes the given broker.Message slice n times, in a round-robin way in case the size of the dataset is smaller
// than n.
func (p *Producer) Publish(ctx context.Context, messages []broker.OutboundMessage, n int) error {
	datasetSize := len(messages)

	if n <= 0 {
		n = math.MaxInt64
	}

	rl := ratelimit.NewUnlimited()
	if p.Frequency > 0 {
		// throws error if rate <= 0
		rl = ratelimit.New(p.Frequency) // per second
	}

	eg, ctx := errgroup.WithContext(ctx)
LOOP:
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}

		message := messages[i%datasetSize]
		if p.EncryptionKey != "" {
			encryptedData, err := janitor.Encrypt(message.Data, p.EncryptionKey)
			if err != nil {
				return err
			}
			message.Data = encryptedData
		}

		rl.Take()
		eg.Go(func() error {
			return p.Topic.Publish(ctx, message)
		})

	}
	return eg.Wait()
}
