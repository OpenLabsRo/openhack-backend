package events

import (
	"backend/internal/models"
	"backend/test/common"
	"context"
	"time"
)

var (
	ActorParticipant = "participant"
	ActorSuperUser   = "superuser"
	ActorJudge       = "judge"
)

var (
	TargetParticipant = "participant"
	TargetTeam        = "team"
	TargetSubmission  = "submission"
)

func (e *Emitter) Emit(evt models.Event) {
	if e.deployment == "test" {
		common.IncrementEventCounter()
	}

	loc, err := time.LoadLocation("Europe/Bucharest")
	if err != nil {
		panic(err)
	}
	evt.TimeStamp = time.Now().In(loc)
	select {
	case e.buf <- evt:
	default:
		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer cancel()
		_ = e.InsertOne(ctx, evt)
	}
}

func (e *Emitter) EmitWindowed(evt models.Event) {
	if e.deployment == "test" {
		common.IncrementEventCounter()
	}

	loc, err := time.LoadLocation("Europe/Bucharest")
	if err != nil {
		panic(err)
	}
	evt.TimeStamp = time.Now().In(loc)
	evt.Key = WindowedKey(
		&evt,
		3*time.Second,
	)
	select {
	case e.buf <- evt:
	default:
		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		defer cancel()
		_ = e.InsertOne(ctx, evt)
	}
}


