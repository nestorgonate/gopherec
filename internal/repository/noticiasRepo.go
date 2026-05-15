package repository

import (
	"context"
	"gopherec/internal/domain/entity"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NoticiasRepo struct {
	db *mongo.Database
}

func NewNoticiasRepo(db *mongo.Database) *NoticiasRepo {
	return &NoticiasRepo{
		db: db,
	}
}

func (n NoticiasRepo) SearchHistory(c context.Context, vector []float64, category string) ([]bson.M, error) {
	collection := n.db.Collection("historia_ecuador")
	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: bson.D{
			{Key: "index", Value: "idx_vectorHistory"}, // Nombre del índice en Atlas
			{Key: "path", Value: "vectorContent"},      // Campo donde están los []float64
			{Key: "queryVector", Value: vector},
			{Key: "numCandidates", Value: 100},
			{Key: "limit", Value: 5},                                           // Traer los 3 hechos más parecidos
			{Key: "filter", Value: bson.D{{Key: "category", Value: category}}}, // Filtro opcional
		}}},
	}
	cursor, err := collection.Aggregate(c, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var result []bson.M
	err = cursor.All(c, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (n NoticiasRepo) Save(c context.Context, noticias ...entity.Noticia) (uint, error) {
	collection := n.db.Collection("noticias")
	var countNew uint = 0
	for _, noticia := range noticias {
		filter := bson.M{"link": noticia.Link}
		// El comando $set reemplaza los campos, $setOnInsert solo se ejecuta si es nueva
		update := bson.M{
			"$set": noticia,
		}
		opts := options.Update().SetUpsert(true)
		result, err := collection.UpdateOne(c, filter, update, opts)
		if err != nil {
			return 0, nil
		}
		if result.UpsertedCount > 0 {
			countNew++
		}
	}
	return countNew, nil
}

func (n NoticiasRepo) GetPending(c context.Context) (entity.Noticia, error) {
	collection := n.db.Collection("noticias")
	count, _ := collection.CountDocuments(c, bson.M{"status": "Pending"})
	log.Printf("DEBUG: Documentos con status Pending encontrados: %d", count)
	filter := bson.M{"status": "Pending"}
	//Opciones: Ordenar por fecha de publicación (la más reciente primero)
	opts := options.FindOne().SetSort(bson.M{"published": -1})
	var noticia entity.Noticia
	err := collection.FindOne(c, filter, opts).Decode(&noticia)
	if err != nil {
		log.Printf("DEBUG: Error al decodificar en Struct: %v", err)
		return entity.Noticia{}, err
	}
	return noticia, nil
}
