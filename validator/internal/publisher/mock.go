package publisher

import (
	"golang.org/x/net/context"

	"github.com/dataphos/lib-brokers/pkg/broker"
)

type MockPublisher struct {
}

type MockTopic struct {
}

func (t *MockTopic) BatchPublish(context.Context, ...broker.OutboundMessage) error {
	return nil
}

func (*MockTopic) Publish(context.Context, broker.OutboundMessage) error {
	return nil
}

func (*MockPublisher) Topic(_ string) (broker.Topic, error) {
	return &MockTopic{}, nil
}
