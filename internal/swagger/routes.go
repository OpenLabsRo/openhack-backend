package swagger

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"backend/internal/env"

	"github.com/gofiber/fiber/v3"
	"gopkg.in/yaml.v3"
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
	format := filepath.Ext(embedPath)

	data, err := swaggerDocs.ReadFile(embedPath)
	if err == nil {
		return stampDoc(format, data), nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	data, err = os.ReadFile(diskPath)
	if err != nil {
		return nil, err
	}

	return stampDoc(format, data), nil
}

func stampDoc(ext string, data []byte) []byte {
	version := strings.TrimSpace(env.VERSION)
	if version == "" {
		return data
	}

	updated, err := applyVersion(ext, data, version)
	if err != nil {
		log.Printf("swagger: failed to stamp version %q into %s doc: %v", version, ext, err)
		return data
	}

	return updated
}

func applyVersion(ext string, data []byte, version string) ([]byte, error) {
	switch ext {
	case ".json":
		return updateJSONDoc(data, version)
	case ".yaml", ".yml":
		return updateYAMLDoc(data, version)
	default:
		return data, nil
	}
}

func updateJSONDoc(data []byte, version string) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	stampDocMap(doc, version)

	encoded, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return nil, err
	}

	if len(encoded) == 0 || encoded[len(encoded)-1] != '\n' {
		encoded = append(encoded, '\n')
	}

	return encoded, nil
}

func updateYAMLDoc(data []byte, version string) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	stampDocMap(doc, version)

	return yaml.Marshal(doc)
}

func stampDocMap(doc map[string]any, version string) {
	if doc == nil {
		return
	}

	info := ensureMap(doc, "info")
	info["version"] = version

	pathsVal, ok := doc["paths"].(map[string]any)
	if !ok {
		return
	}

	versionPath, ok := pathsVal["/version"].(map[string]any)
	if !ok {
		return
	}

	get, ok := versionPath["get"].(map[string]any)
	if !ok {
		return
	}

	responses, ok := get["responses"].(map[string]any)
	if !ok {
		return
	}

	resp200, ok := responses["200"].(map[string]any)
	if !ok {
		return
	}

	resp200["description"] = version

	schema, ok := resp200["schema"].(map[string]any)
	if !ok {
		schema = map[string]any{}
		resp200["schema"] = schema
	}

	schema["example"] = version
}

func ensureMap(root map[string]any, key string) map[string]any {
	val, ok := root[key]
	if ok {
		if existing, ok := val.(map[string]any); ok {
			return existing
		}
	}

	created := map[string]any{}
	root[key] = created
	return created
}
