package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/queue"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type Sender struct {
	logger Logger
	app    Application
	queue  QueueApplication
}

type Logger interface {
	Error(msg ...interface{})
	Info(msg ...interface{})
	Infof(format string, args ...interface{})
	Warn(msg ...interface{})
	Debug(msg ...interface{})
}

type Application interface {
	PatchEvent(ctx context.Context, id uuid.UUID, userID *uuid.UUID, title, description *string, startTime,
		finishTime *storage.EventTime, notifyBefore *int, notificationSent *bool) error
}

type QueueApplication interface {
	Connect() error
	Close() error
	ReadAndProcessNotifications(ctx context.Context, fn app.CallbackFunc) error
}

func New(logger Logger, app Application, queue QueueApplication) *Sender {
	return &Sender{
		logger: logger,
		app:    app,
		queue:  queue,
	}
}

func (s *Sender) Start(ctx context.Context) error {
	err := s.queue.Connect()
	if err != nil {
		s.logger.Error(err)
		return err
	}
	defer s.queue.Close()

	go func() {
		err = s.queue.ReadAndProcessNotifications(ctx, s.SendNotification)
		if err != nil {
			s.logger.Error(err)
		}
	}()

	<-ctx.Done()
	fmt.Println(ctx.Err())
	return nil
}

func (s *Sender) SendNotification(ctx context.Context, body []byte) {
	notification := &queue.Notification{}
	err := json.Unmarshal(body, notification)
	if err != nil {
		s.logger.Error("unmarshal body error: %w", err)
	}

	title := notification.Title
	startTime := time.Time(notification.StartTime).Format(time.DateTime)

	s.logger.Infof("Dear user, pls be reminded on event '%v' at %v", title, startTime)

	notificationSent := true
	err = s.app.PatchEvent(ctx, notification.ID, nil, nil, nil, nil, nil, nil, &notificationSent)
	if err != nil {
		s.logger.Error(err)
		return
	}
}
