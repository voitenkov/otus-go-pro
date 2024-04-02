package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/app"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/logger"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
	initstorage "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage/init"
)

var eventID string

const userID = "14e4a342-2ad9-4e1f-bd83-eff99332a49f"

func prepareServer() *Server {
	cfg := &config.Config{}
	cfg.Logger.Level = "info"
	cfg.Server.Host = "localhost"
	cfg.Server.Port = "8080"
	cfg.DB.Type = "memory"

	logg := logger.New("info")
	memorystorage, err := initstorage.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	calendar := app.New(memorystorage)
	return NewServer(logg, calendar, cfg)
}

func TestServer(t *testing.T) {
	s := prepareServer()
	ctx := context.Background()
	t.Run("createEventHandler test", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/events", s.createEventHandler).Methods("POST")
		server := httptest.NewServer(router)
		defer server.Close()
		client := http.Client{
			Timeout: 30 * time.Second,
		}
		eventReqBodyJSON := `{"title":"Meeting","description":"Very important meeting","startTime": 
		"2024-01-02 15:00:00","finishTime": "2024-01-02 16:00:00","notifyBefore": 30}`
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/events",
			bytes.NewReader([]byte(eventReqBodyJSON)))
		require.NoError(t, err)
		req.Header.Add("X-User-Id", userID)
		response, err := client.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, `{"Status":200,"Message":"event was created"}`, string(respBody))
	})

	t.Run("listEventsByDateHandler test", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/events/bydate", s.listEventsByDateHandler).Methods("GET")
		server := httptest.NewServer(router)
		defer server.Close()
		client := http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/events/bydate?start_date=2024-01-02", nil)
		require.NoError(t, err)
		req.Header.Add("X-User-Id", userID)
		response, err := client.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		events := make([]storage.Event, 0)
		err = json.Unmarshal(respBody, &events)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, 1, len(events))
		eventID = events[0].ID.String()
		require.Equal(t, "Meeting", events[0].Title)
		require.Equal(t, "Very important meeting", events[0].Description)
		require.Equal(t, "2024-01-02 15:00:00", time.Time(events[0].StartTime).Format(time.DateTime))
		require.Equal(t, "2024-01-02 16:00:00", time.Time(events[0].FinishTime).Format(time.DateTime))
		require.Equal(t, 30, events[0].NotifyBefore)
	})

	t.Run("updateEventHandler test", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/events/{ID}", s.updateEventHandler).Methods("PUT")
		server := httptest.NewServer(router)
		defer server.Close()
		client := http.Client{
			Timeout: 30 * time.Second,
		}
		eventReqBodyJSON := `{"title":"Wedding","description":"Very important wedding",
		"startTime": "2024-02-01 15:00:00","finishTime": "2024-02-01 16:00:00","notifyBefore": 60}`
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL+"/events/"+eventID,
			bytes.NewReader([]byte(eventReqBodyJSON)))
		require.NoError(t, err)
		req.Header.Add("X-User-Id", userID)
		response, err := client.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, `{"Status":200,"Message":"event was updated"}`, string(respBody))
	})

	t.Run("listEventsByMonthHandler (after update) test", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/events/bymonth", s.listEventsByMonthHandler).Methods("GET")
		server := httptest.NewServer(router)
		defer server.Close()
		client := http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/events/bymonth?start_date=2024-01-02", nil)
		require.NoError(t, err)
		req.Header.Add("X-User-Id", userID)
		response, err := client.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		events := make([]storage.Event, 0)
		err = json.Unmarshal(respBody, &events)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, 1, len(events))
		eventID = events[0].ID.String()
		require.Equal(t, "Wedding", events[0].Title)
		require.Equal(t, "Very important wedding", events[0].Description)
		require.Equal(t, "2024-02-01 15:00:00", time.Time(events[0].StartTime).Format(time.DateTime))
		require.Equal(t, "2024-02-01 16:00:00", time.Time(events[0].FinishTime).Format(time.DateTime))
		require.Equal(t, 60, events[0].NotifyBefore)
	})

	t.Run("deleteEventHandler test", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/events/{ID}", s.deleteEventHandler).Methods("DELETE")
		server := httptest.NewServer(router)
		defer server.Close()
		client := http.Client{
			Timeout: 30 * time.Second,
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, server.URL+"/events/"+eventID, nil)
		require.NoError(t, err)
		response, err := client.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, `{"Status":200,"Message":"event was deleted"}`, string(respBody))
	})
	t.Run("listEventsByMonthHandler (after delete) test", func(t *testing.T) {
		router := mux.NewRouter()
		router.HandleFunc("/events/bymonth", s.listEventsByMonthHandler).Methods("GET")
		server := httptest.NewServer(router)
		defer server.Close()
		client := http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/events/bymonth?start_date=2024-01-02", nil)
		require.NoError(t, err)
		req.Header.Add("X-User-Id", userID)
		response, err := client.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		events := make([]storage.Event, 0)
		err = json.Unmarshal(respBody, &events)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, 0, len(events))
	})
}
