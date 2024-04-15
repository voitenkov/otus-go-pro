//go:build integration
// +build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	internalhttp "github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/server/http"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type HttpTestSuite struct {
	suite.Suite
	host  string
	port  string
	event *internalhttp.EventRequest
}

var (
	configFile string
	eventID    string
	ctx        context.Context
)

const userID = "14e4a342-2ad9-4e1f-bd83-eff99332a49f"

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/calendar_config_test.yaml", "Path to configuration file")
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, &HttpTestSuite{})
}

func (s *HttpTestSuite) SetupSuite() {
	flag.Parse()

	cfg, err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}

	s.host = cfg.Server.Host
	s.port = cfg.Server.Port
	ctx = context.Background()
}

func (s *HttpTestSuite) SetupTest() {
	startTime, err := time.Parse(time.DateTime, "2024-01-02 15:00:00")
	s.NoError(err)
	finishTime, err := time.Parse(time.DateTime, "2024-01-02 16:00:00")
	s.NoError(err)

	s.event = &internalhttp.EventRequest{
		Title:        "Meeting",
		Description:  "Very important meeting",
		StartTime:    storage.EventTime(startTime),
		FinishTime:   storage.EventTime(finishTime),
		NotifyBefore: 30,
	}
}

func (s *HttpTestSuite) TearDownTest() {
	s.event = nil
}

func (s *HttpTestSuite) request(ctx context.Context, method, path string, body io.Reader) *http.Response {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("http://%s:%s/%s", s.host, s.port, path), body)
	s.NoError(err)
	req.Header.Add("X-User-Id", userID)
	response, err := client.Do(req)
	s.NoError(err)
	return response
}

func (s *HttpTestSuite) respBody(response *http.Response) string {
	respBody, err := io.ReadAll(response.Body)
	s.NoError(err)
	respBodyJSON := string(respBody)
	defer response.Body.Close()
	return respBodyJSON
}

func (s *HttpTestSuite) TestCreateEmptyBody() {
	response := s.request(ctx, http.MethodPost, "events", nil)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Equal(`{"Status":400,"Message":"failed to unmarshal request body"}`, s.respBody(response))
}

func (s *HttpTestSuite) TestNoUserID() {
	reqBody, err := json.Marshal(s.event)
	s.NoError(err)
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s:%s/events", s.host, s.port), bytes.NewReader(reqBody))
	s.NoError(err)
	response, err := client.Do(req)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Equal(`{"Status":400,"Message":"x-user-id header is not provided"}`, s.respBody(response))
}

func (s *HttpTestSuite) TestFull() {
	reqBody, err := json.Marshal(s.event)
	s.NoError(err)

	// Create event test.
	response := s.request(ctx, http.MethodPost, "events", bytes.NewReader(reqBody))
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(`{"Status":200,"Message":"event was created"}`, s.respBody(response))

	// List events by date test.
	response = s.request(ctx, http.MethodGet, "events/bydate?start_date=2024-01-02", bytes.NewReader(reqBody))
	s.Equal(http.StatusOK, response.StatusCode)
	events := make([]storage.Event, 0)
	respBody, err := io.ReadAll(response.Body)
	s.NoError(err)
	response.Body.Close()
	err = json.Unmarshal(respBody, &events)
	s.NoError(err)
	s.Equal(1, len(events))
	eventID = events[0].ID.String()
	s.Equal("Meeting", events[0].Title)
	s.Equal("Very important meeting", events[0].Description)
	s.Equal("2024-01-02 15:00:00", time.Time(events[0].StartTime).Format(time.DateTime))
	s.Equal("2024-01-02 16:00:00", time.Time(events[0].FinishTime).Format(time.DateTime))
	s.Equal(30, events[0].NotifyBefore)
	s.False(events[0].NotificationSent) // check if notification was NOT sent

	// Update event test.
	s.event.Title = "Wedding"
	s.event.Description = "Very important wedding"
	startTime, err := time.Parse(time.DateTime, "2024-02-01 15:00:00")
	s.NoError(err)
	finishTime, err := time.Parse(time.DateTime, "2024-02-01 16:00:00")
	s.NoError(err)
	s.event.StartTime = storage.EventTime(startTime)
	s.event.FinishTime = storage.EventTime(finishTime)
	s.event.NotifyBefore = 60
	reqBody, err = json.Marshal(s.event)
	s.NoError(err)
	response = s.request(ctx, http.MethodPut, "events/"+eventID, bytes.NewReader(reqBody))
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(`{"Status":200,"Message":"event was updated"}`, s.respBody(response))

	time.Sleep(time.Minute)

	// List events by month (after update) test.
	response = s.request(ctx, http.MethodGet, "events/bymonth?start_date=2024-01-02", bytes.NewReader(reqBody))
	s.Equal(http.StatusOK, response.StatusCode)
	respBody, err = io.ReadAll(response.Body)
	s.NoError(err)
	response.Body.Close()
	events = make([]storage.Event, 0)
	err = json.Unmarshal(respBody, &events)
	s.NoError(err)
	s.Equal(1, len(events))
	eventID = events[0].ID.String()
	s.Equal("Wedding", events[0].Title)
	s.Equal("Very important wedding", events[0].Description)
	s.Equal("2024-02-01 15:00:00", time.Time(events[0].StartTime).Format(time.DateTime))
	s.Equal("2024-02-01 16:00:00", time.Time(events[0].FinishTime).Format(time.DateTime))
	s.Equal(60, events[0].NotifyBefore)
	s.True(events[0].NotificationSent) // check if notification was sent

	// Delete event test.
	response = s.request(ctx, http.MethodDelete, "events/"+eventID, nil)
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(`{"Status":200,"Message":"event was deleted"}`, s.respBody(response))

	// List events by month (after delete) test.
	response = s.request(ctx, http.MethodGet, "events/bymonth?start_date=2024-01-02", bytes.NewReader(reqBody))
	s.Equal(http.StatusOK, response.StatusCode)
	respBody, err = io.ReadAll(response.Body)
	s.NoError(err)
	response.Body.Close()
	events = make([]storage.Event, 0)
	err = json.Unmarshal(respBody, &events)
	s.NoError(err)
	s.Equal(0, len(events))
}
