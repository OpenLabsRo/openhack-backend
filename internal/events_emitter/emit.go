package events_emitter

import (
	"backend/internal/models"
	"context"
	"time"
)

func (e *Emitter) Emit(evt *models.Event) {
	evt.TimeStamp = time.Now().UTC()
	select {
	case e.buf <- *evt:
	default:
		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer cancel()
		_ = e.insertOne(ctx, evt)
	}
}

func (e *Emitter) EmitWindowed(evt *models.Event) {
	evt.TimeStamp = time.Now().UTC()
	evt.Key = WindowedKey(
		evt,
		3*time.Second,
	)
	select {
	case e.buf <- *evt:
	default:
		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer cancel()
		_ = e.insertOne(ctx, evt)
	}
}
