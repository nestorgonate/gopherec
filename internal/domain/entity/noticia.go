package entity

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatusEnum string

const (
	Pending    StatusEnum = "pending"
	Rejected   StatusEnum = "rejected"
	Processing StatusEnum = "processing"
	Published  StatusEnum = "published"
)

type Noticia struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title            string             `bson:"title" json:"title"`
	Description      string             `bson:"description" json:"description"`
	Content          string             `bson:"content" json:"content"`
	Link             string             `bson:"link" json:"link"`
	Category         Categoria          `bson:"category" json:"category"`
	Status           StatusEnum         `bson:"status" json:"status"`
	SensitivityLevel int64              `bson:"sensitivityLevel" json:"sensitivityLevel"`
	Published        time.Time          `bson:"published" json:"published"`
}

var ErrNotNewNews error = errors.New("No hay noticias pendientes por opinar")
