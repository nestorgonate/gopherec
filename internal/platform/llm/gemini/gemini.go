package gemini

import (
	"context"
	"encoding/json"
	"gopherec/internal/domain/entity"
	"gopherec/internal/platform/llm"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
)

type Gemini struct {
	client *genai.Client
}

func NewGemini() *Gemini {
	return &Gemini{}
}

func (g *Gemini) NewClient(c context.Context) error {
	var err error
	g.client, err = genai.NewClient(c, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Gemini) GenerateOpinion(c context.Context, noticia entity.Noticia, referencia string) (string, error) {
	log.Println("Gemini esta opinando sobre una noticia")
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(llm.InstructionOpinion, ""),
		Temperature:       new(float32(1)),
	}
	prompt := llm.OpinionPrompt(noticia, referencia)
	result, err := g.client.Models.GenerateContent(c, "gemini-3-flash-preview", genai.Text(prompt), config)
	if err != nil {
		return "", err
	}
	if len(result.Candidates) == 0 {
		return "", llm.ErrNoGeneratedContext
	}
	return result.Text(), nil
}

func (g *Gemini) Categorize(c context.Context, noticia entity.Noticia) (entity.Clasificacion, error) {
	log.Println("Gemini esta categorizando una noticia")
	var classification entity.Clasificacion
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(llm.InstructionClasifier, ""),
		Temperature:       new(float32(0.1)),
	}
	prompt := llm.CategorizacionPrompt(noticia)
	rawJSON, err := g.client.Models.GenerateContent(c, "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		return entity.Clasificacion{}, err
	}
	if len(rawJSON.Candidates) == 0 {
		return entity.Clasificacion{}, llm.ErrNoGeneratedContext
	}
	jsonCleaned := strings.TrimPrefix(rawJSON.Text(), "```json")
	jsonCleaned = strings.TrimSuffix(jsonCleaned, "```")
	log.Printf("Clasificacion de Gemini: %v\n", jsonCleaned)
	err = json.Unmarshal([]byte(jsonCleaned), &classification)
	if err != nil {
		return entity.Clasificacion{}, err
	}
	return classification, nil
}
