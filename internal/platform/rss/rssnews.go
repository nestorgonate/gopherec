package rss

import (
	"context"
	"gopherec/internal/domain/entity"
	"gopherec/scripts"
	"time"

	"github.com/mmcdole/gofeed"
)

type RSS struct {
	fp gofeed.Parser
}

func NewNoticias() *RSS {
	return &RSS{}
}

func (n *RSS) GetPolitics(c context.Context) ([]entity.Noticia, error) {
	var noticia []entity.Noticia
	feed, err := n.fp.ParseURLWithContext("https://www.primicias.ec/rss/politica.xml", c)
	if err != nil {
		return []entity.Noticia{}, err
	}
	if len(feed.Items) == 0 {
		return []entity.Noticia{}, entity.ErrNoNewNewsItem
	}
	for _, item := range feed.Items {
		cleanContent := scripts.CleanHtml(item.Content)
		var published time.Time
		published = scripts.ParsePublished(item.Published)
		noticia = append(noticia, entity.Noticia{
			Title:       item.Title,
			Description: item.Description,
			Content:     cleanContent,
			Link:        item.Link,
			Category:    entity.Politica,
			Status:      entity.Pending,
			Published:   published,
		})
	}
	return noticia, nil
}
