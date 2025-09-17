package events

import "backend/internal/models"

func (e *Emitter) CheckInScan(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "checkin.scan",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "checkin.scan|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) CheckInTagAssign(
	superuserID, accountID, tagID string,
) {
	evt := models.Event{
		Action: "checkin.tag.assigned",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: map[string]any{
			"tagID": tagID,
		},

		Key: "checkin.tag.assigned|" + accountID,
	}

	e.Emit(evt)
}
