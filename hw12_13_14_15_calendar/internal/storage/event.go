package storage

import (
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
