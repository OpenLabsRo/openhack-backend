package meta

import "github.com/gofiber/fiber/v3"

func Routes(r fiber.Router) {
	r.Get("/ping", pingHandler)
	r.Get("/version", versionHandler)
}
