package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEvent_NewEvent проверяет правильность создания объекта.
func TestEvent_NewEvent(t *testing.T) {
	var (
		userID1           = "user-123"
		userID2           = "user-456"
		nilUserID *string = nil
	)

	tests := []struct {
		name   string
		action ActionType
		userID *string
		url    string
	}{
		{
			name:   "event with user ID and action",
			action: Follow,
			userID: &userID1,
			url:    "http://example.com/1",
		},
		{
			name:   "valid event with another user ID and action",
			action: Shorten,
			userID: &userID2,
			url:    "http://example.com/2",
		},
		{
			name:   "event with nil user ID",
			action: Follow,
			userID: nilUserID,
			url:    "http://example.com/3",
		},
		{
			name:   "event with empty URL",
			action: Follow,
			userID: nilUserID,
			url:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(tt.action, tt.userID, tt.url)
			require.NotNil(t, event)

			assert.Positive(t, event.TS)

			assert.Equal(t, tt.action, event.Action)

			if tt.userID == nil {
				assert.Nil(t, event.UserID)
			} else {
				require.NotNil(t, event.UserID)
				assert.Equal(t, *tt.userID, *event.UserID)
			}

			assert.Equal(t, tt.url, event.URL)
		})
	}
}
