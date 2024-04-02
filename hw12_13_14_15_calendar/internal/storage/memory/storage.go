package memorystorage

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type Events map[uuid.UUID]storage.Event

type Storage struct {
	mu     sync.RWMutex
	events Events
}

var (
	errEventExists   = errors.New("event already exists")
	errEventNotFound = errors.New("event not found")
)

func (s *Storage) Connect() error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	_ = context.WithoutCancel(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; exists {
		return errEventExists
	}

	s.events[event.ID] = event

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = context.WithoutCancel(ctx)
	if _, exists := s.events[event.ID]; !exists {
		return errEventNotFound
	}

	s.events[event.ID] = event

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = context.WithoutCancel(ctx)
	if _, exists := s.events[id]; !exists {
		return errEventNotFound
	}

	delete(s.events, id)

	return nil
}

func (s *Storage) ListEventsByDate(ctx context.Context, userID uuid.UUID,
	startDate storage.EventDate,
) ([]storage.Event, error) {
	_ = context.WithoutCancel(ctx)
	return s.ListEventsByPeriod(ctx, userID, startDate, storage.EventDate(time.Time(startDate).AddDate(0, 0, 1)))
}

func (s *Storage) ListEventsByWeek(ctx context.Context, userID uuid.UUID,
	startDate storage.EventDate,
) ([]storage.Event, error) {
	_ = context.WithoutCancel(ctx)
	return s.ListEventsByPeriod(ctx, userID, startDate, storage.EventDate(time.Time(startDate).AddDate(0, 0, 7)))
}

func (s *Storage) ListEventsByMonth(ctx context.Context, userID uuid.UUID,
	startDate storage.EventDate,
) ([]storage.Event, error) {
	_ = context.WithoutCancel(ctx)
	return s.ListEventsByPeriod(ctx, userID, startDate, storage.EventDate(time.Time(startDate).AddDate(0, 1, 0)))
}

func (s *Storage) ListEventsByPeriod(ctx context.Context, userID uuid.UUID, startDate,
	finishDate storage.EventDate,
) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = context.WithoutCancel(ctx)
	result := make([]storage.Event, 0)

	for _, event := range s.events {
		eventStartTime := time.Time(event.StartTime)
		eventFinishTime := time.Time(event.FinishTime)
		periodStartTime := time.Time(startDate)
		periodFinishTime := time.Time(finishDate)
		if event.UserID == userID && eventStartTime.Compare(periodFinishTime) < 0 &&
			eventFinishTime.Compare(periodStartTime) >= 0 {
			result = append(result, event)
		}
	}

	return result, nil
}

func New() *Storage {
	return &Storage{
		events: make(Events, 0),
	}
}
