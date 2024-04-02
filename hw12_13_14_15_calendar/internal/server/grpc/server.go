//go:generate protoc -I ../../../api/ --go_out=. --go-grpc_out=. ../../../api/EventService.proto

package internalgrpc

import (
	"context"
	"net"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type GRPCServer struct {
	host   string
	port   string
	logger Logger
	app    Application
	server *grpc.Server
}

type Logger interface {
	Error(msg ...interface{})
	Info(msg ...interface{})
	Warn(msg ...interface{})
	Debug(msg ...interface{})
	LogGRPCRequest(ctx context.Context, info *grpc.UnaryServerInfo, duration time.Duration, statusCode string)
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

type listEventsFunc func(context.Context, uuid.UUID, storage.EventDate) ([]storage.Event, error)

func NewGRPCServer(logger Logger, app Application, cfg *config.Config) *GRPCServer {
	return &GRPCServer{
		host:   cfg.GRPCServer.Host,
		port:   cfg.GRPCServer.Port,
		logger: logger,
		app:    app,
	}
}

func (s *GRPCServer) Start(ctx context.Context) error {
	server := grpc.NewServer(grpc.UnaryInterceptor(s.loggingInterceptor))
	s.server = server
	RegisterEventServiceServer(server, s)

	addr := net.JoinHostPort(s.host, s.port)

	go func() {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			s.logger.Error(err)
			return
		}

		err = server.Serve(listener)
		if err != nil {
			s.logger.Error(err)
		}
	}()

	s.logger.Info("GRPC server starting on tcp://" + addr)

	<-ctx.Done()
	return nil
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

func (s *GRPCServer) Create(ctx context.Context, event *Event) (*EventResponse, error) {
	userID, err := uuid.FromString(event.GetUserId())
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	startTimeParsed, err := time.Parse(time.DateTime, event.GetStartTime())
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	startTime := storage.EventTime(startTimeParsed)
	finishTimeParsed, err := time.Parse(time.DateTime, event.GetFinishTime())
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	finishTime := storage.EventTime(finishTimeParsed)

	err = s.app.CreateEvent(ctx, userID, event.GetTitle(), event.GetDescription(), startTime, finishTime,
		int(event.GetNotifyBefore()))
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	return &EventResponse{
		Result: 1,
	}, nil
}

func (s *GRPCServer) Update(ctx context.Context, event *EventWithID) (*EventResponse, error) {
	id, err := uuid.FromString(event.GetId())
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	userID, err := uuid.FromString(event.GetEvent().UserId)
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	startTimeParsed, err := time.Parse(time.DateTime, event.GetEvent().StartTime)
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	startTime := storage.EventTime(startTimeParsed)
	finishTimeParsed, err := time.Parse(time.DateTime, event.GetEvent().FinishTime)
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	finishTime := storage.EventTime(finishTimeParsed)

	err = s.app.UpdateEvent(ctx, id, userID, event.GetEvent().Title, event.GetEvent().Description, startTime,
		finishTime, int(event.GetEvent().NotifyBefore))
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	return &EventResponse{
		Result: 1,
	}, nil
}

func (s *GRPCServer) Delete(ctx context.Context, event *EventID) (*EventResponse, error) {
	id, err := uuid.FromString(event.GetId())
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	err = s.app.DeleteEvent(ctx, id)
	if err != nil {
		return &EventResponse{
			Result: 0,
		}, err
	}

	return &EventResponse{
		Result: 1,
	}, nil
}

func (s *GRPCServer) ListEventsByDay(ctx context.Context, request *EventsListRequest) (*EventsListResponse, error) {
	return s.listEventsUntyped(ctx, s.app.ListEventsByDate, request)
}

func (s *GRPCServer) ListEventsByWeek(ctx context.Context, request *EventsListRequest) (*EventsListResponse, error) {
	return s.listEventsUntyped(ctx, s.app.ListEventsByWeek, request)
}

func (s *GRPCServer) ListEventsByMonth(ctx context.Context, request *EventsListRequest) (*EventsListResponse, error) {
	return s.listEventsUntyped(ctx, s.app.ListEventsByMonth, request)
}

func (s *GRPCServer) listEventsUntyped(ctx context.Context, fn listEventsFunc,
	request *EventsListRequest,
) (*EventsListResponse, error) {
	userID, err := uuid.FromString(request.GetUserId())
	if err != nil {
		return &EventsListResponse{
			EventsList: []*EventWithID{},
		}, err
	}

	startDate := request.GetStartDate()
	dateParsed, err := time.Parse(time.DateOnly, startDate)
	if err != nil {
		return &EventsListResponse{
			EventsList: []*EventWithID{},
		}, err
	}

	events, err := fn(ctx, userID, storage.EventDate(dateParsed))
	if err != nil {
		return &EventsListResponse{
			EventsList: []*EventWithID{},
		}, err
	}

	eventsList := make([]*EventWithID, 0)
	for _, event := range events {
		eventStruct := &Event{
			UserId:       event.UserID.String(),
			Title:        event.Title,
			Description:  event.Description,
			StartTime:    time.Time(event.StartTime).Format(time.DateTime),
			FinishTime:   time.Time(event.FinishTime).Format(time.DateTime),
			NotifyBefore: int32(event.NotifyBefore),
		}
		eventWithID := &EventWithID{
			Id:    event.ID.String(),
			Event: eventStruct,
		}
		eventsList = append(eventsList, eventWithID)
	}

	return &EventsListResponse{
		EventsList: eventsList,
	}, err
}

func (s *GRPCServer) mustEmbedUnimplementedEventServiceServer() {
	s.logger.Error("unimplemented server")
}
