package events

import "backend/internal/models"

func (e *Emitter) SuperUserLogin(
	superuserID string,
) {
	evt := models.Event{
		Action: "superuser.login",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: "superuser",
		TargetID:   superuserID,

		Props: nil,
	}

	e.Emit(evt)
}

func (e *Emitter) SuperUserFlagChange(
	superuserID string,
	flags map[string]bool,
) {
	evt := models.Event{
		Action: "superuser.flags.change",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: "flags",
		TargetID:   "flags",

		Props: map[string]any{
			"flags": flags,
		},
	}

	e.Emit(evt)
}

func (e *Emitter) SuperUserSettingChange(
	superuserID string,
	setting string,
	oldValue string,
	value string,
) {
	evt := models.Event{
		Action: "superuser.setting.change",

		ActorRole: ActorSuperUser,
		ActorID:   superuserID,

		TargetType: "setting",
		TargetID:   "setting",

		Props: map[string]any{
			"oldValue": oldValue,
			"newValue": value,
		},
	}

	e.Emit(evt)
}
