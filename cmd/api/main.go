package main

import (
	"context"
	"gopherec/internal/platform/llm/deepseek"
	"gopherec/internal/platform/llm/gemini"
	mongodbplatform "gopherec/internal/platform/mongodb"
	"gopherec/internal/platform/rss"
	"gopherec/internal/platform/twitter"
	"gopherec/internal/repository"
	"gopherec/internal/service"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	c := context.Background()
	_ = godotenv.Load()
	mongoDB := mongodbplatform.NewMongoDB(1 * time.Minute)
	mongoDB.Connect(c)
	geminiClient := gemini.NewGemini()
	if err := geminiClient.NewClient(c); err != nil {
		log.Fatalf("Fallo crítico en Gemini: %v", err)
	}
	deepseekClient := deepseek.NewDeepSeek()
	getRss := rss.NewNoticias()
	noticiasRepo := repository.NewNoticiasRepo(mongoDB.Db)
	noticiasService := service.NewNoticiasService(geminiClient, getRss, noticiasRepo, deepseekClient)
	twitterClient := twitter.NewTwitter()
	opinionesService := service.NewOpinionService(geminiClient, noticiasRepo, deepseekClient, twitterClient)
	go func() {
		ticker := time.NewTicker(2 * time.Hour)
		defer ticker.Stop()
		obtenerNoticias := func() {
			log.Println("Actualizando noticias")
			thereNews, err := noticiasService.Get(c)
			if err != nil {
				log.Printf("No se pudo obtener noticias: %v", err)
			}
			if !thereNews {
				log.Println("No hay noticias nuevas")
				return
			}
			log.Println("Opinando de noticias pendientes en la base de datos")
			opinionesService.GenerateOpinion(c)
		}
		log.Println("Obteniendo noticias al empezar el bot")
		obtenerNoticias()
		for {
			select {
			case <-c.Done():
				log.Println("Cerrando monitor de noticias")
				return
			case <-ticker.C:
				obtenerNoticias()
			}
		}
	}()
	select {}
}
