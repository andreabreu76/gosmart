package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"gosmart/config"
	"gosmart/entities"
	"image"
	"image/png"
	"io"
	"net/http"
	"time"
)

func GetAvailableModels() ([]entities.OpenAIModel, error) {
	apiKey := config.GetEnv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY não definido no arquivo .env")
	}

	url := "https://api.openai.com/v1/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("erro ao fechar o corpo da resposta: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("requisição falhou: %s", string(body))
	}

	var response struct {
		Data []entities.OpenAIModel `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("erro ao processar resposta: %w", err)
	}

	return response.Data, nil
}

func GetBestModel() (string, error) {
	models, err := GetAvailableModels()
	if err != nil {
		return "", err
	}

	priorities := []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}

	for _, priority := range priorities {
		for _, model := range models {
			if model.ID == priority {
				return model.ID, nil
			}
		}
	}

	return "gpt-3.5-turbo", nil
}

func GenerateText(prompt string) (string, error) {
	apiKey := config.GetEnv("OPENAI_API_KEY")
	apiURL := config.GetEnv("OPENAI_API_URL")
	if apiKey == "" || apiURL == "" {
		return "", errors.New("variáveis OPENAI_API_KEY ou OPENAI_API_URL não estão definidas")
	}

	model, err := GetBestModel()
	if err != nil {
		return "", fmt.Errorf("erro ao obter o melhor modelo: %w", err)
	}

	request := entities.ChatCompletionRequest{
		Model: model,
		Messages: []entities.ChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar o payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("erro ao criar a requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar a requisição: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = fmt.Errorf("erro ao fechar o corpo da resposta: %w", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler a resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(body))
	}

	var response entities.ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("erro ao deserializar a resposta: %w", err)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", errors.New("nenhuma resposta válida retornada")
}

func ExtractTextFromImage(img image.Image) (map[string]string, error) {
	apiKey := config.GetEnv("OPENAI_API_KEY")
	apiURL := config.GetEnv("OPENAI_API_URL")
	if apiKey == "" || apiURL == "" {
		return nil, fmt.Errorf("variáveis OPENAI_API_KEY ou OPENAI_API_URL não configuradas")
	}

	model, err := GetBestModel()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter o melhor modelo: %w", err)
	}

	var imgBuffer bytes.Buffer
	if err := png.Encode(&imgBuffer, img); err != nil {
		return nil, fmt.Errorf("erro ao codificar imagem PNG: %w", err)
	}

	var fixedPrompt = `
Sua função é processar arquivos PDFs relacionados a importação de produtos.
Extraia os campos e seus respectivos valores e crie um objeto JSON com as informações.
Os códigos não devem conter pontos ou traços.
Sempre responda em JSON no formato:
{
  'Campo1': 'Valor1',
  'Campo2': 'Valor2'
};
`
	requestBody := entities.ParsePdfRequest{
		Model:       model,
		MaxTokens:   4096,
		Temperature: 0.2,
		Messages: []entities.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "Você é um assistente que processa imagens relacionadas a documentos PDF.",
			},
			{
				Role:    "user",
				Content: fixedPrompt,
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("erro ao fechar o corpo da resposta: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var extractedData map[string]string
	if len(response.Choices) > 0 {
		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &extractedData); err != nil {
			return nil, fmt.Errorf("erro ao parsear JSON retornado: %w", err)
		}
	}

	return extractedData, nil
}

func ProcessPDFPage(pageContent []byte) (map[string]interface{}, error) {
	apiKey := config.GetEnv("OPENAI_API_KEY")
	apiURL := config.GetEnv("OPENAI_API_URL")
	if apiKey == "" || apiURL == "" {
		return nil, fmt.Errorf("variáveis OPENAI_API_KEY ou OPENAI_API_URL não configuradas")
	}

	model, err := GetBestModel()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter o melhor modelo: %w", err)
	}

	pageBase64 := base64.StdEncoding.EncodeToString(pageContent)

	var pdfPagePrompt = `
Você receberá uma página de um arquivo PDF em base64.
Extraia o texto contido na página e organize as informações relevantes em um objeto JSON.
Se não for possível entender o conteúdo, retorne um JSON vazio.
Sempre responda no formato JSON.
`
	requestBody := entities.ParsePdfRequest{
		Model:       model,
		MaxTokens:   4096,
		Temperature: 0.2,
		Messages: []entities.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "Você é um assistente que processa páginas de PDF para extrair texto e informações úteis.",
			},
			{
				Role:    "user",
				Content: pdfPagePrompt + "\nPágina em base64:\n" + pageBase64,
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("erro ao fechar o corpo da resposta: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var result map[string]interface{}
	if len(response.Choices) > 0 {
		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
			return nil, fmt.Errorf("erro ao parsear JSON retornado: %w", err)
		}
	}

	return result, nil
}

func ProcessImagePage(imageContent []byte) (map[string]interface{}, error) {
	apiKey := config.GetEnv("OPENAI_API_KEY")
	apiURL := config.GetEnv("OPENAI_API_URL")
	if apiKey == "" || apiURL == "" {
		return nil, fmt.Errorf("variáveis OPENAI_API_KEY ou OPENAI_API_URL não configuradas")
	}

	model, err := GetBestModel()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter o melhor modelo: %w", err)
	}

	imageBase64 := base64.StdEncoding.EncodeToString(imageContent)

	var imagePrompt = `
Você receberá uma imagem em base64.
Extraia o texto contido na imagem usando OCR e organize as informações relevantes em um objeto JSON.
Se não for possível entender o conteúdo, retorne um JSON vazio.
Sempre responda no formato JSON.
`
	requestBody := entities.ParsePdfRequest{
		Model:       model,
		MaxTokens:   4096,
		Temperature: 0.2,
		Messages: []entities.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "Você é um assistente que processa imagens para extrair texto e informações úteis usando OCR.",
			},
			{
				Role:    "user",
				Content: imagePrompt + "\nImagem em base64:\n" + imageBase64,
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("erro ao fechar o corpo da resposta: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var result map[string]interface{}
	if len(response.Choices) > 0 {
		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
			return nil, fmt.Errorf("erro ao parsear JSON retornado: %w", err)
		}
	}

	return result, nil
}

func ProcessExtractedText(text string) (map[string]interface{}, error) {
	currentTime := time.Now()
	apiKey := config.GetEnv("OPENAI_API_KEY")
	apiURL := config.GetEnv("OPENAI_API_URL")
	if apiKey == "" || apiURL == "" {
		return nil, fmt.Errorf("variáveis OPENAI_API_KEY ou OPENAI_API_URL não configuradas")
	}

	model, err := GetBestModel()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter o melhor modelo: %w", err)
	}

	prompt := fmt.Sprintf(`
    O seguinte texto foi extraído de uma imagem. Corrija erros de OCR e organize os dados em formato JSON 
    com as chaves identificadas e seus respectivos valores.

    Certifique-se de que:
    1. TODOS os produtos encontrados sejam incluídos no JSON, sem nenhuma omissão.
    2. Retorne SOMENTE o JSON completo, sem explicações, títulos ou mensagens adicionais.

    TEXTO:
    %s
`, text)

	requestBody := entities.ParsePdfRequest{
		Model:       model,
		MaxTokens:   6144,
		Temperature: 0.2,
		Messages: []entities.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "Você é um assistente que corrige erros de OCR e organiza dados em JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("erro ao fechar o corpo da resposta: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var result map[string]interface{}
	if len(response.Choices) > 0 {
		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
			return nil, fmt.Errorf("erro ao parsear JSON retornado: %w", err)
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Error("erro ao serializar JSON: ", err)
	}

	processedTime := time.Since(currentTime)
	log.Infof("Texto processado em %s\n", processedTime)
	log.Infof("Texto processado: %s\n\n", string(jsonData))

	return result, nil
}
