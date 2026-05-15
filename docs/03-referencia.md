# Referencia técnica

## Variables de entorno

Todas las variables se cargan desde `cmd/api/.env` mediante `godotenv`.

| Variable | Requerida | Descripción |
|----------|-----------|-------------|
| `DATABASE_URL` | Sí | URI de conexión a MongoDB Atlas |
| `GEMINI_API_KEY` | Sí | API key de Google Gemini |
| `DEEPSEEK_API_KEY` | Sí | API key de DeepSeek (respaldo) |
| `GOTWI_API_KEY` | Sí | API Key de la app de X/Twitter |
| `GOTWI_API_KEY_SECRET` | Sí | API Key Secret de la app de X/Twitter |
| `GOTWI_ACCESS_TOKEN` | Sí | Access Token de X/Twitter (OAuth 1.0a) |
| `GOTWI_ACCESS_TOKEN_SECRET` | Sí | Access Token Secret de X/Twitter (OAuth 1.0a) |

## Estructura del proyecto

```
gopherec/
├── cmd/api/
│   ├── main.go           # Punto de entrada, orquestación e inicialización
│   └── .env              # Variables de entorno (fuera de git)
├── internal/
│   ├── domain/           # Interfaces y entidades del dominio
│   │   ├── entity/
│   │   │   ├── noticia.go      # Struct Noticia, estados, errores
│   │   │   ├── categoria.go    # Enum Categoria, struct Clasificacion
│   │   │   └── historia.go     # Struct Historia (para búsqueda vectorial)
│   │   ├── llm.go              # Interfaz LLMProvider
│   │   ├── rss.go              # Interfaz RSS
│   │   ├── twitter.go          # Interfaz TwitterAPI
│   │   └── repositorio.go      # Interfaz NoticiasRepo
│   ├── platform/
│   │   ├── llm/
│   │   │   ├── prompts.go           # Prompts de sistema y generación
│   │   │   ├── gemini/gemini.go     # Implementación con Gemini
│   │   │   └── deepseek/deepseek.go # Implementación con DeepSeek
│   │   ├── rss/rssnews.go           # Parser RSS (gofeed)
│   │   ├── mongodb/mongodb.go       # Conexión a MongoDB
│   │   └── twitter/api.go           # Cliente X/Twitter (gotwi)
│   ├── service/
│   │   ├── noticias.go    # Obtener y clasificar noticias
│   │   └── opinion.go     # Generar opiniones y publicar
│   └── repository/
│       └── noticiasRepo.go # Operaciones CRUD en MongoDB
├── scripts/
│   └── utils.go           # Utilidades: limpiar HTML, parsear fechas
├── Dockerfile
└── docker-compose.yml
```

## Interfaces del dominio

### LLMProvider

```go
type LLMProvider interface {
    GenerateOpinion(c context.Context, noticia entity.Noticia, referencia string) (string, error)
    Categorize(c context.Context, noticia entity.Noticia) (entity.Clasificacion, error)
}
```

### RSS

```go
type RSS interface {
    GetPolitics(c context.Context) ([]entity.Noticia, error)
}
```

### TwitterAPI

```go
type TwitterAPI interface {
    Post(c context.Context, text string) (string, error)
}
```

### NoticiasRepo

```go
type NoticiasRepo interface {
    SearchHistory(c context.Context, vector []float64, category string) ([]bson.M, error)
    Save(c context.Context, noticias ...entity.Noticia) (uint, primitive.ObjectID, error)
    GetPending(c context.Context) (entity.Noticia, error)
    Update(c context.Context, noticiaId primitive.ObjectID, fieldsUpdate map[string]any) error
    Delete(c context.Context, noticiaId primitive.ObjectID) error
}
```

## Modelos de datos

### Noticia (colección `noticias`)

| Campo | Tipo Go | Tipo MongoDB | Descripción |
|-------|---------|--------------|-------------|
| `ID` | `primitive.ObjectID` | `ObjectId` | Identificador único |
| `Title` | `string` | `string` | Título |
| `Description` | `string` | `string` | Resumen |
| `Content` | `string` | `string` | Contenido (HTML limpiado) |
| `Link` | `string` | `string` | URL original |
| `Category` | `Categoria` | `string` | Categoría asignada |
| `Status` | `StatusEnum` | `string` | Estado (`pending`, `processing`, `published`, `rejected`) |
| `SensitivityLevel` | `int64` | `int64` | Sensibilidad (1-10) |
| `Published` | `time.Time` | `Date` | Fecha de publicación |

### Clasificacion

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `Category` | `Categoria` | Categoría (`politica`, `economia`, `inseguridad`, `sensible`, `otros`) |
| `SensitivityLevel` | `int64` | Nivel de sensibilidad (1-10) |

### Historia (colección `historia_ecuador`, planeado)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `ID` | `primitive.ObjectID` | Identificador único |
| `Title` | `string` | Título del hecho histórico |
| `Content` | `string` | Descripción |
| `VectorContent` | `[]float64` | Embedding para búsqueda semántica |
| `Category` | `Categoria` | Categoría relacionada |
| `Year` | `int` | Año del suceso |

## Categorías y reglas de filtrado

Cuando el LLM clasifica una noticia, se aplican estas reglas:

| Categoría | Condición | Acción |
|-----------|-----------|--------|
| `sensible` | Siempre | Se rechaza (`status: rejected`) |
| `otros` | `sensitivityLevel ≤ 8` | Se elimina de la BD |
| `politica` | `sensitivityLevel < 7` | Se elimina de la BD |
| Cualquier otra | `sensitivityLevel ≥ umbral` | Se conserva y se publica |

## Modelos de LLM

| Operación | Primario | Respaldo |
|-----------|----------|----------|
| Clasificación | `gemini-2.5-flash` (temp: 0.1) | `deepseek-chat` (temp: 0.1) |
| Opinión | `gemini-3-flash-preview` (temp: 1.0) | `deepseek-chat` (temp: 0.9, max: 280 tokens) |

## Dependencias externas

| Dependencia | Propósito |
|-------------|-----------|
| `github.com/mmcdole/gofeed` | Parseo de feeds RSS/Atom |
| `github.com/michimani/gotwi` | Cliente para la API de X/Twitter |
| `google.golang.org/genai` | SDK oficial de Google AI (Gemini) |
| `github.com/sashabaranov/go-openai` | Cliente OpenAI-compatible (DeepSeek) |
| `go.mongodb.org/mongo-driver` | Driver oficial de MongoDB |
| `github.com/joho/godotenv` | Carga de variables de entorno desde `.env` |

## Prompts del sistema

### Clasificador (`InstructionClasifier`)

Sistema experto que categoriza noticias ecuatorianas en 5 categorías y asigna un nivel de sensibilidad del 1 al 10. Responde únicamente con JSON.

### Opinión (`InstructionOpinion`)

Persona: ciudadano ecuatoriano de clase media, indignado y sarcástico. Escribe tweets de máximo 280 caracteres en español ecuatoriano, sin emojis, con máximo un hashtag. El tono varía según el nivel de sensibilidad:

- **≥ 8**: Derrota absoluta, decepción
- **≥ 5**: Indignación activa
- **< 5**: Burlón y sarcástico
