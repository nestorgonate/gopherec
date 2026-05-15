package entity

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Historia struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Content       string             `bson:"content" json:"content"`
	VectorContent []float64          `bson:"vectorContent" json:"vectorContent"`
	Category      Categoria          `bson:"category" json:"category"`
	Year          int                `bson:"year,omitempty" json:"year"`
}

var ErrNoNewNewsItem = errors.New("No hay nuevas noticias")
