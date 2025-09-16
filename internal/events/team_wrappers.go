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

func (e *Emitter) TeamChangeName(
	accountID, teamID, oldName, newName string,
) {
	evt := models.Event{
		Action: "team.change.name",

		ActorRole: "participant",
		ActorID:   accountID,

		TargetType: "team",
		TargetID:   teamID,

		Props: map[string]any{
			"oldName": oldName,
			"newName": newName,
		},
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) TeamMemberJoin(
	accountID, teamID string,
) {
	evt := models.Event{
		Action: "team.member.add",

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
		Action: "team.member.exit",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: nil,
	}
	e.EmitWindowed(evt)
}

func (e *Emitter) TeamMemberKick(
	actorID, teamID string,
	kickedID string,
) {
	evt := models.Event{
		Action: "team.member.exit",

		ActorRole: ActorParticipant,
		ActorID:   actorID,

		TargetType: TargetTeam,
		TargetID:   teamID,

		Props: map[string]any{
			"kickedID": kickedID,
		},
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

		Key: "team.created|" + teamID,
	}

	e.Emit(evt)
}
