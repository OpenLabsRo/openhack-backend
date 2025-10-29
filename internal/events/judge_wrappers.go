package events

import "backend/internal/models"

func (e *Emitter) JudgeConnectTokenIssued(
	superuserID string,
	judgeID string,
) {
	evt := models.Event{
		Action: "judge.connect.token.issued",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetJudge,
		TargetID:   judgeID,

		Props: nil,
	}

	e.Emit(evt)
}

func (e *Emitter) JudgeTokenUpgraded(
	judgeID string,
) {
	evt := models.Event{
		Action: "judge.token.upgraded",

		ActorRole: ActorJudge,
		ActorID:   judgeID,

		TargetType: TargetJudge,
		TargetID:   judgeID,

		Props: nil,
	}

	e.Emit(evt)
}

func (e *Emitter) JudgeInitTeamOrderSet(
	superuserID string,
	teamOrder []string,
) {
	evt := models.Event{
		Action: "judge.init.team.order.set",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: "judging",
		TargetID:   "judging",

		Props: map[string]any{
			"teamOrder": teamOrder,
		},
	}

	e.Emit(evt)
}

func (e *Emitter) JudgeInitJudgeOrderSet(
	superuserID string,
	judgeOrder []string,
) {
	evt := models.Event{
		Action: "judge.init.judge.order.set",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: "judging",
		TargetID:   "judging",

		Props: map[string]any{
			"judgeOrder": judgeOrder,
		},
	}

	e.Emit(evt)
}

func (e *Emitter) JudgeInitOffsetSet(
	superuserID string,
	judgeOffset []int,
	numTeams int,
) {
	evt := models.Event{
		Action: "judge.init.offset.set",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: "judging",
		TargetID:   "judging",

		Props: map[string]any{
			"judgeOffset": judgeOffset,
			"numTeams":    numTeams,
		},
	}

	e.Emit(evt)
}

func (e *Emitter) JudgeNextTeamRequested(
	judgeID string,
	teamID string,
) {
	evt := models.Event{
		Action: "judge.next.team.requested",

		ActorRole: ActorJudge,
		ActorID:   judgeID,

		TargetType: TargetJudge,
		TargetID:   judgeID,

		Props: map[string]any{
			"teamID": teamID,
		},
	}

	e.Emit(evt)
}

func (e *Emitter) JudgeCreated(
	superuserID string,
	judgeID string,
	judgeName string,
) {
	evt := models.Event{
		Action: "judge.created",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: TargetJudge,
		TargetID:   judgeID,

		Props: map[string]any{
			"name": judgeName,
		},
	}

	e.Emit(evt)
}

func (e *Emitter) JudgmentCreated(
	judgeID string,
	winningTeamID string,
	losingTeamID string,
) {
	evt := models.Event{
		Action: "judgment.created",

		ActorRole: ActorJudge,
		ActorID:   judgeID,

		TargetType: "judgment",
		TargetID:   "judgment",

		Props: map[string]any{
			"winningTeamID": winningTeamID,
			"losingTeamID":  losingTeamID,
		},
	}

	e.Emit(evt)
}
