package domain

import (
	"context"
	"gopherec/internal/domain/entity"
)

type LLMProvider interface {
	GenerateOpinion(c context.Context, noticia entity.Noticia, referencia string) (string, error)
	Categorize(c context.Context, noticia entity.Noticia) (entity.Clasificacion, error)
}
