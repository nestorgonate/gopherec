package service

import (
	"context"
	"gopherec/internal/domain"
	"gopherec/internal/domain/entity"
	"log"
)

type NoticiasService struct {
	gemini   domain.LLMProvider
	rss      domain.RSS
	repo     domain.NoticiasRepo
	deepseek domain.LLMProvider
}

func NewNoticiasService(gemini domain.LLMProvider, rss domain.RSS, repo domain.NoticiasRepo, deepseek domain.LLMProvider) *NoticiasService {
	return &NoticiasService{
		gemini:   gemini,
		rss:      rss,
		repo:     repo,
		deepseek: deepseek,
	}
}

// Obtiene noticias de RSS, las clasifica y almacena en mongodb
func (service *NoticiasService) Get(c context.Context) (bool, error) {
	var clasificacion entity.Clasificacion
	log.Println("Obtieniendo noticias RSS")
	noticias, err := service.rss.GetPolitics(c)
	if err != nil {
		return false, err
	}
	noticia := noticias[0]
	clasificacion, err = service.gemini.Categorize(c, noticia)
	if err != nil {
		log.Printf("Gemini falló: %v. Reintentando con DeepSeek...", err)
		clasificacion, err = service.deepseek.Categorize(c, noticia)
		if err != nil {
			log.Printf("Deepseek fallo: %v", err)
			return false, err
		}
	}
	log.Printf("Clasificacion de la noticia: %+v", clasificacion)
	if clasificacion.Category == entity.Sensible {
		log.Printf("Omitiendo noticia categorizada como sensible")
	}
	if clasificacion.Category == entity.Otros && clasificacion.SensitivityLevel <= 2 {
		log.Printf("Omitiendo noticia que no es relevante")
	}
	noticia.SensitivityLevel = clasificacion.SensitivityLevel
	countNew, err := service.repo.Save(c, noticia)
	if err != nil {
		return false, err
	}
	if countNew == 0 {
		return false, nil
	}
	log.Printf("Se agregaron a la base de datos: %v noticias\n", countNew)
	return true, nil
}
