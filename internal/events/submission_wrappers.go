package events

import "backend/internal/models"

func (e *Emitter) SubmissionChangeName(
	accountID, teamID string,
	oldName, newName string,
) {
	evt := models.Event{
		Action: "submission.name.change",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetSubmission,
		TargetID:   teamID,

		Props: map[string]any{
			"oldName": oldName,
			"newName": newName,
		},
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) SubmissionChangeDesc(
	accountID, teamID string,
	oldDesc, newDesc string,
) {
	evt := models.Event{
		Action: "submission.desc.change",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetSubmission,
		TargetID:   teamID,

		Props: map[string]any{
			"oldDesc": oldDesc,
			"newDesc": newDesc,
		},
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) SubmissionChangeRepo(
	accountID, teamID string,
	oldRepo, newRepo string,
) {
	evt := models.Event{
		Action: "submission.repo.change",

		ActorRole: ActorParticipant,
		ActorID:   accountID,

		TargetType: TargetSubmission,
		TargetID:   teamID,

		Props: map[string]any{
			"oldRepo": oldRepo,
			"newRepo": newRepo,
		},
	}

	e.EmitWindowed(evt)
}

func (e *Emitter) SubmissionChangePres(
	accountID, teamID string,
	oldRepo, newRepo string,
) {
	evt := models.Event{
		Action: "submission.repo.change",

		ActorRole: TargetParticipant,
		ActorID:   accountID,

		TargetType: TargetSubmission,
		TargetID:   teamID,

		Props: map[string]any{
			"oldRepo": oldRepo,
			"newRepo": newRepo,
		},
	}

	e.EmitWindowed(evt)
}
