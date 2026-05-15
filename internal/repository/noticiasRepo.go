package repository

import (
	"context"
	"errors"
	"gopherec/internal/domain/entity"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (n NoticiasRepo) Save(c context.Context, noticias ...entity.Noticia) (uint, primitive.ObjectID, error) {
	collection := n.db.Collection("noticias")
	var countNew uint = 0
	var id primitive.ObjectID
	for _, noticia := range noticias {
		filter := bson.M{"link": noticia.Link}
		// El comando $set reemplaza los campos, $setOnInsert solo se ejecuta si es nueva
		update := bson.M{
			"$setOnInsert": noticia,
		}
		opts := options.Update().SetUpsert(true)
		result, err := collection.UpdateOne(c, filter, update, opts)
		if err != nil {
			return 0, primitive.ObjectID{}, nil
		}
		if result.UpsertedCount > 0 {
			countNew++
			id = result.UpsertedID.(primitive.ObjectID)
		}
	}
	return countNew, id, nil
}

func (n NoticiasRepo) GetPending(c context.Context) (entity.Noticia, error) {
	collection := n.db.Collection("noticias")
	count, _ := collection.CountDocuments(c, bson.M{"$in": []string{"pending", "processing"}})
	log.Printf("DEBUG: Documentos con status Pending o Processing encontrados: %d", count)
	filter := bson.M{"status": bson.M{"$in": []string{"pending"}}}
	update := bson.M{"$set": bson.M{"status": "processing"}}
	//Opciones: Ordenar por fecha de publicación (la más reciente primero)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetSort(bson.M{"published": -1})
	var noticia entity.Noticia
	err := collection.FindOneAndUpdate(c, filter, update, opts).Decode(&noticia)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entity.Noticia{}, entity.ErrNotPendingOrProcessingNews
		}
		log.Printf("DEBUG: Error al decodificar en Struct: %v", err)
		return entity.Noticia{}, err
	}
	log.Printf("DEBUG: Noticia de la base de datos: %s, Status actual: %s\n", noticia.ID.Hex(), noticia.Status)
	return noticia, nil
}

func (n NoticiasRepo) Update(c context.Context, noticiaId primitive.ObjectID, fieldsUpdate map[string]any) error {
	collection := n.db.Collection("noticias")
	if len(fieldsUpdate) == 0 {
		return nil
	}
	update := bson.M{
		"$set": fieldsUpdate,
	}
	filter := bson.M{"_id": noticiaId}
	result, err := collection.UpdateOne(c, filter, update)
	if err != nil {
		log.Printf("ERROR: No se pudo actualizar la noticia: %v\n", err)
		return err
	}
	if result.MatchedCount == 0 {
		log.Println("ERROR: No se encontro el ID de la noticia para actualizar el sensitivityLevel")
		return nil
	}
	for key, _ := range fieldsUpdate {
		log.Printf("DEBUG: Se actualizo el %v de %d noticias\n", key, result.ModifiedCount)
	}
	return nil
}

func (n NoticiasRepo) Delete(c context.Context, noticiaId primitive.ObjectID) error {
	collection := n.db.Collection("noticias")
	filter := bson.M{"_id": noticiaId}
	result, err := collection.DeleteOne(c, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		log.Println("ERROR: No se encontro el ID de la noticia para borrar")
		return nil
	}
	log.Printf("DEBUG: Borrando %d noticias\n", result.DeletedCount)
	return nil
}
