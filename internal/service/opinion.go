package service

import (
	"context"
	"gopherec/internal/domain"
	"log"
)

type OpinionService struct {
	gemini    domain.LLMProvider
	repo      domain.NoticiasRepo
	deepkseek domain.LLMProvider
}

func NewOpinionService(gemini domain.LLMProvider, repo domain.NoticiasRepo, deepseek domain.LLMProvider) *OpinionService {
	return &OpinionService{
		gemini:    gemini,
		repo:      repo,
		deepkseek: deepseek,
	}
}

func (service *OpinionService) GenerateOpinion(c context.Context) (string, error) {
	var opinion string
	noticia, err := service.repo.GetPending(c)
	log.Printf("Noticia: %+v\n", noticia)
	opinion, err = service.gemini.GenerateOpinion(c, noticia, "No hay referencias historicas aun, utiliza el conocimiento con el que fuiste entrenado")
	if err != nil {
		log.Printf("Gemini no pudo opinar de la noticia: %v\n", err)
		opinion, err = service.deepkseek.GenerateOpinion(c, noticia, "No hay referencias historicas aun, utiliza el conocimiento con el que fuiste entrenado")
		if err != nil {
			log.Printf("Deepseek fallo: %v", err)
			return "", err
		}
	}
	return opinion, nil
}
