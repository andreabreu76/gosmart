package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"gosmart/services"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ProcessPDFHandler godoc
// @Summary Processa um arquivo PDF
// @Description Recebe um arquivo PDF e processa cada página, retornando os resultados
// @Tags PDF
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF file to be processed"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string "Failed to receive the file"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /process-pdf [post]
func ProcessPDFHandler(c *fiber.Ctx) error {
	currentTime := time.Now()
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Falha ao receber o arquivo"})
	}

	uniqueID := uuid.New().String()

	tempDir := "./pdf_temp"
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao criar diretório temporário"})
	}

	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%s.pdf", uniqueID))
	if err := c.SaveFile(file, tempFilePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao salvar arquivo PDF"})
	}

	// Converte o PDF em imagens
	imageFiles, err := convertPDFToImages(tempFilePath, tempDir, uniqueID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erro ao converter PDF para imagens"})
	}

	results := make([]map[string]interface{}, len(imageFiles))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 2)

	for i, imageFile := range imageFiles {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, imgPath string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// Extrai texto da imagem usando Tesseract
			extractedText, err := extractTextWithTesseract(imgPath)
			if err != nil {
				log.Printf("Erro ao extrair texto da imagem %d: %v", idx+1, err)
				results[idx] = map[string]interface{}{"error": "Erro ao extrair texto da imagem"}
				return
			}

			log.Printf("Texto extraído da imagem %d: %s", idx+1, extractedText)

			// Processa o texto com OpenAI
			result, err := services.ProcessExtractedText(extractedText)
			if err != nil {
				log.Printf("Erro ao processar texto extraído da imagem %d: %v", idx+1, err)
				results[idx] = map[string]interface{}{"error": "Erro ao processar texto extraído"}
				return
			}

			results[idx] = result
		}(i, imageFile)
	}

	wg.Wait()

	elapsedTime := time.Since(currentTime)
	log.Printf("Tempo total de processamento: %v", elapsedTime)
	return c.JSON(results)
}

func convertPDFToImages(pdfPath string, outputDir string, uniqueID string) ([]string, error) {
	imagesOutputDir := filepath.Join(outputDir, uniqueID+"_images")
	if err := os.MkdirAll(imagesOutputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório para imagens: %w", err)
	}

	imagePattern := filepath.Join(imagesOutputDir, fmt.Sprintf("%s_page_%%d.png", uniqueID))
	cmd := exec.Command("mutool", "draw", "-o", imagePattern, pdfPath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao converter PDF para imagens: %w", err)
	}

	imageFiles, err := filepath.Glob(filepath.Join(imagesOutputDir, fmt.Sprintf("%s_page_*.png", uniqueID)))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar imagens geradas: %w", err)
	}

	return imageFiles, nil
}

func extractTextWithTesseract(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "--psm", "6") // --psm 6 é ideal para tabelas
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("erro ao executar tesseract: %w", err)
	}
	return string(output), nil
}
