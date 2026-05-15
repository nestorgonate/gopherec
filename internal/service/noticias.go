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

// Obtiene noticias de RSS, almacena en MongoDB temporalmente, Gemini o Deepseek las clasifica, borra o mantiene la noticia dependiendo del resultado
func (service *NoticiasService) Get(c context.Context) (bool, error) {
	var clasificacion entity.Clasificacion
	fieldUpdate := make(map[string]any)
	log.Println("Obtieniendo noticias RSS")
	noticias, err := service.rss.GetPolitics(c)
	if err != nil {
		return false, err
	}
	noticia := noticias[0]
	countNew, noticiaId, err := service.repo.Save(c, noticia)
	if err != nil {
		return false, err
	}
	if countNew == 0 {
		log.Println("No hay noticias nuevas, omitiendo clasificacion con Gemini o Deepseek")
		return false, nil
	}
	clasificacion, err = service.gemini.Categorize(c, noticia)
	if err != nil {
		log.Printf("Gemini falló: %v. Reintentando con DeepSeek...", err)
		clasificacion, err = service.deepseek.Categorize(c, noticia)
		if err != nil {
			log.Printf("Deepseek fallo: %v", err)
			return false, err
		}
	}
	if clasificacion.Category == entity.Sensible {
		log.Printf("Omitiendo noticia categorizada como sensible")
		fieldUpdate["status"] = entity.Rejected
		err := service.repo.Update(c, noticiaId, fieldUpdate)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	if clasificacion.Category == entity.Otros && clasificacion.SensitivityLevel <= 8 {
		log.Printf("Omitiendo noticia politica que no es relevante")
		err := service.repo.Delete(c, noticiaId)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	if clasificacion.Category == entity.Politica && clasificacion.SensitivityLevel < 7 {
		log.Printf("Omitiendo noticia de politica que no es relevante")
		err := service.repo.Delete(c, noticiaId)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	log.Printf("Clasificacion de la noticia: %+v", clasificacion)
	fieldUpdate["sensitivityLevel"] = clasificacion.SensitivityLevel
	err = service.repo.Update(c, noticiaId, fieldUpdate)
	if err != nil {
		return false, err
	}
	log.Printf("Se agregaron a la base de datos: %v noticias\n", countNew)
	return true, nil
}
