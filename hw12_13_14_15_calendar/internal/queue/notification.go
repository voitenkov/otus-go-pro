package queue

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"

	"github.com/voitenkov/otus-go-pro/hw12_13_14_15_calendar/internal/storage"
)

type Notification struct {
	ID        uuid.UUID         // Уникальный идентификатор события
	UserID    uuid.UUID         // ID пользователя, владельца события
	Title     string            // Короткий текст
	StartTime storage.EventTime // Дата и время начала события
}

func (e Notification) MarshalJSON() ([]byte, error) {
	var tmp struct {
		ID        string
		UserID    string
		Title     string
		StartTime string
	}

	tmp.ID = e.ID.String()
	tmp.UserID = e.UserID.String()
	tmp.Title = e.Title
	tmp.StartTime = time.Time(e.StartTime).Format(time.DateTime)
	json, err := json.Marshal(tmp)
	return json, err
}

func (e *Notification) UnmarshalJSON(data []byte) (err error) {
	var startTime time.Time
	var tmp struct {
		ID        string
		UserID    string
		Title     string
		StartTime string
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	e.ID, err = uuid.FromString(tmp.ID)
	if err != nil {
		return err
	}

	e.UserID, err = uuid.FromString(tmp.UserID)
	if err != nil {
		return err
	}

	e.Title = tmp.Title
	startTime, err = time.Parse(time.DateTime, tmp.StartTime)
	if err != nil {
		return err
	}

	e.StartTime = storage.EventTime(startTime)
	if err != nil {
		return err
	}

	return err
}
