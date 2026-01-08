# Image Warehousing Tool

AI-powered image warehousing system for artists with automatic categorization and semantic search using Google Gemini.

## Features

✅ **2D Image Support**: Upload photos, paintings, concept art

✅ **3D Object Support**: Upload 3D models with 6-view analysis (front, back, left, right, top, bottom)

✅ **AI-Powered Categorization**: Gemini automatically analyzes and categorizes uploads

✅ **Semantic Search**: Natural language queries ("dark cat image" finds both "black cat" and "cat at night")

✅ **Filesystem-Based**: No database needed - organized folders + markdown index

✅ **Async Processing**: Fast upload response with background AI analysis

✅ **RESTful API**: HTTP endpoints for all operations

## Quick Start

### 1. Prerequisites

- Go 1.22 or higher
- Google Gemini API key ([Get one here](https://aistudio.google.com/app/apikey))

### 2. Setup

```bash
# Navigate to project
cd image-warehousing

# Copy environment file
copy .env.example .env
# Edit .env and add your GEMINI_API_KEY

# Install dependencies
go mod download
```

### 3. Run

```bash
make run
# OR
go run cmd/server/main.go
```

Server starts on `http://localhost:8080`

## API Examples

### Upload 2D Image
```bash
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "image=@photo.jpg" \
  -F "title=Mountain Landscape" \
  -F "artist=John Doe" \
  -F "tags=[\"landscape\",\"outdoor\"]"
```

### Upload 3D Object (6 views)
```bash
curl -X POST http://localhost:8080/api/v1/images/upload-3d \
  -F "front=@model_front.jpg" \
  -F "back=@model_back.jpg" \
  -F "left=@model_left.jpg" \
  -F "right=@model_right.jpg" \
  -F "top=@model_top.jpg" \
  -F "bottom=@model_bottom.jpg" \
  -F "title=Robot Character" \
  -F "artist=Jane Smith"
```

### Search
```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"query": "dark cat image", "limit": 10}'
```

## How It Works

1. **Upload** → Image saved to temp, immediate response
2. **Background Worker** → Gemini analyzes image
3. **Auto-Categorize** → Moves to category folder (e.g., `animals/cats/`)
4. **Index** → Adds metadata to `data/index.md`
5. **Search** → Gemini performs semantic search on index

## Project Structure

```
data/categories/
  animals/cats/550e8400...jpg
  landscapes/mountains/660e8400...jpg
  3d-renders/characters/770e8400.../
    front.jpg, back.jpg, ...
data/index.md  # Searchable markdown index
```

## Configuration (.env)

```bash
# Server Configuration
SERVER_PORT=8080

# Gemini AI Configuration
GEMINI_API_KEY=your_api_key_here
# Available models:
#   - gemini-3-flash (default): Fast, cost-effective, excellent vision
#   - gemini-3-pro: Best-in-class vision analysis, higher accuracy & cost
GEMINI_MODEL=gemini-3-flash

# Storage Configuration
DATA_DIR=./data
MAX_UPLOAD_SIZE=52428800  # 50MB

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000
```

## Commands

```bash
make run        # Run server
make build      # Build binary
make test       # Run tests
make dev-setup  # Complete setup
```

## Testing

### Run All Tests
```bash
make test
```

### Test Coverage
- **40 tests** covering services, handlers, and integration workflows
- **Storage Service**: File operations, thumbnails, image dimensions
- **Index Service**: Markdown generation, file locking, persistence
- **Search Service**: Result limiting, validation, query handling
- **Integration Tests**: End-to-end indexing and search workflows

### Integration Tests
The integration tests demonstrate the complete workflow:
- **Indexing Workflow**: Upload → Analyze → Index → Persist
- **Search Integration**: Populate index → Search by tags/description
- **Persistence**: Index survives service restarts

```bash
# Run specific test suites
go test -v ./internal/service -run TestIndexing     # Indexing tests
go test -v ./internal/service -run Integration     # Integration tests
go test -v ./internal/api/handlers                  # Handler tests
```

## Documentation

- `plan.md` - Future pipeline plan (2D→3D generation with Gemini Pro + Tripo)
- `README.md` - This file
- `.env.example` - Configuration template
