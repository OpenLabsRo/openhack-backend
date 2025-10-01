package main

// Tag catalog (order reflects Swagger sidebar)
// @tag.name General
// @tag.description Service-wide operational endpoints.

// @tag.name Superusers Auth
// @tag.description Superuser authentication flows.
// @tag.name Superusers Health
// @tag.description Operational endpoints for superuser services.

// @tag.name Superusers Accounts
// @tag.description Staff tooling for participant account shells.
// @tag.name Superusers Flags
// @tag.description Feature flag administration endpoints.
// @tag.name Superusers Flag Stages
// @tag.description Stage-based flag rollout orchestration endpoints.
// @tag.name Superusers Staff
// @tag.description Staff passport scanning and badge tooling endpoints.
// @tag.name Superusers Check-In
// @tag.description Badge assignment and check-in tooling for staff.

// @tag.name Accounts Auth
// @tag.description Registration and login flows for participants.
// @tag.name Accounts Health
// @tag.description Health probes for participant account services.
// @tag.name Accounts Identity
// @tag.description Identity introspection for signed-in participants.
// @tag.name Accounts Profile
// @tag.description Participant profile maintenance endpoints.
// @tag.name Accounts Flags
// @tag.description Feature flag lookup for participants.

// @tag.name Teams Health
// @tag.description Health probes for team services.
// @tag.name Teams Core
// @tag.description Core team lifecycle management endpoints.
// @tag.name Teams Members
// @tag.description Team membership management endpoints.
// @tag.name Teams Submission
// @tag.description Submission metadata update endpoints.

import (
	"backend/internal"
	"backend/internal/env"
	"backend/internal/swagger"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// @title OpenHack Backend API
// @version 25.10.01.0
// @description Backend API for OpenHack handling participant accounts, teams, feature flags, and superuser check-in tooling.
// @BasePath /
// @securityDefinitions.apikey AccountAuth
// @in header
// @name Authorization
// @description Provide the participant bearer token as `Bearer <token>`.
// @securityDefinitions.apikey SuperUserAuth
// @in header
// @name Authorization
// @description Provide the superuser bearer token as `Bearer <token>`.

func main() {
	// these are required flags
	// for whichever deployment profile is being used
	deploymentFlag := flag.String("deployment", "", "deployment profile (dev|test|prod)")
	portFlag := flag.String("port", "", "port to listen on")

	envRoot := flag.String("env-root", "", "directory containing environment files")
	appVersion := flag.String("app-version", "", "application version override")

	// parsing the flags
	flag.Parse()

	// the deployment flag first
	deployment := strings.TrimSpace(*deploymentFlag)
	if deployment == "" {
		args := flag.Args()
		if len(args) == 0 {
			fmt.Println("Usage: server --deployment <type> --port <port> [--env-root <dir>] [--app-version <version>]")
			os.Exit(1)
		}
		deployment = strings.TrimSpace(args[0])
	}

	if deployment == "" {
		log.Fatal("deployment is required")
	}

	port := strings.TrimSpace(*portFlag)
	if port == "" {
		log.Fatal("port is required")
	}

	app := internal.SetupApp(deployment, *envRoot, *appVersion)
	swagger.Register(app)

	fmt.Println("APP VERSION:", env.VERSION)
	if err := app.Listen(fmt.Sprintf(":%s", port), fiber.ListenConfig{
		EnablePrefork: env.PREFORK,
	}); err != nil {
		log.Fatalf("Error listening on port %s: %v", port, err)
	}
}
