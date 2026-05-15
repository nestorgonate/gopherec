# Planeación futura

## Contexto histórico con búsqueda vectorial

### Situación actual

Hoy, Gopherec genera opiniones sobre noticias sin ningún contexto histórico. El LLM opina únicamente con su conocimiento interno de entrenamiento, lo que produce opiniones genéricas que no aprovechan la rica (y a menudo repetitiva) historia política ecuatoriana.

### La visión

Queremos que el bot pueda **referenciar hechos históricos reales** al opinar, haciendo comentarios como:

> "Otra vez el mismo cuento, como en el 2005 cuando... #Ecuador"

Para lograrlo, necesitamos un pipeline que permita:

1. **Ingerir hechos históricos** de Ecuador (golpes de estado, crisis bancarias, casos de corrupción, etc.)
2. **Generar embeddings vectoriales** de cada hecho
3. **Almacenarlos** en MongoDB Atlas con un índice de búsqueda vectorial
4. **Consultar hechos similares** al recibir una noticia nueva
5. **Pasar esos hechos como referencia** al prompt de generación de opinión

### Estado actual en el código

Parte de la infraestructura ya está preparada:

- **Entidad `Historia`** (`internal/domain/entity/historia.go`): ya definida con campos para título, contenido, vector embebido, categoría y año
- **Índice de búsqueda**: el nombre `idx_vectorHistory` ya está referenciado en el código
- **Método `SearchHistory`**: ya implementado en `noticiasRepo.go`, usa `$vectorSearch` con filtro por categoría

```go
func (n NoticiasRepo) SearchHistory(c context.Context, vector []float64, category string) ([]bson.M, error) {
    pipeline := mongo.Pipeline{
        {{Key: "$vectorSearch", Value: bson.D{
            {Key: "index", Value: "idx_vectorHistory"},
            {Key: "path", Value: "vectorContent"},
            {Key: "queryVector", Value: vector},
            {Key: "numCandidates", Value: 100},
            {Key: "limit", Value: 5},
            {Key: "filter", Value: bson.D{{Key: "category", Value: category}}},
        }}},
    }
    // ...
}
```

### Lo que falta por definir

#### 1. Mecanismo de ingesta

Aún no está determinado cómo se poblará la colección `historia_ecuador`. Hay dos alternativas en consideración:

| Opción | Descripción | Ventajas | Desventajas |
|--------|-------------|----------|-------------|
| **Endpoint REST** | Un servidor HTTP dentro del bot (o separado) que recibe hechos históricos vía API | Control explícito, validación, idempotencia | Mayor complejidad operativa |
| **Script automatizado** | Un script que lea desde un archivo JSON/CSV y los inserte directamente | Simple, fácil de mantener | Menos flexible, requiere despliegue manual |

#### 2. Generación de embeddings

Para poblar `vectorContent` se necesita un modelo de embeddings. Opciones:

- Usar **Gemini Embedding API** (consistente con el stack actual)
- Usar un modelo open-source local
- Usar **MongoDB Atlas Embeddings** (integrado con el servicio de búsqueda)

#### 3. Integración con el pipeline de opinión

Actualmente `OpinionPrompt` recibe un string `referencia` que siempre es:

```
"No hay referencias historicas aun, utiliza el conocimiento con el que fuiste entrenado"
```

El plan es reemplazar esto con una llamada real a `SearchHistory`:

```go
func (service *OpinionService) GenerateOpinion(c context.Context) {
    // ...
    referencias, _ := service.repo.SearchHistory(c, embeddingNoticia, noticia.Category)
    referencia := formatearReferencias(referencias)
    opinion, err := service.gemini.GenerateOpinion(c, noticia, referencia)
    // ...
}
```

### Preguntas abiertas para la comunidad

1. **¿REST endpoint o script automatizado?** ¿Cuál prefieres para la ingesta de datos históricos?
2. **¿Qué modelo de embeddings usar?** ¿Gemini, open-source, o el integrado de Atlas?
3. **¿Fuentes de datos históricos?** ¿Conoces datasets o archivos con la historia ecuatoriana que podamos usar?
4. **¿Autenticación del endpoint?** Si hacemos REST, ¿debería tener API key o estar abierto?

### Roadmap tentativo

| Fase | Descripción | Estado |
|------|-------------|--------|
| 1 | Definir modelo de datos y crear colección `historia_ecuador` | ✅ Hecho |
| 2 | Crear índice `idx_vectorHistory` en Atlas Search | Pendiente |
| 3 | Implementar ingesta de datos (endpoint o script) | Por definir |
| 4 | Integrar `SearchHistory` en el pipeline de opinión | Pendiente |
| 5 | Probar y afinar prompts con contexto histórico real | Pendiente |
