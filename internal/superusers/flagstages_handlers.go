package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// flagStagesGetHandler returns the catalog of configured stages.
// @Summary List available flag stages
// @Description Lists named flag stage presets so admins can preview rollout scripts.
// @Tags Superusers Flag Stages
// @Security SuperUserAuth
// @Produce json
// @Success 200 {array} models.FlagStage
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/flags/stages [get]
func flagStagesGetHandler(c fiber.Ctx) error {
	flagStages, err := models.GetFlagStages()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flagStages)
}

// flagStagesCreateHandler stores a new stage configuration.
// @Summary Create a new flag stage
// @Description Saves a stage blueprint listing which flags to enable or disable.
// @Tags Superusers Flag Stages
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body FlagStageCreateRequest true "Flag stage"
// @Success 200 {object} models.FlagStage
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/flags/stages [post]
func flagStagesCreateHandler(c fiber.Ctx) error {
	flagStage := models.FlagStage{}
	json.Unmarshal(c.Body(), &flagStage)

	err := flagStage.Create()

	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flagStage)
}

// flagStagesDeleteHandler removes a stored stage by identifier.
// @Summary Delete an existing flag stage
// @Description Deletes an unused flag stage and returns the removed record.
// @Tags Superusers Flag Stages
// @Security SuperUserAuth
// @Produce json
// @Param id query string true "Flag stage ID"
// @Success 200 {object} models.FlagStage
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/flags/stages [delete]
func flagStagesDeleteHandler(c fiber.Ctx) error {
	flagStage := models.FlagStage{ID: c.Query("id")}

	err := flagStage.Delete()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flagStage)
}

// flagStagesExecuteHandler executes the toggles defined in a stage.
// @Summary Apply a flag stage
// @Description Fetches the stage, applies its instructions, and returns the resulting flags payload.
// @Tags Superusers Flag Stages
// @Security SuperUserAuth
// @Produce json
// @Param id query string true "Flag stage ID"
// @Success 200 {object} models.Flags
// @Failure 401 {object} swagger.StatusErrorDoc
// @Failure 404 {object} swagger.StatusErrorDoc
// @Failure 500 {object} swagger.StatusErrorDoc
// @Router /superusers/flags/stages/execute [post]
func flagStagesExecuteHandler(c fiber.Ctx) error {
	flagStage := models.FlagStage{ID: c.Query("id")}
	serr := flagStage.Get()
	if serr != errmsg.EmptyStatusError {
		return utils.StatusError(c, serr)
	}

	err := flagStage.Execute()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	flags := models.Flags{}
	err = flags.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flags)

}
