package swagger

import (
	"backend/internal/env"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	embedJSONPath = "docs/swagger.json"
	diskJSONPath  = "internal/swagger/docs/swagger.json"
	// swaggerUIPath = "https://unpkg.com/swagger-ui-dist@latest"
	swaggerUIPath = "https://openhack-swagger.vercel.app/"
)

var uiTemplate = fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>OpenHack Backend API Docs</title>
  <link rel="icon" type="image/png" href="https://dl.openhack.ro/icons/logo.png" />
  <link rel="stylesheet" href="%s/swagger-ui.css" />
  
	<style>
	html, body { margin: 0; padding: 0; background: #000000; }`+
	// #swagger-ui { box-sizing: border-box; background: #171717; }
	// .swagger-ui .topbar { background: #000 !important; }
	// .swagger-ui .info { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .opblock { background: #171717 !important; border-color: #333 !important; }
	// .swagger-ui .opblock-summary { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .parameters { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .responses { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .response { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .model { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .model-title { color: #fff !important; }
	// .swagger-ui .prop { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .parameter { background: #171717 !important; color: #fff !important; }
	// .swagger-ui .parameter__name { color: #fff !important; }
	// .swagger-ui .parameter__type { color: #fff !important; }
	// .swagger-ui .response__title { color: #fff !important; }
	// .swagger-ui .response__description { color: #fff !important; }
	// .swagger-ui .markdown p { color: #fff !important; }
	// .swagger-ui .markdown code { background: #333 !important; color: #fff !important; }
	// .swagger-ui .btn { background: #333 !important; color: #fff !important; border-color: #555 !important; }
	// .swagger-ui .btn:hover { background: #555 !important; }
	// .swagger-ui .select { background: #333 !important; color: #fff !important; border-color: #555 !important; }
	// .swagger-ui input { background: #333 !important; color: #fff !important; border-color: #555 !important; }
	// .swagger-ui textarea { background: #333 !important; color: #fff !important; border-color: #555 !important; }
	`</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="%s/swagger-ui-bundle.js"></script>
  <script src="%s/swagger-ui-standalone-preset.js"></script>
	<script>
	window.onload = () => {
		window.ui = SwaggerUIBundle({
			url: './docs/doc.json',
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
</html>`, swaggerUIPath, swaggerUIPath, swaggerUIPath)

// Register wires swagger-ui routes backed by the generated doc files.
func Register(app *fiber.App) {
	if app == nil {
		return
	}

	version := env.VERSION
	if version == "" {
		version = "docs" // fallback if VERSION not set
	}

	// group := app.Group("/" + version)

	app.Get("/docs", func(c fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.SendString(uiTemplate)
	})

	app.Get("/docs/doc.json", func(c fiber.Ctx) error {
		data, err := loadDoc(embedJSONPath, diskJSONPath)
		if err != nil {
			return missingSpec(c, err)
		}

		version := env.VERSION
		if version == "" {
			version = "docs"
		}

		if os.Getenv("NO_HYPER") == "false" || os.Getenv("NO_HYPER") == "" {
			data = stampJSON(c, data, version)
		}

		c.Type("json", "utf-8")
		return c.Send(data)
	})
}

func missingSpec(c fiber.Ctx, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Swagger spec not generated yet. Run `swag init -g cmd/server/main.go -o internal/swagger/docs` to create doc.json.",
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
	// If NO_HYPER is explicitly set to "true", skip version stamping.
	if strings.EqualFold(strings.TrimSpace(env.NO_HYPER), "true") {
		return data
	}

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

func stampJSON(c fiber.Ctx, data []byte, version string) []byte {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return data
	}

	// swagger: 2.0
	u, err := url.Parse(c.BaseURL())
	if err == nil {
		// basePath should include the version prefix
		doc["basePath"] = "/" + strings.Trim(version, "/")
		doc["schemes"] = []string{u.Scheme}
	}

	encoded, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return data
	}

	return encoded
}
