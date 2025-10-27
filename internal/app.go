package internal

import (
	"backend/internal/accounts"
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/events"
	"backend/internal/judge"
	"backend/internal/meta"
	"backend/internal/models"
	"backend/internal/superusers"
	"backend/internal/teams"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func getEmitterConfig(deployment string) events.Config {
	switch deployment {
	case "test":
		return events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 100 * time.Millisecond,
		}
	case "dev":
		return events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 1 * time.Second,
		}
	default:
		return events.Config{
			Buffer:     1000,
			BatchSize:  50,
			FlushEvery: 2 * time.Second,
		}
	}
}

func initBadgePileSalt() {
	setting := &models.Setting{Name: models.SettingBadgePileSalt}

	serr := setting.Get()
	if serr == errmsg.EmptyStatusError {
		if salt, ok := setting.Value.(string); ok {
			env.BADGE_PILES_SALT = salt
		}
		return
	}

	if serr != errmsg.SettingNotFound {
		log.Printf("failed to load badge pile salt from settings: %s", serr.Message)
	}
}

func SetupApp(deployment string, envRoot string, appVersion string) *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	// initializing environment
	env.Init(envRoot, appVersion)

	// initializing db
	if err := db.InitDB(deployment); err != nil {
		log.Fatal("Could not connect to MongoDB")
		return nil
	}

	// initializing cache
	if err := db.InitCache(deployment); err != nil {
		log.Fatal("Could not connect to Redis")
		return nil
	}

	// creating the events emitter
	events.Em = events.NewEmitter(
		db.Events,
		getEmitterConfig(deployment),
		deployment,
	)

	// loading the BADGE_PILE_SALT
	initBadgePileSalt()

	meta.Routes(app.Group("/meta"))
	superusers.Routes(app.Group("/superusers"))
	accounts.Routes(app.Group("/accounts"))
	teams.Routes(app.Group("/teams"))
	judge.Routes(app.Group("/judge"))

	// temporary for list-unsubscribe
	app.Get("/unsubscribe", func(c fiber.Ctx) error {
		events.Em.ListUnsubscribe(c.Query("email"))

		return c.SendString("OK")
	})

	return app
}
