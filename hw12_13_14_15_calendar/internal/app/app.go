package app

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	storage Storage
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	ListEventsByDate(ctx context.Context, userID uuid.UUID, startDate storage.EventDate) ([]storage.Event, error)
	ListEventsByWeek(ctx context.Context, userID uuid.UUID, startDate storage.EventDate) ([]storage.Event, error)
	ListEventsByMonth(ctx context.Context, userID uuid.UUID, startDate storage.EventDate) ([]storage.Event, error)
	ListEventsByPeriod(ctx context.Context, userID uuid.UUID, startDate,
		finishDate storage.EventDate) ([]storage.Event, error)
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, ID uuid.UUID) error
	Connect() error
	Close() error
}

func (a *App) CreateEvent(ctx context.Context, userID uuid.UUID, title, description string, startTime,
	finishTime storage.EventTime, notifyBefore int,
) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	event := buildEvent(id, userID, title, description, startTime, finishTime, notifyBefore)
	return a.storage.CreateEvent(ctx, *event)
}

func (a *App) ListEventsByDate(ctx context.Context, userID uuid.UUID, date storage.EventDate) ([]storage.Event, error) {
	return a.storage.ListEventsByDate(ctx, userID, date)
}

func (a *App) ListEventsByWeek(ctx context.Context, userID uuid.UUID,
	date storage.EventDate,
) ([]storage.Event, error) {
	return a.storage.ListEventsByWeek(ctx, userID, date)
}

func (a *App) ListEventsByMonth(ctx context.Context, userID uuid.UUID,
	date storage.EventDate,
) ([]storage.Event, error) {
	return a.storage.ListEventsByMonth(ctx, userID, date)
}

func (a *App) UpdateEvent(ctx context.Context, id, userID uuid.UUID, title, description string, startTime,
	finishTime storage.EventTime, notifyBefore int,
) error {
	event := buildEvent(id, userID, title, description, startTime, finishTime, notifyBefore)
	return a.storage.UpdateEvent(ctx, *event)
}

func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return a.storage.DeleteEvent(ctx, id)
}

func New(storage Storage) *App {
	return &App{
		storage: storage,
	}
}

func buildEvent(id, userID uuid.UUID, title, description string, startTime, finishTime storage.EventTime,
	notifyBefore int,
) *storage.Event {
	event := &storage.Event{
		ID:           id,
		UserID:       userID,
		Title:        title,
		Description:  description,
		StartTime:    startTime,
		FinishTime:   finishTime,
		NotifyBefore: notifyBefore,
	}
	return event
}
