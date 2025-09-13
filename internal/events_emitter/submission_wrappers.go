package events_emitter

import "backend/internal/models"

func (e *Emitter) SubmissionModifyName(
	accountID, teamID string,
	oldName, newName string,
) {
	evt := &models.Event{
		Action: "submission.change.name",

		ActorRole: "parcitipant",
		ActorID:   accountID,

		TargetType: "submission",
		TargetID:   teamID,

		Props: map[string]any{
			"oldName": oldName,
			"newName": newName,
		},
	}

	e.EmitWindowed(evt)
}
func (e *Emitter) SubmissionModifyDesc(
	accountID, teamID string,
	oldDesc, newDesc string,
) {
	evt := &models.Event{
		Action: "submission.change.desc",

		ActorRole: "parcitipant",
		ActorID:   accountID,

		TargetType: "submission",
		TargetID:   teamID,

		Props: map[string]any{
			"oldDesc": oldDesc,
			"newDesc": newDesc,
		},
	}

	e.EmitWindowed(evt)
}
func (e *Emitter) SubmissionModifyLink(
	accountID, teamID string,
	oldLink, newLink string,
) {
	evt := &models.Event{
		Action: "submission.change.link",

		ActorRole: "parcitipant",
		ActorID:   accountID,

		TargetType: "submission",
		TargetID:   teamID,

		Props: map[string]any{
			"oldLink": oldLink,
			"newLink": newLink,
		},
	}

	e.EmitWindowed(evt)
}
