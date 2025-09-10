package superusers

import (
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
		return utils.Error(c, 500, err)
	}

	return c.JSON(flags.Flags)
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
		return utils.Error(c, 500, err)
	}

	err = flags.Set(body.Flag, body.Value)
	if err != nil {
		return utils.Error(c, 500, err)
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
		return utils.Error(c, 500, err)
	}

	err = flags.Unset(body.Flag)
	if err != nil {
		return utils.Error(c, 500, err)
	}

	return c.JSON(flags.Flags)
}

func flagsTestHandler(c fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("it passed")
}
