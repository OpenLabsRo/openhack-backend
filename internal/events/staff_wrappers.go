package events

import "backend/internal/models"

func (e *Emitter) StaffRegister(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "staff.register",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "checkin.scan|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) StaffTagAssign(
	superuserID, accountID, tagID string,
) {
	evt := models.Event{
		Action: "staff.tag.assigned",

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

func (e *Emitter) StaffCheckIn(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "staff.checkin",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "checkin|" + accountID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) StaffCheckOut(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "staff.checkout",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "checkout|" + accountID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) StaffConsumablesUpdated(
	superuserID, accountID string,
	consumables models.Consumables,
) {
	evt := models.Event{
		Action: "staff.consumables.updated",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: map[string]any{
			"consumables": consumables,
		},

		Key: "consumables.updated|" + accountID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) StaffPresentIn(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "staff.present.in",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "present.in|" + accountID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) StaffPresentOut(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "staff.present.out",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "present.out|" + accountID,
	}

	e.EmitWindowed(evt)
}
