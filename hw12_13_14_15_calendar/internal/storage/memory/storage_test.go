package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage(t *testing.T) {
	id, _ := uuid.NewV4()
	userID, _ := uuid.NewV4()
	memstor := New()
	createdEvent := &storage.Event{
		ID:           id,
		UserID:       userID,
		Title:        "Meeting",
		StartTime:    storage.EventTime(time.Date(2024, time.January, 0o1, 10, 0, 0, 0, time.UTC)),
		FinishTime:   storage.EventTime(time.Date(2024, time.January, 0o1, 11, 0, 0, 0, time.UTC)),
		NotifyBefore: 60,
	}

	updatedEvent := &storage.Event{
		ID:           id,
		UserID:       userID,
		Title:        "Party",
		StartTime:    storage.EventTime(time.Date(2024, time.January, 0o1, 10, 0, 0, 0, time.UTC)),
		FinishTime:   storage.EventTime(time.Date(2024, time.January, 0o1, 11, 0, 0, 0, time.UTC)),
		NotifyBefore: 60,
	}

	ctx := context.Background()

	t.Run("create event", func(t *testing.T) {
		err := memstor.CreateEvent(ctx, *createdEvent)
		require.NoError(t, err)
		require.Equal(t, *createdEvent, memstor.events[id])
	})

	t.Run("event already exists", func(t *testing.T) {
		err := memstor.CreateEvent(ctx, *createdEvent)
		require.ErrorIs(t, err, errEventExists)
	})

	t.Run("update event", func(t *testing.T) {
		err := memstor.UpdateEvent(ctx, *updatedEvent)
		require.NoError(t, err)
		require.Equal(t, "Party", memstor.events[id].Title)
	})

	t.Run("list events by date", func(t *testing.T) {
		events, err := memstor.ListEventsByDate(ctx, userID, storage.EventDate(time.Date(2024, time.January,
			0o1, 0, 0, 0, 0, time.UTC)))
		require.NoError(t, err)
		require.Equal(t, *updatedEvent, events[0])
		require.NotEqual(t, *createdEvent, events[0])
	})

	t.Run("events not listed", func(t *testing.T) {
		events, err := memstor.ListEventsByWeek(ctx, userID, storage.EventDate(time.Date(2024, time.January,
			0o2, 0, 0, 0, 0, time.UTC)))
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})

	t.Run("delete event", func(t *testing.T) {
		err := memstor.DeleteEvent(ctx, id)
		require.NoError(t, err)
		_, found := memstor.events[id]
		require.False(t, found)
	})

	t.Run("update non-existed event", func(t *testing.T) {
		err := memstor.UpdateEvent(ctx, *updatedEvent)
		require.ErrorIs(t, err, errEventNotFound)
	})
}

func TestStorageConcurrent(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	memstor := New()
	userID, _ := uuid.NewV4()
	ctx := context.Background()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			id, _ := uuid.NewV4()
			testEvent := &storage.Event{
				ID:           id,
				UserID:       userID,
				Title:        "Meeting",
				StartTime:    storage.EventTime(time.Date(2024, time.January, 0o1, 10, 0, 0, 0, time.UTC)),
				FinishTime:   storage.EventTime(time.Date(2024, time.January, 0o1, 11, 0, 0, 0, time.UTC)),
				NotifyBefore: 60,
			}
			err := memstor.CreateEvent(ctx, *testEvent)
			require.Nil(t, err)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			id, _ := uuid.NewV4()
			testEvent := &storage.Event{
				ID:           id,
				UserID:       userID,
				Title:        "Meeting",
				StartTime:    storage.EventTime(time.Date(2024, time.January, 0o1, 10, 0, 0, 0, time.UTC)),
				FinishTime:   storage.EventTime(time.Date(2024, time.January, 0o1, 11, 0, 0, 0, time.UTC)),
				NotifyBefore: 60,
			}
			err := memstor.CreateEvent(ctx, *testEvent)
			require.Nil(t, err)
		}
	}()

	wg.Wait()
}
