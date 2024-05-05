package initstorage

import (
	"fmt"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	rmqqueue "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/queue/rmq"
)

func New(cfg *config.Config) (app.Queue, error) {
	switch cfg.Queue.Type {
	case "rmq":
		rmqConf := cfg.Queue.RMQ
		return rmqqueue.New(cfg, GetURI(rmqConf)), nil
	default:
		return nil, fmt.Errorf("unknown queue type: %q", cfg.Queue.Type)
	}
}

func GetURI(rmq config.RMQConf) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", rmq.User, rmq.Password, rmq.Host, rmq.Port)
}
