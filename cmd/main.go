package main

import (
	"net/http"
	"os"

	"github.com/fabrianivan-id/technical-test-sawitpro/generated"
	"github.com/fabrianivan-id/technical-test-sawitpro/handler"
	"github.com/fabrianivan-id/technical-test-sawitpro/repository"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const port = ":1323"

func main() {
	e := echo.New()

	// Initialize server
	var server generated.ServerInterface = newServer()

	// Register handlers
	generated.RegisterHandlers(e, server)

	// Serve Swagger JSON
	e.GET("/swagger.json", func(c echo.Context) error {
		swagger, err := generated.GetSwagger()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error loading Swagger spec")
		}

		// Serialize Swagger spec to JSON
		jsonBytes, err := swagger.MarshalJSON()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to serialize Swagger spec")
		}
		return c.JSONBlob(http.StatusOK, jsonBytes)
	})

	// Serve Swagger UI
	e.GET("/swagger", func(c echo.Context) error {
		htmlContent := `
        <!DOCTYPE html>
        <html lang="en">
          <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>API Documentation</title>
            <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui.css" />
          </head>
          <body>
            <div id="swagger-ui"></div>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-bundle.js"></script>
            <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-standalone-preset.js"></script>
            <script>
              const ui = SwaggerUIBundle({
                url: "/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                  SwaggerUIBundle.presets.apis,
                  SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout"
              });
            </script>
          </body>
        </html>
        `
		return c.HTML(http.StatusOK, htmlContent)
	})

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover()) // Recover middleware for better error handling

	// Start server
	e.Logger.Fatal(e.Start(port))
}

func newServer() *handler.Server {
	dbDsn := os.Getenv("DATABASE_URL")
	if dbDsn == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	var repo repository.RepositoryInterface = repository.NewRepository(repository.NewRepositoryOptions{
		Dsn: dbDsn,
	})

	opts := handler.NewServerOptions{
		Repository: repo,
	}

	return handler.NewServer(opts)
}
