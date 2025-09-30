package swagger

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

const (
	embedJSONPath = "docs/swagger.json"
	embedYAMLPath = "docs/swagger.yaml"
	diskJSONPath  = "internal/swagger/docs/swagger.json"
	diskYAMLPath  = "internal/swagger/docs/swagger.yaml"
)

const uiTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>OpenHack Backend API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
  <style>
    html, body { margin: 0; padding: 0; background: #fafafa; }
    #swagger-ui { box-sizing: border-box; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js"></script>
  <script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/swagger/doc.json',
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: 'StandaloneLayout',
      deepLinking: true,
      displayRequestDuration: true,
      persistAuthorization: true,
      requestInterceptor: (req) => {
        const authHeader = req.headers && req.headers.Authorization;
        if (authHeader && !/^Bearer /i.test(authHeader)) {
          req.headers.Authorization = 'Bearer ' + authHeader;
        }
        return req;
      },
    });
  };
  </script>
</body>
</html>`

// Register wires swagger-ui routes backed by the generated doc files.
func Register(app *fiber.App) {
	if app == nil {
		return
	}

	app.Get("/swagger", renderUI)
	app.Get("/swagger/index.html", renderUI)

	app.Get("/swagger/doc.json", func(c fiber.Ctx) error {
		data, err := loadDoc(embedJSONPath, diskJSONPath)
		if err != nil {
			return missingSpec(c, err)
		}

		c.Type("json", "utf-8")
		return c.Send(data)
	})

	app.Get("/swagger/doc.yaml", func(c fiber.Ctx) error {
		data, err := loadDoc(embedYAMLPath, diskYAMLPath)
		if err != nil {
			return missingSpec(c, err)
		}

		c.Type("yaml", "utf-8")
		return c.Send(data)
	})
}

func renderUI(c fiber.Ctx) error {
	c.Type("html", "utf-8")
	return c.SendString(uiTemplate)
}

func missingSpec(c fiber.Ctx, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Swagger spec not generated yet. Run `swag init -g cmd/server/main.go -o internal/swagger/docs` to create doc.json/yaml.",
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message": "Failed to read swagger spec",
		"error":   err.Error(),
		"path":    filepath.Clean(diskJSONPath),
	})
}

func loadDoc(embedPath string, diskPath string) ([]byte, error) {
	data, err := swaggerDocs.ReadFile(embedPath)
	if err == nil {
		return data, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	return os.ReadFile(diskPath)
}
