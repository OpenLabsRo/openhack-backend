package events

import "backend/internal/models"

func (e *Emitter) TeamCreate(
	accountID, teamID string,
) {
	evt := models.Event{
		Action: "team.created",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: nil,

		Key: "team.created|" + teamID,
	}

	e.Emit(evt)
}

func (e *Emitter) TeamNameChange(
	accountID, teamID, oldName, newName string,
) {
	evt := models.Event{
		Action: "team.name.change",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: map[string]any{
			"oldName": oldName,
			"newName": newName,
		},

		Key: "team.name.change|" + teamID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) TeamTableChange(
	accountID, teamID, oldTable, newTable string,
) {
	evt := models.Event{
		Action: "team.table.change",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: map[string]any{
			"oldTable": oldTable,
			"newTable": newTable,
		},

		Key: "team.table.change|" + teamID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) TeamMemberJoin(
	accountID, teamID string,
) {
	evt := models.Event{
		Action: "team.members.add",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: nil,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) TeamMemberLeave(
	accountID, teamID string,
) {
	evt := models.Event{
		Action: "team.members.exit",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: nil,

		Key: "team.members.exit|" + teamID + "|" + accountID,
	}
	e.EmitWindowed(evt)
}

func (e *Emitter) TeamMemberKick(
	actorID, teamID string,
	kickedID string,
) {
	evt := models.Event{
		Action: "team.members.exit",

		ActorRole: ActorParticipant,
		ActorID:   actorID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: map[string]any{
			"kickedID": kickedID,
		},

		Key: "team.members.exit|" + teamID + "|" + kickedID,
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) TeamDelete(
	accountID, teamID string,
) {
	evt := models.Event{
		Action: "team.deleted",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: nil,

		Key: "team.deleted|" + teamID,
	}

	e.Emit(evt)
}
