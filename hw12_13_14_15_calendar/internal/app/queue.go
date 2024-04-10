package app

import (
	"context"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type CallbackFunc func([]byte)

type QueueApp struct {
	queue Queue
}

type Queue interface {
	Connect() error
	Close() error
	PublishNotifications(ctx context.Context, events []storage.Event) (eventsOut []storage.Event, err error)
	ReadAndProcessNotifications(ctx context.Context, fn CallbackFunc) error
}

func (a *QueueApp) PublishNotifications(ctx context.Context,
	events []storage.Event,
) (eventsOut []storage.Event, err error) {
	return a.queue.PublishNotifications(ctx, events)
}

func (a *QueueApp) ReadAndProcessNotifications(ctx context.Context, fn CallbackFunc) error {
	return a.queue.ReadAndProcessNotifications(ctx, fn)
}
