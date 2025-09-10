package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func flagStagesGetHandler(c fiber.Ctx) error {
	flagStages, err := models.GetFlagStages()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError)
	}

	return c.JSON(flagStages)
}

func flagStagesCreateHandler(c fiber.Ctx) error {
	flagStage := models.FlagStage{}
	json.Unmarshal(c.Body(), &flagStage)

	err := flagStage.Create()

	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError)
	}

	return c.JSON(flagStage)
}

func flagStagesDeleteHandler(c fiber.Ctx) error {
	flagStage := models.FlagStage{ID: c.Query("id")}

	err := flagStage.Delete()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError)
	}

	return c.JSON(flagStage)
}

func flagStagesExecuteHandler(c fiber.Ctx) error {
	flagStage := models.FlagStage{ID: c.Query("id")}
	serr := flagStage.Get()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	err := flagStage.Execute()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError)
	}

	flags := models.Flags{}
	err = flags.Get()
	if err != nil {
		return utils.StatusError(c, errmsg.InternalServerError)
	}

	return c.JSON(flags)

}
