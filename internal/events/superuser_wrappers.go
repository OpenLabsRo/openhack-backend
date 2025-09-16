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
