package mongodbplatform

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	Db      *mongo.Database
	Client  *mongo.Client // Guardamos el cliente por si necesitas sesiones o transacciones
	Timeout time.Duration
}

func NewMongoDB(timeout time.Duration) *MongoDB {
	return &MongoDB{
		Timeout: timeout,
	}
}

func (db *MongoDB) Connect(c context.Context) {
	url := os.Getenv("DATABASE_URL")
	clientOptions := options.Client().ApplyURI(url)
	var err error
	client, err := mongo.Connect(c, clientOptions)
	if err != nil {
		log.Fatalf("No se pudo conectar a MongoDB: %v\n", err)
	}
	ping, cancel := context.WithTimeout(c, db.Timeout)
	defer cancel()
	err = client.Ping(ping, readpref.Primary())
	if err != nil {
		log.Fatalf("No se pudo conectar a MongoDB: %v\n", err)
	}
	db.Client = client
	db.Db = client.Database("gopherec")
}
