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
	opinionesService := service.NewOpinionService(geminiClient, noticiasRepo, deepseekClient)
	twitterClient := twitter.NewTwitter()
	alertaNoticia := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		log.Println("Obtener noticias al empezar el bot")
		thereNews, _ := noticiasService.Get(c)
		if thereNews {
			alertaNoticia <- true
		}
		for {
			select {
			case <-c.Done():
				return
			case <-ticker.C:
				log.Println("Obteniendo noticias")
				thereNews, err := noticiasService.Get(c)
				if err != nil {
					log.Printf("No se pudo obtener noticias: %v", err)
					continue
				}
				if thereNews {
					select {
					case alertaNoticia <- true:
					default:
					}
				}
			}
		}
	}()

	go func() {
		for range alertaNoticia {
			log.Println("Opinando de noticias pendientes en la base de datos")
			opinion, err := opinionesService.GenerateOpinion(c)
			if err != nil {
				log.Printf("Error al obtener opinion: %v", err)
				continue
			}
			log.Printf("Opinion: %v\n", opinion)
			postId, err := twitterClient.Post(c, opinion)
			if err != nil {
				log.Printf("Error al realizar un nuevo post: %v", err)
				continue
			}
			log.Printf("Post realizado: %v\n", postId)
			time.Sleep(1 * time.Minute)
		}
	}()
	select {}
}
