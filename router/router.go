package router

import (
	"gosmart/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/example", handlers.ExampleHandler)
	app.Post("/openai", handlers.OpenAIHandler)
	app.Post("/process-pdf", handlers.ProcessPDFHandler)
}
