# AstroVista API

API for managing NASA APOD (Astronomy Picture of the Day) data with advanced features including interactive documentation, cache system, and support for multiple languages.

## Features

-   **Interactive documentation** with Swagger/OpenAPI
-   **Cache system** with Redis for better performance
-   **Internationalization (i18n)** with support for multiple languages - English (default)
    -   Brazilian Portuguese
    -   Spanish
    -   French
    -   German
    -   Italian
    -   Support for easily adding new languages

## Configuration

### Requirements

-   Go 1.18 or higher
-   MongoDB (for data storage)
-   Redis (optional, for caching)

### Environment Variables

-   `PORT` - Server port (default: 8081)
-   `MONGODB_URI` - MongoDB connection URI
-   `REDIS_URL` - Redis server URL (optional)
-   `REDIS_PASSWORD` - Redis password (optional)
-   `GOOGLE_TRANSLATE_API_KEY` - Google Translate API key (optional)
-   `DEEPL_API_KEY` - DeepL API key for translations (optional)

## Translation Services

The API can use different translation services:

### Google Translate

To use Google Translate, you need to:

1. Create an account on [Google Cloud Platform](https://cloud.google.com/)
2. Create a new project
3. Enable the Cloud Translation API
4. Create an API key
5. Set the environment variable `GOOGLE_TRANSLATE_API_KEY`

```bash
export GOOGLE_TRANSLATE_API_KEY="your-key-here"
```

### DeepL

To use DeepL, you need to:

1. Create an account on [DeepL API](https://www.deepl.com/pro-api)
2. Get your authentication key
3. Set the environment variable `DEEPL_API_KEY`

```bash
export DEEPL_API_KEY="your-key-here"
```

### Mock

If no API key is configured, the API will use a mock translation service for development.

## Translation Cache System

To improve performance and avoid repeated calls to translation APIs, we implemented a two-level cache system:

1. **In-memory cache**: Stores recent translations in memory for quick access
2. **Redis cache**: If Redis is available, translations are also stored persistently

Translations are stored for 30 days in the Redis cache, significantly reducing the number of calls to external APIs.

## Running the API

```bash
# Clone the repository
git clone https://github.com/your-username/astrovista-api.git
cd astrovista-api

# Install dependencies
go get -u

# Run the API
go run main.go
```

## Endpoints

### Documentation

-   `/swagger/` - Swagger interactive documentation

### Main endpoints

-   `GET /apod` - Get the most recent APOD
-   `GET /apod/{date}` - Get an APOD by specific date
-   `GET /apods` - List all registered APODs
-   `GET /apods/search` - Advanced search with filters
-   `GET /apods/date-range` - Search APODs by date range
-   `GET /languages` - List supported languages
-   `POST /apod` - Add a new APOD

## Language Support

To get responses in a specific language, you can:

1. Send the `Accept-Language` header in the request

    ```
    Accept-Language: pt-BR
    ```

2. Or add the `lang` parameter in the URL
    ```
    /apod?lang=pt-BR
    ```

## Cache

API responses include the `X-Cache` header to indicate if the result came from cache:

-   `X-Cache: HIT` - Response retrieved from cache
-   `X-Cache: MISS` - Response obtained from the database

## License

MIT
