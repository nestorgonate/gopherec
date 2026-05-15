package domain

import (
	"context"
	"gopherec/internal/domain/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type NoticiasRepo interface {
	SearchHistory(c context.Context, vector []float64, category string) ([]bson.M, error)
	Save(c context.Context, noticias ...entity.Noticia) (uint, error)
	GetPending(c context.Context) (entity.Noticia, error)
}
