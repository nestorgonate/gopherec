# Guía de instalación

Esta guía te muestra paso a paso cómo configurar y ejecutar Gopherec.

## Requisitos previos

- **Go 1.26+** (si ejecutas sin Docker)
- **Docker y Docker Compose** (opcional, alternativa recomendada)
- **Conexión a internet** para acceder a RSS, APIs externas y MongoDB Atlas

### Cuentas externas necesarias

| Servicio | Propósito | Enlace |
|----------|-----------|--------|
| MongoDB Atlas | Base de datos (incluye Vector Search) | [mongodb.com/atlas](https://www.mongodb.com/atlas) |
| Google AI Studio | API Key de Gemini | [aistudio.google.com](https://aistudio.google.com/) |
| DeepSeek | API Key de DeepSeek (respaldo) | [platform.deepseek.com](https://platform.deepseek.com/) |
| X/Twitter Developer | OAuth 1.0a para publicar tweets | [developer.x.com](https://developer.x.com/) |

---

## Paso 1: Clonar el repositorio

```bash
git clone https://github.com/tuusuario/gopherec
cd gopherec
```

## Paso 2: Configurar variables de entorno

Copia el archivo de ejemplo:

```bash
cp cmd/api/.env.example cmd/api/.env
```

Edita `cmd/api/.env` con tus credenciales:

```env
DATABASE_URL=mongodb+srv://usuario:password@cluster.mongodb.net/gopherec?retryWrites=true&w=majority
GEMINI_API_KEY=AIzaSy...
DEEPSEEK_API_KEY=sk-...
GOTWI_API_KEY=tu_api_key_de_twitter
GOTWI_API_KEY_SECRET=tu_api_key_secret
GOTWI_ACCESS_TOKEN=tu_access_token
GOTWI_ACCESS_TOKEN_SECRET=tu_access_token_secret
```

> **⚠️ Importante:** Nunca subas el archivo `.env` al repositorio. Está incluido en `.gitignore`.

### Cómo obtener cada credencial

<details>
<summary><strong>MongoDB Atlas</strong></summary>

1. Crea una cuenta en [MongoDB Atlas](https://www.mongodb.com/atlas)
2. Crea un clúster gratuito (M0)
3. Ve a **Database Access** y crea un usuario con contraseña
4. Ve a **Network Access** y agrega tu IP (o `0.0.0.0/0` para desarrollo)
5. Haz clic en **Connect** → **Drivers** y copia la URI de conexión
6. Reemplaza `<password>` con la contraseña de tu usuario

La URI se ve así:
```
mongodb+srv://<usuario>:<password>@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority
```
</details>

<details>
<summary><strong>Gemini API Key</strong></summary>

1. Ve a [Google AI Studio](https://aistudio.google.com/)
2. Inicia sesión con tu cuenta de Google
3. Haz clic en **Get API Key** en la barra lateral
4. Crea una nueva API key
5. Cópiala y pégala en `GEMINI_API_KEY`

La clave empieza con `AIzaSy...`.
</details>

<details>
<summary><strong>DeepSeek API Key</strong></summary>

1. Ve a [platform.deepseek.com](https://platform.deepseek.com/)
2. Regístrate e inicia sesión
3. Ve a **API Keys** y crea una nueva clave
4. Cópiala y pégala en `DEEPSEEK_API_KEY`

La clave empieza con `sk-...`.
</details>

<details>
<summary><strong>X / Twitter API (OAuth 1.0a)</strong></summary>

1. Ve a [developer.x.com](https://developer.x.com/) y crea una cuenta de desarrollador
2. Crea un proyecto y una app
3. En la configuración de la app, ve a **User Authentication Settings**
4. Configura **OAuth 1.0a** con permisos de lectura y escritura
5. Copia las siguientes credenciales:
   - **API Key** → `GOTWI_API_KEY`
   - **API Key Secret** → `GOTWI_API_KEY_SECRET`
   - **Access Token** → `GOTWI_ACCESS_TOKEN`
   - **Access Token Secret** → `GOTWI_ACCESS_TOKEN_SECRET`

**Nota:** El Access Token debe tener permisos de escritura (Read + Write).
</details>

## Paso 3: Configurar MongoDB Atlas Vector Search

Para que la funcionalidad de búsqueda histórica funcione (cuando esté implementada), necesitas crear un índice de búsqueda vectorial en Atlas:

1. En MongoDB Atlas, ve a tu clúster
2. Haz clic en **Search** → **Create Search Index**
3. Selecciona **Atlas Vector Search** → **JSON Editor**
4. Usa esta configuración:

```json
{
  "name": "idx_vectorHistory",
  "type": "vectorSearch",
  "fields": [
    {
      "type": "vector",
      "path": "vectorContent",
      "numDimensions": 768,
      "similarity": "cosine"
    }
  ]
}
```

5. Selecciona la base de datos `gopherec` y la colección `historia_ecuador`
6. Haz clic en **Create Index**

> **Nota:** Este índice se usa para el método `SearchHistory()` que actualmente no está en uso. Se implementará en una versión futura (ver [Planeación futura](04-planeacion-futura.md)).

## Paso 4: Ejecutar el bot

### Opción A: Con Docker (recomendada)

```bash
docker compose up
```

Esto construye la imagen y ejecuta el bot. Los logs se mostrarán en la terminal.

### Opción B: Directamente con Go

```bash
cd cmd/api
go run main.go
```

## Verificar que funciona

Cuando el bot se ejecute correctamente, verás logs como estos:

```
Obteniendo noticias al empezar el bot
Actualizando noticias
Obtieniendo noticias RSS
Gemini esta categorizando una noticia
Clasificacion de Gemini: {"category": "politica", "sensitivityLevel": 7}
Se agregaron a la base de datos: 1 noticias
Opinando de noticias pendientes en la base de datos
Gemini esta opinando sobre una noticia
Nuevo post: 1234567890
```

Si Gemini falla, verás:

```
Gemini falló: ... Reintentando con DeepSeek...
DeepSeek está categorizando una noticia
```

## Solución de problemas

| Problema | Posible causa | Solución |
|----------|---------------|----------|
| `No se pudo conectar a MongoDB` | URI incorrecta o IP no autorizada | Verifica `DATABASE_URL` y la whitelist de IPs en Atlas |
| `Fallo crítico en Gemini` | API Key inválida o sin saldo | Verifica `GEMINI_API_KEY` |
| `DeepSeek no devolvió candidatos` | API Key inválida o sin saldo | Verifica `DEEPSEEK_API_KEY` |
| `No hay noticias nuevas` | No hay nuevos artículos en el RSS | Espera a que publiquen nuevas noticias |
