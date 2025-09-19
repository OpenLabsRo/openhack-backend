package common

import "backend/internal/db"

func IncrementEventCounter() {
	db.RDB.Incr(db.Ctx, "event_counter")
}