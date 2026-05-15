package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopherec/internal/domain/entity"
	"gopherec/internal/platform/llm"
	"log"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type DeepSeek struct {
	client *openai.Client
}

func NewDeepSeek() *DeepSeek {
	config := openai.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"))
	config.BaseURL = "https://api.deepseek.com"

	return &DeepSeek{
		client: openai.NewClientWithConfig(config),
	}
}

func (d DeepSeek) GenerateOpinion(c context.Context, noticia entity.Noticia, referencia string) (string, error) {
	log.Println("Deepseek esta opinando una noticia")
	resp, err := d.client.CreateChatCompletion(
		c,
		openai.ChatCompletionRequest{
			Model: "deepseek-chat",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: llm.InstructionOpinion,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: llm.OpinionPrompt(noticia, referencia),
				},
			},
			MaxCompletionTokens: 280,
			Temperature:         0.9,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (d DeepSeek) Categorize(c context.Context, noticia entity.Noticia) (entity.Clasificacion, error) {
	log.Println("DeepSeek está categorizando una noticia")
	var classification entity.Clasificacion

	prompt := llm.CategorizacionPrompt(noticia)

	resp, err := d.client.CreateChatCompletion(
		c,
		openai.ChatCompletionRequest{
			Model: "deepseek-chat",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: llm.InstructionClasifier,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return entity.Clasificacion{}, fmt.Errorf("error en DeepSeek: %w", err)
	}

	if len(resp.Choices) == 0 {
		return entity.Clasificacion{}, errors.New("DeepSeek no devolvió candidatos")
	}

	rawJSON := resp.Choices[0].Message.Content
	jsonCleaned := strings.TrimSpace(rawJSON)
	jsonCleaned = strings.TrimPrefix(jsonCleaned, "```json")
	jsonCleaned = strings.TrimSuffix(jsonCleaned, "```")
	jsonCleaned = strings.TrimSpace(jsonCleaned)

	log.Printf("Clasificación de DeepSeek: %v\n", jsonCleaned)

	err = json.Unmarshal([]byte(jsonCleaned), &classification)
	if err != nil {
		return entity.Clasificacion{}, fmt.Errorf("error al decodificar JSON: %w", err)
	}

	return classification, nil
}
