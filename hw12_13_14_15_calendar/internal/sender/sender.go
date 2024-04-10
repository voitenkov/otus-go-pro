package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/queue"
)

type Sender struct {
	logger Logger
	queue  QueueApplication
}

type Logger interface {
	Error(msg ...interface{})
	Info(msg ...interface{})
	Infof(format string, args ...interface{})
	Warn(msg ...interface{})
	Debug(msg ...interface{})
}

type QueueApplication interface {
	Connect() error
	Close() error
	ReadAndProcessNotifications(ctx context.Context, fn app.CallbackFunc) error
}

func New(logger Logger, queue QueueApplication) *Sender {
	return &Sender{
		logger: logger,
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

func (s *Sender) SendNotification(body []byte) {
	notification := &queue.Notification{}
	err := json.Unmarshal(body, notification)
	if err != nil {
		s.logger.Error("unmarshal body error: %w", err)
	}

	title := notification.Title
	startTime := time.Time(notification.StartTime).Format(time.DateTime)

	s.logger.Infof("Dear user, pls be reminded on event '%v' at %v", title, startTime)
}
