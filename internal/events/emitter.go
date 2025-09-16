package events

import (
	"backend/internal/models"
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var Em *Emitter

type Config struct {
	Buffer     int // channel capacity
	BatchSize  int // InsertMany size
	FlushEvery time.Duration
}

type Emitter struct {
	col *mongo.Collection

	// hooks for testing
	insertMany func(context.Context, []any) error
	insertOne  func(context.Context, any) error

	buf        chan models.Event
	batchSize  int
	flushEvery time.Duration

	wg        sync.WaitGroup
	onceClose sync.Once
}

func NewEmitter(col *mongo.Collection, cfg Config) *Emitter {
	if cfg.Buffer < 10 {
		cfg.Buffer = 10
	}
	if cfg.BatchSize < 1 {
		cfg.BatchSize = 1
	}
	if cfg.FlushEvery <= 0 {
		cfg.FlushEvery = 2 * time.Second
	}

	e := &Emitter{
		col:        col,
		buf:        make(chan models.Event, cfg.Buffer),
		batchSize:  cfg.BatchSize,
		flushEvery: cfg.FlushEvery,
	}

	e.insertMany = func(ctx context.Context, docs []interface{}) error {
		_, err := col.InsertMany(ctx, docs)
		return err
	}

	e.insertOne = func(ctx context.Context, doc interface{}) error {
		_, err := col.InsertOne(ctx, doc)
		return err
	}

	e.wg.Add(1)
	go e.worker()
	return e
}

func newEmitterNoWorker() *Emitter {
	em := &Emitter{
		buf:        make(chan models.Event, 1),
		batchSize:  100,
		flushEvery: time.Hour,
	}

	return em
}

func (e *Emitter) Close() {
	e.onceClose.Do(func() {
		close(e.buf)
		e.wg.Wait()
	})
}

func (e *Emitter) worker() {
	defer e.wg.Done()

	batch := make([]interface{}, 0, e.batchSize)
	timer := time.NewTimer(e.flushEvery)

	defer timer.Stop()

	flush := func() {
		// if there's nothing in the back
		if len(batch) == 0 {
			timer.Reset(e.flushEvery) // just reset the timer
			return                    // and return
		}

		// creating a background context
		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second, // with a max 2 second timeout
		)

		// push to the db
		_ = e.insertMany(ctx, batch)

		// cancel the timer
		// after succeeding
		cancel()

		batch = batch[:0]         // resetting the batch
		timer.Reset(e.flushEvery) // reset the timer
	}

	for {
		select { // apparently,
		// select will do one of the actions on a channel action
		case evt, ok := <-e.buf: // pushing to the channel
			// if it's not ok
			// meaning we have exited the Emitter
			// and the server in effect
			// we flush
			if !ok {
				flush()
				return
			}

			// if it is ok,
			// we append to the batch
			batch = append(batch, evt)

			// but if the batch is full
			if len(batch) >= e.batchSize {
				flush() // we flush
			}
		case <-timer.C:
			flush()
		}
	}
}
