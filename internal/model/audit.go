package model

import "time"

// ActionType представляет тип действия, выполняемого в рамках аудита.
type ActionType string

const (
	// Follow представляет действие перехода по короткому URL.
	Follow ActionType = "follow"
	// Shorten представляет действие создания короткого URL.
	Shorten ActionType = "shorten"
)

// Event описывает событие аудита, содержащее временную метку,
// тип действия, идентификатор пользователя и связанный URL.
type Event struct {
	TS     int64      `json:"ts"`
	Action ActionType `json:"action"`
	UserID *string    `json:"user_id,omitempty"`
	URL    string     `json:"url"`
}

// NewEvent создает и возвращает новое событие аудита с текущей временной меткой,
// заданным типом действия, идентификатором пользователя и URL.
func NewEvent(action ActionType, userID *string, url string) *Event {
	return &Event{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
