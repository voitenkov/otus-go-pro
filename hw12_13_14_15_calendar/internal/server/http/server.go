package internalhttp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

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
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	StartTime    storage.EventTime     `json:"startTime"`
	Duration     storage.EventDuration `json:"duration"`
	NotifyBefore storage.EventDuration `json:"notifyBefore"`
}

type UserID string

func NewServer(logger Logger, app Application, cfg *config.Config) *Server {
	return &Server{
		host:   cfg.Server.Host,
		port:   cfg.Server.Port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	userID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(s.host, s.port)
	helloWorldHandlerFunc := http.HandlerFunc(s.HelloWorld)
	injectUserIDHandler := injectUserID(helloWorldHandlerFunc, userID)
	handler := s.loggingMiddleware(injectUserIDHandler)
	http.Handle("GET /hello/", handler)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 5,
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
func (s *Server) HelloWorld(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if m := r.Context().Value("userID"); m != nil {
		if value, ok := m.(string); ok {
			userID = value
		}
	}
	w.Header().Add("userID", userID)
	w.Write([]byte("Hello, world, from UserID " + userID))
}

func injectUserID(next http.Handler, userID uuid.UUID) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserID("userID"), userID)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

// TODO
