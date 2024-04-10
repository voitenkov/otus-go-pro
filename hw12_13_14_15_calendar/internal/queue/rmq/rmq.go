package rmqqueue

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/queue"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type Queue struct {
	config config.Config
	uri    string
	conn   *amqp.Connection
}

var rmqqueue amqp.Queue

func (q *Queue) Connect() error {
	var err error
	q.conn, err = amqp.Dial(q.uri)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	return nil
}

func (q *Queue) Close() error {
	return q.conn.Close()
}

func (q *Queue) PublishNotifications(ctx context.Context,
	events []storage.Event,
) (eventsOut []storage.Event, err error) {
	connection, err := amqp.Dial(q.uri)
	queueName := q.config.Queue.RMQ.Name
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel error: %w", err)
	}

	if rmqqueue, err = channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	); err != nil {
		return nil, fmt.Errorf("queue declaration error: %w", err)
	}

	for i, event := range events {
		body, err := json.Marshal(GetNotificationFromEvent(event))
		if err != nil {
			return nil, err
		}
		if err = channel.PublishWithContext(
			ctx,
			"",            // publish to an exchange
			rmqqueue.Name, // routing to 0 or more queues
			// Mandatory flag tells the server how to react if a message cannot be routed to a queue.
			// Specifically, if mandatory is set and after running the bindings
			// the message was placed on zero queues then the message is returned to the sender (with a basic.return).
			// If mandatory had not been set under the same circumstances the server would silently drop the message.
			false,
			// immediate.
			// If there is at least one consumer connected to my queue that can take delivery of a message
			// right this moment, deliver this message to them immediately.
			// If there are no consumers connected then there's no point in having my message consumed later
			// and they'll never see it. They snooze, they lose.
			false,
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            body,
				DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			},
		); err != nil {
			return nil, fmt.Errorf("exchange publishing error: %w", err)
		}
		events[i].NotificationSent = true
	}
	return events, nil
}

func GetNotificationFromEvent(event storage.Event) queue.Notification {
	notification := &queue.Notification{
		ID:        event.ID,
		UserID:    event.UserID,
		Title:     event.Title,
		StartTime: event.StartTime,
	}
	return *notification
}

func (q *Queue) ReadAndProcessNotifications(ctx context.Context, fn app.CallbackFunc) error {
	queueName := q.config.Queue.RMQ.Name
	channel, err := q.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel error: %w", err)
	}

	if rmqqueue, err = channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	); err != nil {
		return fmt.Errorf("queue declaration error: %w", err)
	}

	notifications, err := channel.ConsumeWithContext(
		ctx,
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed open channel: %w", err)
	}

	go func() {
		for notification := range notifications {
			fn(notification.Body)
		}
	}()

	<-ctx.Done()
	return nil
}

func New(config *config.Config, uri string) *Queue {
	return &Queue{
		config: *config,
		uri:    uri,
	}
}
