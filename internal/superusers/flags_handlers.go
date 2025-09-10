package superusers

import (
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

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

func flagsSetHandler(c fiber.Ctx) error {
	var body struct {
		Flag  string `json:"flag"`
		Value bool   `json:"value"`
	}
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

func flagsSetBulkHandler(c fiber.Ctx) error {
	var body map[string]bool
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

func flagsUnsetHandler(c fiber.Ctx) error {
	var body struct {
		Flag string `json:"flag"`
	}
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

func flagsTestHandler(c fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("it passed")
}
