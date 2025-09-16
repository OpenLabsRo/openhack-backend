package events

import (
	"backend/internal/models"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func WindowedKey(
	evt *models.Event,
	window time.Duration,
) string {
	// the bucket is window in which this event can happen
	// for idempotency purposes
	bucket := evt.TimeStamp.UnixNano() / int64(window)

	// initializing the hash
	h := sha256.New()

	// writing the event information
	h.Write([]byte(evt.Action))
	h.Write([]byte("|"))
	h.Write([]byte(evt.ActorID))
	h.Write([]byte("|"))
	h.Write([]byte(evt.TargetType))
	h.Write([]byte("|"))
	h.Write([]byte(evt.TargetID))
	h.Write([]byte("|"))

	// include common props if present
	if evt.Props != nil {
		if v, ok := evt.Props["new"].(string); ok {
			h.Write([]byte(v))
			h.Write([]byte("|"))
		}
	}

	fmt.Fprintf(h, "%d", bucket)

	return hex.EncodeToString(
		h.Sum(nil),
	)
}
