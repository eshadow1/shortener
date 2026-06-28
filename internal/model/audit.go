package model

import "time"

type ActionType string

const (
	Follow  ActionType = "follow"
	Shorten ActionType = "shorten"
)

type Event struct {
	TS     int64      `json:"ts"`
	Action ActionType `json:"action"`
	UserID *string    `json:"user_id,omitempty"`
	URL    string     `json:"url"`
}

func NewEvent(action ActionType, userID *string, url string) *Event {
	return &Event{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	}
}
