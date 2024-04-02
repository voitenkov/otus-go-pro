package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

var ErrUserIDHeader = errors.New("error getting userid from header ")

type Server struct {
	host   string
	port   string
	logger Logger
	app    Application
	server *http.Server
}

type Logger interface {
	Error(msg ...interface{})
	Info(msg ...interface{})
	Warn(msg ...interface{})
	Debug(msg ...interface{})
	LogHTTPRequest(request *http.Request, duration time.Duration, statusCode int)
}

type Application interface {
	CreateEvent(ctx context.Context, userID uuid.UUID, title, description string, startTime,
		finishTime storage.EventTime, notifyBefore int) error
	ListEventsByDate(ctx context.Context, userID uuid.UUID, date storage.EventDate) ([]storage.Event, error)
	ListEventsByWeek(ctx context.Context, userID uuid.UUID, date storage.EventDate) ([]storage.Event, error)
	ListEventsByMonth(ctx context.Context, userID uuid.UUID, date storage.EventDate) ([]storage.Event, error)
	UpdateEvent(ctx context.Context, ID, userID uuid.UUID, title, description string, startTime,
		finishTime storage.EventTime, notifyBefore int) error
	DeleteEvent(ctx context.Context, ID uuid.UUID) error
}

type EventRequest struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	StartTime    storage.EventTime `json:"startTime"`
	FinishTime   storage.EventTime `json:"finishTime"`
	NotifyBefore int               `json:"notifyBefore"`
}

func (er *EventRequest) UnmarshalJSON(data []byte) (err error) {
	var startTime, finishTime time.Time
	var tmp struct {
		Title        string
		Description  string
		StartTime    string
		FinishTime   string
		NotifyBefore int
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	er.Title = tmp.Title
	er.Description = tmp.Description
	er.NotifyBefore = tmp.NotifyBefore
	startTime, err = time.Parse(time.DateTime, tmp.StartTime)
	if err != nil {
		return err
	}

	er.StartTime = storage.EventTime(startTime)
	finishTime, err = time.Parse(time.DateTime, tmp.FinishTime)
	if err != nil {
		return err
	}

	er.FinishTime = storage.EventTime(finishTime)
	er.NotifyBefore = tmp.NotifyBefore
	return err
}

type ServerResponse struct {
	Status  int
	Message string
}

type UserID string

type listEventsFunc func(context.Context, uuid.UUID, storage.EventDate) ([]storage.Event, error)

func NewServer(logger Logger, app Application, cfg *config.Config) *Server {
	return &Server{
		host:   cfg.Server.Host,
		port:   cfg.Server.Port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.host, s.port)
	router := mux.NewRouter()
	router.HandleFunc("/hello", s.helloWorldHandler)
	router.HandleFunc("/events", s.createEventHandler).Methods("POST")
	router.HandleFunc("/events/{ID}", s.updateEventHandler).Methods("PUT")
	router.HandleFunc("/events/{ID}", s.deleteEventHandler).Methods("DELETE")
	router.HandleFunc("/events/bydate", s.listEventsByDateHandler).Methods("GET")
	router.HandleFunc("/events/byweek", s.listEventsByWeekHandler).Methods("GET")
	router.HandleFunc("/events/bymonth", s.listEventsByMonthHandler).Methods("GET")
	router.Use(s.loggingMiddleware)

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 5,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}

	s.server = server

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			s.logger.Error(err)
		}
	}()

	s.logger.Info("server starting on http://" + addr)

	<-ctx.Done()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Hello world handler.
func (s *Server) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	userID := ""
	key := UserID("userID")
	if m := r.Context().Value(key); m != nil {
		userID = m.(string)
	}
	w.Header().Add("userID", userID)
	w.Write([]byte("Hello, world, from UserID " + userID))
}

// Create event handler.
func (s *Server) createEventHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := s.getUserID(w, r)
	if err != nil {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to read request body", w)
		return
	}
	defer r.Body.Close()

	data := EventRequest{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to unmarshal request body", w)
		return
	}

	err = s.app.CreateEvent(r.Context(), userID, data.Title, data.Description, data.StartTime, data.FinishTime,
		data.NotifyBefore)
	if err != nil {
		s.writeResponse(http.StatusInternalServerError, err.Error(), w)
		s.logger.Error(err)
		return
	}

	s.writeResponse(http.StatusOK, "event was created", w)
}

// Update event handler.
func (s *Server) updateEventHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := s.getUserID(w, r)
	if err != nil {
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["ID"])
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to parse id path parameter", w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to read request body", w)
		return
	}
	defer r.Body.Close()

	data := &EventRequest{}
	err = json.Unmarshal(body, data)
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to unmarshal request body", w)
		return
	}

	err = s.app.UpdateEvent(r.Context(), id, userID, data.Title, data.Description, data.StartTime,
		data.FinishTime, data.NotifyBefore)
	if err != nil {
		s.writeResponse(http.StatusInternalServerError, err.Error(), w)
		s.logger.Error(err)
		return
	}

	s.writeResponse(http.StatusOK, "event was updated", w)
}

// Delete event handler.
func (s *Server) deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.FromString(vars["ID"])
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to parse id path parameter", w)
		return
	}

	err = s.app.DeleteEvent(r.Context(), id)
	if err != nil {
		s.writeResponse(http.StatusInternalServerError, err.Error(), w)
		s.logger.Error(err)
		return
	}

	s.writeResponse(http.StatusOK, "event was deleted", w)
}

// List events by date handler.
func (s *Server) listEventsByDateHandler(w http.ResponseWriter, r *http.Request) {
	s.listEventsUntyped(s.app.ListEventsByDate, w, r)
}

// List events by week handler.
func (s *Server) listEventsByWeekHandler(w http.ResponseWriter, r *http.Request) {
	s.listEventsUntyped(s.app.ListEventsByWeek, w, r)
}

// List events by month handler.
func (s *Server) listEventsByMonthHandler(w http.ResponseWriter, r *http.Request) {
	s.listEventsUntyped(s.app.ListEventsByMonth, w, r)
}

func (s *Server) listEventsUntyped(fn listEventsFunc, w http.ResponseWriter, r *http.Request) {
	userID, err := s.getUserID(w, r)
	if err != nil {
		return
	}

	query := r.URL.Query()
	startDate := query.Get("start_date")
	dateParsed, err := time.Parse(time.DateOnly, startDate)
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to parse start_date query parameter", w)
		return
	}

	events, err := fn(r.Context(), userID, storage.EventDate(dateParsed))
	if err != nil {
		s.writeResponse(http.StatusInternalServerError, "internal server error", w)
		s.logger.Error(err)
		return
	}

	res, err := json.Marshal(events)
	if err != nil {
		s.writeResponse(http.StatusInternalServerError, "internal server error", w)
		s.logger.Error(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		s.writeResponse(http.StatusInternalServerError, "internal server error", w)
		s.logger.Error(err)
	}
}

func (s *Server) writeResponse(status int, message string, w http.ResponseWriter) {
	res, err := json.Marshal(ServerResponse{
		Status:  status,
		Message: message,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error(err)
	}

	w.WriteHeader(status)

	_, err = w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error(err)
	}
}

func (s *Server) getUserID(w http.ResponseWriter, r *http.Request) (userID uuid.UUID, err error) {
	userIDSlice, found := r.Header["X-User-Id"]
	if !found {
		s.writeResponse(http.StatusBadRequest, "x-user-id header is not provided", w)
		return userID, ErrUserIDHeader
	}

	if len(userIDSlice) == 0 {
		s.writeResponse(http.StatusBadRequest, "x-user-id header is empty", w)
		return userID, ErrUserIDHeader
	}

	userID, err = uuid.FromString(userIDSlice[0])
	if err != nil {
		s.writeResponse(http.StatusBadRequest, "failed to parse x-user-id header", w)
		return userID, ErrUserIDHeader
	}

	return userID, nil
}
