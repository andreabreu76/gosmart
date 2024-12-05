package handlers

import (
	"gosmart/entities"
	"gosmart/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/protobuf/proto"
)

func ExampleHandler(c *fiber.Ctx) error {
	body := c.Body()

	req := &entities.ExampleRequest{}

	if err := proto.Unmarshal(body, req); err != nil {
		log.Println("Erro ao desserializar Protobuf: ", err)
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Invalid Protobuf data"})
	}

	if err := services.LogToRedis("logg", req.Input); err != nil {
		log.Println("Erro ao logar no Redis: ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	res := &entities.ExampleResponse{Output: "Processed: " + req.Input}

	responseBytes, err := proto.Marshal(res)
	if err != nil {
		log.Println("Erro ao serializar resposta Protobuf: ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	c.Set("Content-Type", "application/protobuf")
	return c.Send(responseBytes)
}
