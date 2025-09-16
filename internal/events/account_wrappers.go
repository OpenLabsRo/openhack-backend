package events

import (
	"backend/internal/models"
)

func (e *Emitter) AccountInitialized(
	superuserID, accountID string,
) {
	evt := models.Event{
		Action: "account.initialized",

		ActorRole: ActorParticipant,
		ActorID:   superuserID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "account.initialized|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) AccountRegisterSuccess(
	accountID string,
) {
	evt := models.Event{
		Action: "account.register",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,

		Key: "account.registerd|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) AccountRegisterFailure(
	accountID string,
	reason string,
) {
	evt := models.Event{
		Action: "account.register.failure",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: map[string]any{
			"reason": reason,
		},

		Key: "account.registerd|" + accountID,
	}

	e.Emit(evt)
}

func (e *Emitter) AccountLoginSuccess(
	accountID string,
) {
	evt := models.Event{
		Action: "account.login.success",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetParticipant,
		TargetID:   accountID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) AccountLoginFailure(
	accountID string,
	reason string,
) {
	evt := models.Event{
		Action: "account.login.failure",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetParticipant,
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
	evt := models.Event{
		Action: "account.login.success",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetParticipant,
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
	evt := models.Event{
		Action: "account.team.join",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) AccountTeamExit(
	accountID string,
	teamID string,
) {
	evt := models.Event{
		Action: "account.team.exit",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}
