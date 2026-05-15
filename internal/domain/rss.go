package domain

import (
	"context"
	"gopherec/internal/domain/entity"
)

type RSS interface {
	GetPolitics(c context.Context) ([]entity.Noticia, error)
}
