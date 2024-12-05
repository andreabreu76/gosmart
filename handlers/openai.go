package handlers

import (
	"github.com/gofiber/fiber/v2/log"
	"gosmart/services"

	"github.com/gofiber/fiber/v2"
)

// OpenAIHandler
// @Summary Gera uma resposta da OpenAI
// @Description Recebe um prompt e retorna uma resposta gerada pelo modelo OpenAI
// @Tags OpenAI
// @Accept json
// @Produce json
// @Param request body entities.OpenAIRequest true "Prompt para a OpenAI"
// @Success 200 {object} map[string]string "Resposta gerada"
// @Failure 400 {object} map[string]string "Erro de validação"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /openai [post]
func OpenAIHandler(c *fiber.Ctx) error {
	type Request struct {
		Prompt string `json:"prompt"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		log.Error("Erro ao processar operação OpenAI: ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	response, err := services.GenerateText(req.Prompt)
	if err != nil {
		log.Error("Erro ao processar operação OpenAI: ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"response": response})
}
