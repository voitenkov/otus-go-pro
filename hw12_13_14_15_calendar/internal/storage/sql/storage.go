package sqlstorage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	_ "github.com/jackc/pgx/stdlib" // no lint
	"github.com/jmoiron/sqlx"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/config"
	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	config config.Config
	dsn    string
	db     *sqlx.DB
}

func (s *Storage) Connect() error {
	var err error
	s.db, err = sqlx.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	// s.db = db
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	query := `insert into events(id, user_id, title, description, start_time, finish_time, notify_before, 
		        notification_sent) 
	          values($1, $2, $3, $4, $5, $6, $7, $8)`
	// strconv.FormatInt(int64(time.Duration(event.Duration)), 10),
	_, err := s.db.ExecContext(
		ctx,
		query,
		event.ID.String(),
		event.UserID.String(),
		event.Title,
		event.Description,
		time.Time(event.StartTime).Format(time.RFC3339),
		time.Time(event.FinishTime).Format(time.RFC3339),
		strconv.Itoa(event.NotifyBefore),
		false)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	query := `update
			    events
			  set
			    user_id = $2,
				title = $3,
				description = $4,
				start_time = $5,
				finish_time = $6,
				notify_before = $7,
				notification_sent = $8
			  where
			    id = $1`

	_, err := s.db.ExecContext(
		ctx,
		query,
		event.ID.String(),
		event.UserID.String(),
		event.Title,
		event.Description,
		time.Time(event.StartTime).Format(time.RFC3339),
		time.Time(event.FinishTime).Format(time.RFC3339),
		event.NotifyBefore,
		event.NotificationSent)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) PatchEvent(ctx context.Context, id uuid.UUID, userID *uuid.UUID, title, description *string,
	startTime, finishTime *storage.EventTime, notifyBefore *int, notificationSent *bool,
) error {
	var event storage.Event
	query := `select
		id,
		user_id,
		title,
		description,
		start_time,
		finish_time,
		notify_before,
		notification_sent
	  from
		events
	  where
		  id = $1`
	rows, err := s.db.QueryxContext(ctx, query, id)
	if err != nil {
		return err
	}

	defer rows.Close()

	err = rows.Scan(&event.ID, &event.UserID, &event.Title, &event.Description, &event.StartTime,
		&event.FinishTime, &event.NotifyBefore, &event.NotificationSent)
	if err != nil {
		return err
	}

	if userID != nil {
		event.UserID = *userID
	}

	if title != nil {
		event.Title = *title
	}

	if description != nil {
		event.Description = *description
	}

	if startTime != nil {
		event.StartTime = *startTime
	}

	if finishTime != nil {
		event.FinishTime = *finishTime
	}

	if notifyBefore != nil {
		event.NotifyBefore = *notifyBefore
	}

	if notificationSent != nil {
		event.NotificationSent = *notificationSent
	}

	query = `update
			events
		set
			user_id = $2,
			title = $3,
			description = $4,
			start_time = $5,
			finish_time = $6,
			notify_before = $7,
			notification_sent = $8
		where
			id = $1`

	_, err = s.db.ExecContext(
		ctx,
		query,
		event.ID.String(),
		event.UserID.String(),
		event.Title,
		event.Description,
		time.Time(event.StartTime).Format(time.RFC3339),
		time.Time(event.FinishTime).Format(time.RFC3339),
		event.NotifyBefore,
		event.NotificationSent)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	query := "delete from events where id = $1"
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ListEventsByDate(ctx context.Context, userID uuid.UUID,
	startDate storage.EventDate,
) ([]storage.Event, error) {
	return s.ListEventsByPeriod(ctx, userID, startDate, storage.EventDate(time.Time(startDate).AddDate(0, 0, 1)))
}

func (s *Storage) ListEventsByWeek(ctx context.Context, userID uuid.UUID,
	startDate storage.EventDate,
) ([]storage.Event, error) {
	return s.ListEventsByPeriod(ctx, userID, startDate, storage.EventDate(time.Time(startDate).AddDate(0, 0, 7)))
}

func (s *Storage) ListEventsByMonth(ctx context.Context, userID uuid.UUID,
	startDate storage.EventDate,
) ([]storage.Event, error) {
	return s.ListEventsByPeriod(ctx, userID, startDate, storage.EventDate(time.Time(startDate).AddDate(0, 1, 0)))
}

func (s *Storage) ListEventsByPeriod(ctx context.Context, userID uuid.UUID, startDate,
	finishDate storage.EventDate,
) ([]storage.Event, error) {
	result := make([]storage.Event, 0)

	query := `select
				id,
				user_id,
				title,
				description,
				start_time,
				finish_time,
				notify_before,
				notification_sent
			  from
			    events
			  where
			  	user_id = $1 and start_time < $3 and finish_time > $2`
	rows, err := s.db.QueryxContext(ctx, query, userID, time.Time(startDate).Format(time.RFC3339),
		time.Time(finishDate).Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event storage.Event
		err := rows.Scan(&event.ID, &event.UserID, &event.Title, &event.Description, &event.StartTime,
			&event.FinishTime, &event.NotifyBefore, &event.NotificationSent)
		if err != nil {
			return nil, err
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Storage) SelectEventsToNotify(ctx context.Context) ([]storage.Event, error) {
	result := make([]storage.Event, 0)

	query := `select
				id,
				user_id,
				title,
				description,
				start_time,
				finish_time,
				notify_before,
				notification_sent
			  from
			    events
			  where
			    notification_sent is not true and start_time <= now() + interval '1 minute' * notify_before`
	rows, err := s.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event storage.Event
		err := rows.Scan(&event.ID, &event.UserID, &event.Title, &event.Description, &event.StartTime,
			&event.FinishTime, &event.NotifyBefore, &event.NotificationSent)
		if err != nil {
			return nil, err
		}

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Storage) PurgeEvents(ctx context.Context, purgeIntervalDays int) (purgedEvents int64, err error) {
	query := "delete from events where finish_time < now() - interval '1 day' * $1"
	result, err := s.db.ExecContext(ctx, query, purgeIntervalDays)
	if err != nil {
		return 0, err
	}

	purgedEvents, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return purgedEvents, nil
}

func New(config *config.Config, dsn string) *Storage {
	return &Storage{
		config: *config,
		dsn:    dsn,
	}
}
