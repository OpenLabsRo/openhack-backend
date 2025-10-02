package flags

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// flagsGetHandler returns the current flag assignments and active stage.
// @Summary Retrieve all feature flags
// @Description Reads the cached flag document so console operators can inspect rollout state.
// @Tags Superusers Flags
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} models.Flags
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/flags [get]
func flagsGetHandler(c fiber.Ctx) error {
	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flags)
}

// flagsSetHandler flips a single feature flag.
// @Summary Set a feature flag
// @Description Updates one flag value and returns the refreshed flag map for verification.
// @Tags Superusers Flags
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body FlagSetRequest true "Flag toggle"
// @Success 200 {object} FlagAssignments
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/flags [post]
func flagsSetHandler(c fiber.Ctx) error {
	var body FlagSetRequest
	json.Unmarshal(c.Body(), &body)

	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	err = flags.Set(body.Flag, body.Value)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flags.Flags)
}

// flagsSetBulkHandler overwrites multiple flags in one request.
// @Summary Bulk update feature flags
// @Description Applies a map of flag values and echoes the resulting assignments.
// @Tags Superusers Flags
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body FlagAssignments true "Flag assignments"
// @Success 200 {object} FlagAssignments
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/flags [put]
func flagsSetBulkHandler(c fiber.Ctx) error {
	var body FlagAssignments
	json.Unmarshal(c.Body(), &body)

	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	err = flags.SetBulk(body)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flags.Flags)
}

// flagsUnsetHandler removes a stored flag entry.
// @Summary Remove a feature flag entry
// @Description Deletes a flag from the assignments and returns the remaining map.
// @Tags Superusers Flags
// @Security SuperUserAuth
// @Accept json
// @Produce json
// @Param payload body FlagUnsetRequest true "Flag identifier"
// @Success 200 {object} FlagAssignments
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/flags [delete]
func flagsUnsetHandler(c fiber.Ctx) error {
	var body FlagUnsetRequest
	json.Unmarshal(c.Body(), &body)

	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	err = flags.Unset(body.Flag)
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flags.Flags)
}

// flagsResetHandler disables every feature flag.
// @Summary Reset all feature flags to false
// @Description Sets every tracked flag back to false and returns the cleared map.
// @Tags Superusers Flags
// @Security SuperUserAuth
// @Produce json
// @Success 200 {object} FlagAssignments
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/flags/reset [post]
func flagsResetHandler(c fiber.Ctx) error {

	flags := models.Flags{}
	err := flags.Get()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	err = flags.Reset()
	if err != nil {
		return utils.StatusError(
			c, errmsg.InternalServerError(err),
		)
	}

	return c.JSON(flags.Flags)

}

// flagsTestHandler verifies that the current token satisfies middleware requirements.
// @Summary Validate that the current JWT satisfies flag middleware
// @Description Ensures the caller holds both the required role and feature toggles before allowing access.
// @Tags Superusers Flags
// @Security SuperUserAuth
// @Produce plain
// @Success 200 {string} string "it passed"
// @Failure 401 {object} errmsg._SuperUserNoToken
// @Failure 401 {object} errmsg._FlagRequired
// @Failure 500 {object} errmsg._InternalServerError
// @Router /superusers/flags/test [get]
func flagsTestHandler(c fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("it passed")
}
