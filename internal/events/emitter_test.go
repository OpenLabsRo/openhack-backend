package events

import (
	"backend/internal/models"
	"context"
	"sync"
	"testing"
	"time"
)

func sampleEvt(i int) models.Event {
	return models.Event{
		TimeStamp: time.Now().UTC(),
		Action:    "test",

		ActorID:   "u",
		ActorRole: "participant",

		TargetID:   "t",
		TargetType: "team",

		Props: map[string]any{
			"i": i,
		},
	}
}

func TestFlushOnBatchSize(t *testing.T) {
	em := NewEmitter(
		nil,
		Config{
			Buffer:     10,
			BatchSize:  3,
			FlushEvery: time.Hour,
		},
	)
	defer em.Close()

	var mu sync.Mutex
	var many [][]any

	em.insertMany = func(_ context.Context, docs []any) error {
		mu.Lock()
		defer mu.Unlock()

		cp := append([]any{}, docs...)
		many = append(many, cp)

		return nil
	}
	em.insertOne = func(_ context.Context, doc any) error {
		return nil
	}

	em.buf <- sampleEvt(1)
	em.buf <- sampleEvt(2)
	em.buf <- sampleEvt(3)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(many) != 1 || len(many[0]) != 3 {
		t.Fatalf("expected 1 InsertMany flush of 3, got %+v", many)
	}
}

func TestFlushOnTimer(t *testing.T) {
	em := NewEmitter(
		nil,
		Config{
			Buffer:     10,
			BatchSize:  100,
			FlushEvery: 50 * time.Millisecond,
		},
	)
	defer em.Close()

	var mu sync.Mutex
	var many [][]any
	em.insertMany = func(_ context.Context, docs []any) error {
		mu.Lock()
		defer mu.Unlock()

		cp := append([]any{}, docs...)
		many = append(many, cp)

		return nil
	}
	em.insertOne = func(_ context.Context, doc any) error {
		return nil
	}

	em.buf <- sampleEvt(1)

	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(many) != 1 || len(many[0]) != 1 {
		t.Fatalf("expected timer flush of 1, got %+v", many)
	}
}

func TestFallbackInsertOneWhenBufferFull(t *testing.T) {
	em := newEmitterNoWorker()
	defer em.Close()

	var mu sync.Mutex
	var ones []any
	em.insertMany = func(_ context.Context, docs []any) error {
		return nil
	}
	em.insertOne = func(_ context.Context, doc any) error {
		mu.Lock()
		defer mu.Unlock()
		ones = append(ones, doc)
		return nil
	}

	// occupy channel
	em.buf <- sampleEvt(1)
	// force fallback path
	evt := sampleEvt(2)
	em.Emit(evt)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(ones) != 1 {
		t.Fatalf("expected 1 InsertOne fallback, got %d", len(ones))
	}
}

func TestCloseFlushesRemaining(t *testing.T) {
	em := NewEmitter(
		nil,
		Config{
			Buffer:     10,
			BatchSize:  10,
			FlushEvery: time.Hour,
		},
	)

	var mu sync.Mutex
	var many [][]any
	em.insertMany = func(_ context.Context, docs []any) error {
		mu.Lock()
		defer mu.Unlock()
		cp := append([]any{}, docs...)
		many = append(many, cp)
		return nil
	}
	em.insertOne = func(_ context.Context, doc any) error { return nil }

	// enqueue fewer than batch size so only Close() flushes them
	for i := range 4 {
		em.buf <- sampleEvt(i)
	}

	em.Close()

	mu.Lock()
	defer mu.Unlock()
	if len(many) != 1 || len(many[0]) != 4 {
		t.Fatalf(
			"expected 1 final flush of 4, got many=%d size=%d",
			len(many),
			len(many[0]),
		)
	}
}
