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
	twitter   domain.TwitterAPI
}

func NewOpinionService(gemini domain.LLMProvider, repo domain.NoticiasRepo, deepseek domain.LLMProvider, twitter domain.TwitterAPI) *OpinionService {
	return &OpinionService{
		gemini:    gemini,
		repo:      repo,
		deepkseek: deepseek,
		twitter:   twitter,
	}
}

func (service *OpinionService) GenerateOpinion(c context.Context) {
	var opinion string
	noticia, err := service.repo.GetPending(c)
	if err != nil {
		log.Printf("ERROR: No se pudo obtener noticias pendientes de la base de datos: %v\n", err)
		return
	}
	if noticia.ID.IsZero() {
		return
	}
	log.Printf("DEBUG: Noticia: %+v\n", noticia)
	opinion, err = service.gemini.GenerateOpinion(c, noticia, "No hay referencias historicas aun, utiliza el conocimiento con el que fuiste entrenado")
	if err != nil {
		log.Printf("ERROR: Gemini no pudo opinar de la noticia: %v\n", err)
		opinion, err = service.deepkseek.GenerateOpinion(c, noticia, "No hay referencias historicas aun, utiliza el conocimiento con el que fuiste entrenado")
		if err != nil {
			log.Printf("ERROR: Deepseek fallo: %v", err)
			return
		}
	}
	runes := []rune(opinion)
	if len(runes) > 280 {
		opinion = string(runes[:280])
	}
	postId, err := service.twitter.Post(c, opinion)
	if err != nil {
		log.Printf("ERROR: Twitter no pudo opinar de la noticia: %v\n", err)
		return
	}
	log.Printf("Nuevo post: %v\n", postId)
	fieldUpdate := make(map[string]any)
	fieldUpdate["status"] = "published"
	err = service.repo.Update(c, noticia.ID, fieldUpdate)
	if err != nil {
		return
	}
}
