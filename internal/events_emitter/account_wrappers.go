package events_emitter

import (
	"backend/internal/models"
)

func (e *Emitter) AccountInitialized(
	superuserID, accountID string,
) {
	evt := &models.Event{
		Action: "account.initialized",

		ActorRole: "superuser",
		ActorID:   superuserID,

		TargetType: "account",
		TargetID:   accountID,

		Props: nil,

		Key: "account.initialized|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) AccountRegistered(
	accountID string,
) {
	evt := &models.Event{
		Action: "account.registered",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "account",
		TargetID:   accountID,

		Props: nil,

		Key: "account.registerd|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) AccountLoginSuccess(
	accountID string,
) {
	evt := &models.Event{
		Action: "account.login.success",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "account",
		TargetID:   accountID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) AccountLoginFailure(
	accountID string,
	reason string,
) {
	evt := &models.Event{
		Action: "account.login.failure",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "account",
		TargetID:   accountID,

		Props: map[string]any{
			"reason": reason,
		},
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) AccountNameChanged(
	accountID string,
	oldName string,
	newName string,
) {
	evt := &models.Event{
		Action: "account.login.success",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "account",
		TargetID:   accountID,

		Props: map[string]any{
			"oldName": oldName,
			"newName": newName,
		},
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) AccountTeamJoin(
	accountID string,
	teamID string,
) {
	evt := &models.Event{
		Action: "account.team.join",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) AccountTeamExit(
	accountID string,
	teamID string,
) {
	evt := &models.Event{
		Action: "account.team.exit",

		ActorRole: "parcitipant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}
