package service

import (
	"testing"

	"github.com/eshadow1/shortener/internal/model"
	mockservice "github.com/eshadow1/shortener/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditBroker_NewAuditBroker(t *testing.T) {
	broker := NewAuditBroker()
	require.NotNil(t, broker)
	broker.Close()
}

func TestAuditBroker_Register(t *testing.T) {
	mobs1 := mockservice.NewMockObserver(t)
	mobs1.On("Close").Return().Maybe()

	mobs2 := mockservice.NewMockObserver(t)
	mobs2.On("Close").Return().Maybe()

	tests := []struct {
		name      string
		observers []*mockservice.MockObserver
	}{
		{
			name:      "register one observer",
			observers: []*mockservice.MockObserver{mobs1},
		},
		{
			name:      "register multiple observers",
			observers: []*mockservice.MockObserver{mobs1, mobs2},
		},
		{
			name:      "register zero observer",
			observers: []*mockservice.MockObserver{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broker := NewAuditBroker()
			defer broker.Close()
			for _, obs := range tt.observers {
				assert.NotPanics(t, func() {
					broker.Register(obs)
				})
			}
			assert.Len(t, broker.observers, len(tt.observers))
		})
	}
}

func TestAuditBroker_Notify(t *testing.T) {
	broker := NewAuditBroker()
	defer broker.Close()

	event := model.Event{Action: model.Follow, URL: "http://example.com"}

	obs1 := mockservice.NewMockObserver(t)
	obs1.On("Close").Return().Maybe()
	obs1.On("Notify", event).Return().Maybe()

	obs2 := mockservice.NewMockObserver(t)
	obs2.On("Close").Return().Maybe()
	obs2.On("Notify", event).Return().Maybe()

	broker.Register(obs1)
	broker.Register(obs2)

	assert.NotPanics(t, func() {
		broker.Notify(event)
	})
}

func TestAuditBroker_RegisterAfterClose(t *testing.T) {
	broker := NewAuditBroker()
	broker.Close()

	obs := mockservice.NewMockObserver(t)
	obs.On("Close").Return().Maybe()

	assert.NotPanics(t, func() {
		broker.Register(obs)
	})
}
