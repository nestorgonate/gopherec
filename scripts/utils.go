package scripts

import (
	"html"
	"log"
	"regexp"
	"strings"
	"time"
)

func CleanHtml(src string) string {
	re := regexp.MustCompile("<[^>]*>")
	src = re.ReplaceAllString(src, " ")
	src = html.UnescapeString(src)
	src = strings.Join(strings.Fields(src), " ")
	src = strings.TrimSpace(src)
	src = strings.TrimSuffix(src, "Leer más ]]>")
	return src
}

func ParsePublished(src string) time.Time {
	published, err := time.Parse(time.RFC1123Z, src)
	if err != nil {
		log.Printf("No se pudo parsear la fecha de publicacion: %s", err)
		return time.Now()
	}
	return published
}
