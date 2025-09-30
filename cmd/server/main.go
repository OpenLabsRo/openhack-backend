package main

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
