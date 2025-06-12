<div align="center">
  <img src="https://github.com/user-attachments/assets/74564f36-a574-4f84-91d0-a7c98a29a14f" alt="AstroVista Logo">
  <br>
  <p>
    <strong>Explore the cosmos with NASA's Astronomy Picture of the Day data</strong>
  </p>
  <p>
    <a href="#features">Features</a> •
    <a href="#api-reference">API Reference</a> •
    <a href="#getting-started">Getting Started</a> •
    <a href="#internationalization">Internationalization</a> •
    <a href="#caching">Caching</a> •
    <a href="#examples">Examples</a>
  </p>  <p>
    <img alt="Version" src="https://img.shields.io/badge/version-1.0.0-blue.svg" />
    <img alt="Go Version" src="https://img.shields.io/badge/Go-1.24.3-00ADD8.svg" />
    <img alt="License" src="https://img.shields.io/badge/License-MIT-green.svg" />
  </p>
</div>

## Overview

AstroVista API is a high-performance RESTful service that provides access to NASA's Astronomy Picture of the Day (APOD) data. Built with Go, it delivers astronomy images and data with multilingual support, advanced caching, and comprehensive search capabilities.

Whether you're building an astronomy application, educational platform, or simply want to incorporate stunning space imagery into your project, AstroVista API provides a reliable and developer-friendly solution.

## Features

-   **🚀 High Performance**: Fast response times with multi-level caching
-   **🌍 Full Internationalization**: Support for 22 languages and growing
-   **📚 Rich Metadata**: Detailed information about each astronomy image
-   **🔍 Advanced Search**: Find images by date, media type, or content
-   **📊 Pagination**: Control your data consumption with flexible paging
-   **📄 Interactive Documentation**: Thoroughly documented API with live testing
-   **🛡️ Rate Limiting**: Protection against excessive usage
-   **💾 Persistent Storage**: MongoDB-backed data persistence
-   **🧠 Smart Caching**: Redis and in-memory caching for optimal performance

## Technologies Used

-   **Go (Golang) 1.24.3**: The primary programming language for the API.
-   **MongoDB**: Used for persistent storage of APOD data.
-   **Redis**: Utilized as an optional, but highly recommended, caching layer for performance optimization.
-   **Gorilla Mux**: A powerful URL router and dispatcher for building web servers in Go.
-   **Swagger/OpenAPI**: For API documentation and interaction.
-   **Go-i18n**: Library for internationalization.
-   **DeepL API / Google Translate API**: External services for translation (optional).

## API Reference

### Base URL

```
https://your-api-domain.com
```

For local development:

```
http://localhost:8080
```

### Authentication

Most endpoints are publicly accessible. The `POST /apod` endpoint requires authentication using an API token in the header:

```
X-API-Token: your_api_token_here
```

### Response Format

All responses are returned in JSON format with consistent structure:

```json
{
	"_id": "6847c55a5205b7b5dd0ef2a7",
	"title": "Enceladus in True Color",
	"date": "2025-06-10",
	"hdurl": "https://apod.nasa.gov/apod/image/2506/EnceladusTrue_Cassini_960.jpg",
	"url": "https://apod.nasa.gov/apod/image/2506/EnceladusTrue_Cassini_960.jpg",
	"media_type": "image",
	"service_version": "v1",
	"explanation": "Do the oceans beneath..."
}
```

### Error Responses

Errors follow a standard format:

```json
{
	"error": "Error message",
	"details": "Additional details about the error"
}
```

### Endpoints

#### APOD Operations

| Method | Endpoint            | Description                 |
| ------ | ------------------- | --------------------------- |
| GET    | `/apod`             | Get the most recent APOD    |
| GET    | `/apod/{date}`      | Get APOD for specific date  |
| GET    | `/apods`            | List all APODs              |
| GET    | `/apods/search`     | Search APODs with filters   |
| GET    | `/apods/date-range` | Get APODs within date range |
| POST   | `/apod`             | Fetch and store latest APOD |

#### Configuration

| Method | Endpoint     | Description                   |
| ------ | ------------ | ----------------------------- |
| GET    | `/languages` | List supported languages      |
| GET    | `/swagger/`  | Interactive API documentation |

### Detailed Endpoint Documentation

#### `GET /apod`

Returns the most recent Astronomy Picture of the Day.

**Query Parameters:**

-   `lang` (optional): Desired language code (e.g., `en`, `es`, `zh`)

**Example Request:**

```http
GET /apod?lang=ja
```

**Response:** `200 OK`

```json
{
	"_id": "6847c55a5205b7b5dd0ef2a7",
	"title": "エンケラドゥスの真の姿",
	"date": "2025-06-10",
	"hdurl": "https://apod.nasa.gov/apod/image/2506/EnceladusTrue_Cassini_960.jpg",
	"url": "https://apod.nasa.gov/apod/image/2506/EnceladusTrue_Cassini_960.jpg",
	"media_type": "image",
	"service_version": "v1",
	"explanation": "土星の衛星エンケラドゥスの氷の下の海には生命が存在するのだろうか？..."
}
```

**Cache Duration:** 1 hour

#### `GET /apod/{date}`

Retrieves an APOD for a specific date.

**Path Parameters:**

-   `date`: Date in `YYYY-MM-DD` format (e.g., `2023-01-15`)

**Query Parameters:**

-   `lang` (optional): Desired language code

**Example Request:**

```http
GET /apod/2023-05-20?lang=fr
```

**Response:** `200 OK`

```json
{
	"_id": "6839c45b8291a3c6dd0ef1b5",
	"title": "Le Système Solaire à l'Échelle",
	"date": "2023-05-20",
	"hdurl": "https://apod.nasa.gov/apod/image/2305/SolarSystem_Scale_2400.jpg",
	"url": "https://apod.nasa.gov/apod/image/2305/SolarSystem_Scale_800.jpg",
	"media_type": "image",
	"service_version": "v1",
	"explanation": "Les planètes dans notre système solaire..."
}
```

**Cache Duration:** 30 days for historical APODs

#### `GET /apods`

Lists all registered Astronomy Pictures of the Day.

**Response:** `200 OK`

```json
{
	"count": 120,
	"items": [
		{
			"_id": "6847c55a5205b7b5dd0ef2a7",
			"title": "Enceladus in True Color",
			"date": "2025-06-10",
			"hdurl": "https://apod.nasa.gov/apod/image/2506/EnceladusTrue_Cassini_960.jpg",
			"url": "https://apod.nasa.gov/apod/image/2506/EnceladusTrue_Cassini_960.jpg",
			"media_type": "image",
			"service_version": "v1",
			"explanation": "Do the oceans beneath Saturn's moon Enceladus contain life?..."
		}
		// Additional items...
	]
}
```

#### `GET /apods/search`

Provides advanced search capabilities with filters, pagination, and sorting.

**Query Parameters:**

-   `page` (optional): Page number (default: 1, min: 1)
-   `perPage` (optional): Items per page (default: 20, range: 1-200)
-   `mediaType` (optional): Filter by media type (`image`, `video`, or `any`)
-   `search` (optional): Text to search in title and explanation fields
-   `startDate` (optional): Start date for filtering (YYYY-MM-DD)
-   `endDate` (optional): End date for filtering (YYYY-MM-DD)
-   `sort` (optional): Sort order (`asc` or `desc` by date, default: `desc`)

**Example Request:**

```http
GET /apods/search?mediaType=image&search=galaxy&page=2&perPage=10
```

**Response:** `200 OK`

```json
{
	"total_results": 45,
	"page": 2,
	"per_page": 10,
	"total_pages": 5,
	"results": [
		{
			"_id": "6823a45b2305c7b4ed0ef1c2",
			"title": "Andromeda Galaxy in Ultraviolet",
			"date": "2025-05-15",
			"hdurl": "https://example.com/andromeda_hd.jpg",
			"url": "https://example.com/andromeda.jpg",
			"media_type": "image",
			"service_version": "v1",
			"explanation": "The Andromeda Galaxy, captured in ultraviolet light..."
		}
		// Additional results...
	]
}
```

**Cache Duration:** 5 minutes

#### `GET /apods/date-range`

Returns APODs within a specified date range.

**Query Parameters:**

-   `start` (optional): Start date (YYYY-MM-DD)
-   `end` (optional): End date (YYYY-MM-DD, defaults to current date)

**Example Request:**

```http
GET /apods/date-range?start=2025-01-01&end=2025-01-31
```

**Response:** `200 OK`

```json
{
	"count": 31,
	"items": [
		{
			"_id": "6814b22a4105d7b5dd0ef1a9",
			"title": "Jupiter's Great Red Spot",
			"date": "2025-01-31",
			"hdurl": "https://example.com/jupiter_hd.jpg",
			"url": "https://example.com/jupiter.jpg",
			"media_type": "image",
			"service_version": "v1",
			"explanation": "Jupiter's Great Red Spot is a persistent anticyclonic storm..."
		}
		// Additional items...
	]
}
```

**Cache Duration:** 12 hours

#### `GET /languages`

Returns a list of all supported languages by the API.

**Response:** `200 OK`

```json
{
  "code": "en",
  "name": "English",
  "nativeName": "English"
},
{
  "code": "ja",
  "name": "Japanese",
  "nativeName": "日本語"
},
// Additional languages...
]
```

#### `POST /apod`

Fetches the most recent APOD from the NASA API and adds it to the database.

**Headers:**

-   `X-API-Token` (required): Internal API token for authorization

**Response Status Codes:**

-   `201 Created`: APOD successfully added
-   `401 Unauthorized`: Invalid API token
-   `409 Conflict`: APOD for the current date already exists
-   `429 Too Many Requests`: Rate limit exceeded (includes `Retry-After` header)

**Rate Limit:** 1 request per minute

## Getting Started

### Prerequisites

Ensure you have the following installed:

-   Go 1.24.3 or higher
-   MongoDB
-   Redis (optional, but recommended for caching)

### Installation

#### Using Docker

```bash
# Pull and run with Docker
docker pull astrovistaorg/api
docker run -p 8080:8080 \
  -e MONGODB_URI="mongodb://localhost:27017" \
  -e MONGODB_DATABASE="apod_db" \
  -e MONGODB_COLLECTION="apods" \
  -e NASA_API_KEY="your_nasa_api_key" \
  astrovistaorg/api
```

#### Manual Setup

1. **Clone the repository**

    ```bash
    git clone https://github.com/astrovistaorg/api.git
    cd api
    ```

2. **Set up environment variables**
   Create a `.env` file:

    ```
    PORT=8080
    MONGODB_URI=mongodb://localhost:27017
    MONGODB_DATABASE=apod_db
    MONGODB_COLLECTION=apods
    NASA_API_KEY=your_nasa_api_key
    INTERNAL_API_TOKEN=your_secure_token
    ```

3. **Install dependencies**

    ```bash
    go get -u
    ```

4. **Run the API**

    ```bash
    go run main.go
    ```

5. **Visit Swagger documentation**
   Open your browser and navigate to:
    ```
    http://localhost:8080/swagger/
    ```

### Configuration

#### Environment Variables

| Variable                   | Description              | Default          | Required |
| -------------------------- | ------------------------ | ---------------- | -------- |
| `PORT`                     | Server port              | `8080`           | No       |
| `MONGODB_URI`              | MongoDB connection URI   |                  | Yes      |
| `MONGODB_DATABASE`         | MongoDB database name    |                  | Yes      |
| `MONGODB_COLLECTION`       | Collection for APODs     |                  | Yes      |
| `REDIS_URL`                | Redis server URL         | `localhost:6379` | No       |
| `REDIS_PASSWORD`           | Redis password           |                  | No       |
| `GOOGLE_TRANSLATE_API_KEY` | Google Translate API key |                  | No       |
| `DEEPL_API_KEY`            | DeepL API key            |                  | No       |
| `NASA_API_KEY`             | NASA API key             | `DEMO_KEY`       | No       |
| `INTERNAL_API_TOKEN`       | Token for POST endpoint  |                  | Yes      |

## Internationalization

AstroVista API provides comprehensive language support for a global audience.

### Supported Languages

The API currently supports 22 languages:

| Code  | Language             | Native Name         |
| ----- | -------------------- | ------------------- |
| en    | English              | English             |
| ar    | Arabic               | العربية             |
| cs    | Czech                | Čeština             |
| de    | German               | Deutsch             |
| es    | Spanish              | Español             |
| fa    | Persian              | فارسی               |
| fr    | French               | Français            |
| hu    | Hungarian            | Magyar              |
| id    | Indonesian           | Bahasa Indonesia    |
| it    | Italian              | Italiano            |
| ja    | Japanese             | 日本語              |
| ko    | Korean               | 한국어              |
| nl    | Dutch/Flemish        | Nederlands          |
| pl    | Polish               | Polski              |
| pt-BR | Brazilian Portuguese | Português do Brasil |
| ro    | Romanian             | Română              |
| ru    | Russian              | Русский             |
| sv    | Swedish              | Svenska             |
| tr    | Turkish              | Türkçe              |
| uk    | Ukrainian            | Українська          |
| vi    | Vietnamese           | Tiếng Việt          |
| zh    | Chinese              | 中文                |

### Setting Language Preference

You can request content in your preferred language through either:

1. The `lang` query parameter (higher priority):

    ```
    GET /apod?lang=es
    ```

2. The `Accept-Language` HTTP header:
    ```
    Accept-Language: es
    ```

If no language is specified or the specified language is not supported, English will be used as the default.

### Translation Services

By default, the API uses Google Translate for dynamic content translation. You can configure:

-   **Google Translate**: Set `GOOGLE_TRANSLATE_API_KEY` environment variable
-   **DeepL**: Set `DEEPL_API_KEY` environment variable

If neither is configured, a mock translation service is used for development.

## Caching

AstroVista API implements a sophisticated caching system to minimize external API calls and database queries.

### Cache Headers

All API responses include an `X-Cache` header:

-   `X-Cache: HIT` - Response was served from cache
-   `X-Cache: MISS` - Response was generated fresh

### Cache Architecture

The caching system operates at two levels:

1. **In-memory Cache**:

    - Stores recent translations
    - Low latency access
    - Persists until server restart

2. **Redis Cache**:
    - Persistent cache storage
    - Default 30-day expiration for translations
    - Functions across server restarts

### Cache Duration by Endpoint

| Endpoint            | Cache Duration |
| ------------------- | -------------- |
| `/apod`             | 1 hour         |
| `/apod/{date}`      | 30 days        |
| `/apods/search`     | 5 minutes      |
| `/apods/date-range` | 12 hours       |

## Examples

### Basic Usage with cURL

Get today's astronomy picture:

```bash
curl -X GET "http://localhost:8080/apod"
```

Get a specific date's picture in Japanese:

```bash
curl -X GET "http://localhost:8080/apod/2025-01-15?lang=ja"
```

Search for galaxy-related images:

```bash
curl -X GET "http://localhost:8080/apods/search?search=galaxy&mediaType=image"
```

### Using the API with JavaScript

```javascript
// Get today's APOD
async function getTodaysApod() {
	const response = await fetch("https://your-api-domain.com/apod")
	const data = await response.json()

	console.log(`Today's astronomy picture: ${data.title}`)
	document.getElementById("apodImage").src = data.url
	document.getElementById("apodDescription").textContent = data.explanation
}

// Search for APODs with "nebula" in different languages
async function searchNebulasInSpanish() {
	const response = await fetch(
		"https://your-api-domain.com/apods/search?search=nebula&lang=es&page=1&perPage=5"
	)
	const data = await response.json()

	console.log(`Found ${data.total_results} results`)
	data.results.forEach((apod) => {
		console.log(`${apod.date}: ${apod.title}`)
	})
}
```

### Using the API with Python

```python
import requests

# Get APODs from a date range
def get_apods_from_january_2025():
    response = requests.get(
        "https://your-api-domain.com/apods/date-range",
        params={"start": "2025-01-01", "end": "2025-01-31", "lang": "fr"}
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Found {data['count']} astronomy pictures")

        for apod in data['items']:
            print(f"{apod['date']}: {apod['title']}")
    else:
        print(f"Error: {response.status_code}")

# Add a new APOD (requires authentication)
def add_latest_apod(api_token):
    headers = {"X-API-Token": api_token}
    response = requests.post("https://your-api-domain.com/apod", headers=headers)

    if response.status_code == 201:
        print("Successfully added latest APOD")
    else:
        print(f"Error: {response.status_code} - {response.json()['error']}")
```

## Contributing

We welcome contributions to the AstroVista API project! Here's how to get started:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

For adding support for a new language, see our [Language Contribution Guide](CONTRIBUTING.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

-   NASA for providing the APOD API
-   The Go community for excellent libraries and tools
-   All contributors who have helped improve this project
