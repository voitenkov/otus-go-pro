package internalgrpc

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/logger"
	initstorage "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage/init"
)

var eventID string

const userID = "14e4a342-2ad9-4e1f-bd83-eff99332a49f"

func prepareServer() *GRPCServer {
	cfg := &config.Config{}
	cfg.Logger.Level = "info"
	cfg.Server.Host = "localhost"
	cfg.Server.Port = "8081"
	cfg.DB.Type = "memory"

	logg := logger.New("info")
	memorystorage, err := initstorage.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	calendar := app.New(memorystorage)
	return NewGRPCServer(logg, calendar, cfg)
}

func TestServer(t *testing.T) {
	s := prepareServer()
	ctx := context.Background()

	t.Run("Create rpc test", func(t *testing.T) {
		request := &Event{
			UserId:       userID,
			Title:        "Meeting",
			Description:  "Very important meeting",
			StartTime:    "2024-01-02 15:00:00",
			FinishTime:   "2024-01-02 16:00:00",
			NotifyBefore: 30,
		}

		response, err := s.Create(ctx, request)
		require.NoError(t, err)
		require.Equal(t, 1, int(response.Result))
	})

	t.Run("ListEventsByDay rpc test", func(t *testing.T) {
		request := &EventsListRequest{
			UserId:    userID,
			StartDate: "2024-01-02",
		}

		response, err := s.ListEventsByDay(ctx, request)
		require.NoError(t, err)
		require.Equal(t, 1, len(response.EventsList))
		require.Equal(t, "Meeting", response.EventsList[0].GetEvent().GetTitle())
		require.Equal(t, "Very important meeting", response.EventsList[0].GetEvent().GetDescription())
		require.Equal(t, "2024-01-02 15:00:00", response.EventsList[0].GetEvent().GetStartTime())
		require.Equal(t, "2024-01-02 16:00:00", response.EventsList[0].GetEvent().GetFinishTime())
		require.Equal(t, 30, int(response.EventsList[0].GetEvent().GetNotifyBefore()))
		eventID = response.EventsList[0].GetId()
	})

	t.Run("Update rpc test", func(t *testing.T) {
		request := &EventWithID{
			Id: eventID,
			Event: &Event{
				UserId:       userID,
				Title:        "Wedding",
				Description:  "Very important wedding",
				StartTime:    "2024-02-01 15:00:00",
				FinishTime:   "2024-02-01 16:00:00",
				NotifyBefore: 60,
			},
		}

		response, err := s.Update(ctx, request)
		require.NoError(t, err)
		require.Equal(t, 1, int(response.Result))
	})

	t.Run("ListEventsByMonth rpc test (after update)", func(t *testing.T) {
		request := &EventsListRequest{
			UserId:    userID,
			StartDate: "2024-01-02",
		}

		response, err := s.ListEventsByMonth(ctx, request)
		require.NoError(t, err)
		require.Equal(t, 1, len(response.EventsList))
		require.Equal(t, "Wedding", response.EventsList[0].GetEvent().GetTitle())
		require.Equal(t, "Very important wedding", response.EventsList[0].GetEvent().GetDescription())
		require.Equal(t, "2024-02-01 15:00:00", response.EventsList[0].GetEvent().GetStartTime())
		require.Equal(t, "2024-02-01 16:00:00", response.EventsList[0].GetEvent().GetFinishTime())
		require.Equal(t, 60, int(response.EventsList[0].GetEvent().GetNotifyBefore()))
		eventID = response.EventsList[0].GetId()
	})

	t.Run("Delete rpc test", func(t *testing.T) {
		request := &EventID{
			Id: eventID,
		}

		response, err := s.Delete(ctx, request)
		require.NoError(t, err)
		require.Equal(t, 1, int(response.Result))
	})

	t.Run("ListEventsByMonth rpc test (after delete)", func(t *testing.T) {
		request := &EventsListRequest{
			UserId:    userID,
			StartDate: "2024-01-02",
		}

		response, err := s.ListEventsByMonth(ctx, request)
		require.NoError(t, err)
		require.Equal(t, 0, len(response.EventsList))
	})
}
