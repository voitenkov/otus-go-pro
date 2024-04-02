package storage

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

type (
	EventTime time.Time
	EventDate time.Time
)

type Event struct {
	ID           uuid.UUID // Уникальный идентификатор события
	UserID       uuid.UUID // ID пользователя, владельца события
	Title        string    // Короткий текст
	Description  string    // Описание события - длинный текст, опционально
	StartTime    EventTime // Дата и время начала события
	FinishTime   EventTime // Дата и время окончания события
	NotifyBefore int       // За сколько времени (минуты) высылать уведомление, опционально
}

func (e Event) MarshalJSON() ([]byte, error) {
	var tmp struct {
		ID           string
		UserID       string
		Title        string
		Description  string
		StartTime    string
		FinishTime   string
		NotifyBefore int
	}

	tmp.ID = e.ID.String()
	tmp.UserID = e.UserID.String()
	tmp.Title = e.Title
	tmp.Description = e.Description
	tmp.StartTime = time.Time(e.StartTime).Format(time.DateTime)
	tmp.FinishTime = time.Time(e.FinishTime).Format(time.DateTime)
	tmp.NotifyBefore = e.NotifyBefore
	json, err := json.Marshal(tmp)
	return json, err
}

func (e *Event) UnmarshalJSON(data []byte) (err error) {
	var startTime, finishTime time.Time
	var tmp struct {
		ID           string
		UserID       string
		Title        string
		Description  string
		StartTime    string
		FinishTime   string
		NotifyBefore int
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	e.ID, err = uuid.FromString(tmp.ID)
	if err != nil {
		return err
	}

	e.Title = tmp.Title
	e.Description = tmp.Description
	e.NotifyBefore = tmp.NotifyBefore
	startTime, err = time.Parse(time.DateTime, tmp.StartTime)
	if err != nil {
		return err
	}

	e.StartTime = EventTime(startTime)
	finishTime, err = time.Parse(time.DateTime, tmp.FinishTime)
	if err != nil {
		return err
	}

	e.FinishTime = EventTime(finishTime)
	e.NotifyBefore = tmp.NotifyBefore
	return err
}
