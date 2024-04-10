package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type Scheduler struct {
	purgeIntervalDays int
	logger            Logger
	app               Application
	queue             QueueApplication
}

type Logger interface {
	Error(msg ...interface{})
	Info(msg ...interface{})
	Infof(format string, args ...interface{})
	Warn(msg ...interface{})
	Debug(msg ...interface{})
}

type Application interface {
	SelectEventsToNotify(ctx context.Context) ([]storage.Event, error)
	PurgeEvents(ctx context.Context, purgeIntervalDays int) (purgedEvents int64, err error)
	UpdateEvent(ctx context.Context, ID, userID uuid.UUID, title, description string, startTime,
		finishTime storage.EventTime, notifyBefore int, notificationSent bool) error
}

type QueueApplication interface {
	PublishNotifications(ctx context.Context, events []storage.Event) (eventsOut []storage.Event, err error)
}

func New(logger Logger, app Application, queue QueueApplication, cfg *config.Config) *Scheduler {
	return &Scheduler{
		purgeIntervalDays: cfg.Scheduler.PurgeIntervalDays,
		logger:            logger,
		app:               app,
		queue:             queue,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	stop := make(chan bool)

	go func() {
		defer func() { stop <- true }()
		for {
			select {
			case <-ticker.C:
				s.purgeEvents(ctx)
				s.selectEventsToNotify(ctx)
			case <-stop:
				return
			}
		}
	}()

	<-ctx.Done()
	fmt.Println(ctx.Err())
	ticker.Stop()
	stop <- true
	<-stop
	return nil
}

func (s *Scheduler) purgeEvents(ctx context.Context) {
	purgedEvents, err := s.app.PurgeEvents(ctx, s.purgeIntervalDays)
	if err != nil {
		s.logger.Error(err)
		return
	}

	s.logger.Infof("purge events: %v events purged", purgedEvents)
}

func (s *Scheduler) selectEventsToNotify(ctx context.Context) {
	events, err := s.app.SelectEventsToNotify(ctx)
	if err != nil {
		s.logger.Error(err)
		return
	}

	s.logger.Infof("select events to notify: %v events selected", len(events))
	eventsOut, err := s.queue.PublishNotifications(ctx, events)
	if err != nil {
		s.logger.Error(err)
	}

	for _, event := range eventsOut {
		err := s.app.UpdateEvent(ctx, event.ID, event.UserID, event.Title, event.Description, event.StartTime,
			event.FinishTime, event.NotifyBefore, event.NotificationSent)
		if err != nil {
			s.logger.Error(err)
			return
		}
	}
}
