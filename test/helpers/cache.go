package helpers

import (
	"backend/internal/db"
)

func ResetTestCache() {
	if db.RDB == nil {
		return
	}
	_ = db.RDB.FlushDB(db.Ctx).Err()
}
