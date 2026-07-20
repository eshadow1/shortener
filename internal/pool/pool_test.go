package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testObj — тестовая структура, реализующая интерфейс с методом Reset().
type testObj struct {
	value      int
	resetCalls int
}

func (o *testObj) Reset() {
	o.value = 0
	o.resetCalls++
}

func TestPool_Get(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(p *Pool[*testObj])
		wantValue        int
		wantCreatorCalls int
	}{
		{
			name:             "empty Pool returns value",
			setup:            func(*Pool[*testObj]) {},
			wantValue:        999,
			wantCreatorCalls: 1,
		},
		{
			name: "pool with one object returns it without calling factory",
			setup: func(p *Pool[*testObj]) {
				p.Put(&testObj{value: 10})
			},
			wantValue:        0,
			wantCreatorCalls: 0,
		},
		{
			name: "pool with multiple objects returns one without calling factory",
			setup: func(p *Pool[*testObj]) {
				p.Put(&testObj{value: 10})
				p.Put(&testObj{value: 20})
				p.Put(&testObj{value: 30})
			},
			wantValue:        0,
			wantCreatorCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creatorCalls := 0
			creator := func() *testObj {
				creatorCalls++
				return &testObj{value: 999}
			}

			pool := NewPool(creator)
			tt.setup(pool)

			got := pool.Get()

			require.NotNil(t, got)
			assert.Equal(t, tt.wantValue, got.value)
			assert.Equal(t, tt.wantCreatorCalls, creatorCalls)
		})
	}
}

func TestPool_Put(t *testing.T) {
	tests := []struct {
		name           string
		obj            *testObj
		wantResetCalls int
		wantValue      int
	}{
		{
			name:           "Put resets object state",
			obj:            &testObj{value: 100, resetCalls: 0},
			wantResetCalls: 1,
			wantValue:      0,
		},
		{
			name:           "Put resets already reset object",
			obj:            &testObj{value: 50, resetCalls: 3},
			wantResetCalls: 4,
			wantValue:      0,
		},
		{
			name:           "Put with zero value",
			obj:            &testObj{value: 0, resetCalls: 0},
			wantResetCalls: 1,
			wantValue:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(func() *testObj { return &testObj{} })

			pool.Put(tt.obj)

			assert.Equal(t, tt.wantResetCalls, tt.obj.resetCalls)
			assert.Equal(t, tt.wantValue, tt.obj.value)

			got := pool.Get()

			assert.Same(t, tt.obj, got)
		})
	}
}
