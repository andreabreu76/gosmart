package main

import (
	"gosmart/config"
	"gosmart/router"
	"gosmart/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/swagger"
	_ "gosmart/docs"
)

// @title GoSmart API
// @version 1.0
// @description API para o sistema GoSmart com integração OpenAI e suporte a logs
// @termsOfService http://swagger.io/terms/
// @contact.name Suporte
// @contact.url http://github.com/yooga/gosmart
// @contact.email suporte@gosmart.com
// @host localhost:3000
// @BasePath /
func main() {
	config.LoadEnv()
	services.InitRedis()

	log.Info("Servidor iniciado na porta 3000")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		return c.Next()
	})

	router.SetupRoutes(app)
	app.Get("/swagger/*", swagger.HandlerDefault)

	log.Fatal(app.Listen(":3000"))
}
