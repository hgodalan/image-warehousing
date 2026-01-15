# Image Warehousing Tool

AI-powered image warehousing system for artists with automatic categorization and semantic search using Google Gemini.

**🌐 Live Demo**: [https://hgodalan.github.io/image-warehousing/](https://hgodalan.github.io/image-warehousing/) (Static UI - backend required for full functionality)

## Features

✅ **2D Image Support**: Upload photos, paintings, concept art

✅ **3D Object Support**: Upload 3D models (.glb, .gltf, .stl, .obj, .fbx) with 4 or 6 surface views, interactive 3D preview

✅ **AI-Powered Categorization**: Gemini automatically analyzes and categorizes uploads

✅ **Semantic Search**: Natural language queries ("dark cat image" finds both "black cat" and "cat at night")

✅ **Filesystem-Based**: No database needed - organized folders + markdown index

✅ **Async Processing**: Fast upload response with background AI analysis

✅ **RESTful API**: HTTP endpoints for all operations

✅ **Web UI**: Interactive frontend with drag-and-drop upload, warehouse browsing, and 3D model viewer

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

### 4. Access Web UI

Open your browser and navigate to:
- **Web Interface**: `http://localhost:8080/`
- **API Base**: `http://localhost:8080/api/v1/`

The web UI provides:
- 📤 **Upload Tab**: Drag-and-drop for 2D images
- 📦 **3D Upload Tab**: Upload 3D models with surface views (4 or 6 surfaces)
- 🏛️ **Warehouse Tab**: Browse all uploaded images with category filtering
- 🔍 **Search Tab**: Semantic search with natural language queries

## API Examples

### Upload 2D Image
```bash
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "image=@photo.jpg" \
  -F "title=Mountain Landscape" \
  -F "artist=John Doe" \
  -F "tags=[\"landscape\",\"outdoor\"]"
```

### Upload 3D Object
```bash
# 6-surface mode (front, back, left, right, top, bottom)
curl -X POST http://localhost:8080/api/v1/images/upload-3d \
  -F "model=@sculpture.glb" \
  -F "mode=6" \
  -F "front=@model_front.jpg" \
  -F "back=@model_back.jpg" \
  -F "left=@model_left.jpg" \
  -F "right=@model_right.jpg" \
  -F "top=@model_top.jpg" \
  -F "bottom=@model_bottom.jpg" \
  -F "title=Robot Character" \
  -F "artist=Jane Smith"

# 4-surface mode (front, back, left, right)
curl -X POST http://localhost:8080/api/v1/images/upload-3d \
  -F "model=@model.stl" \
  -F "mode=4" \
  -F "front=@front.jpg" \
  -F "back=@back.jpg" \
  -F "left=@left.jpg" \
  -F "right=@right.jpg" \
  -F "title=Abstract Sculpture" \
  -F "artist=Jane Smith"
```

**Supported 3D formats**: .glb, .gltf, .stl, .obj, .fbx, .blend, .dae

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
data/
├── categories/
│   ├── animals/cats/550e8400...jpg
│   ├── landscapes/mountains/660e8400...jpg
│   └── sculpture/anthropomorphic/770e8400.../
│       ├── model.stl                    # Original 3D model file
│       ├── front.jpg, back.jpg, ...     # Surface view images
│       └── front_thumb.jpg, ...         # Thumbnails
├── index.md                              # Searchable markdown index
└── temp/                                 # Temporary upload storage
frontend/                                 # Web UI files
├── index.html
├── style.css
└── app.js
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

## Technology Stack

**Backend:**
- Go 1.22+ with standard library HTTP server
- Google Gemini 3 (Flash/Pro) for AI vision analysis
- Filesystem-based storage (no database required)

**Frontend:**
- Vanilla JavaScript with ES modules
- Three.js for interactive 3D preview (.stl, .obj, .fbx)
- Google Model Viewer for GLB/GLTF preview
- Responsive CSS with modern gradients

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

## Deployment

### GitHub Pages (Frontend Only)
The web UI automatically deploys to GitHub Pages on push to main:
- **Live URL**: https://hgodalan.github.io/image-warehousing/
- **Note**: Backend required for full functionality (upload, search, AI analysis)
- **Workflow**: `.github/workflows/deploy-frontend.yml`

### Self-Hosting (Full Stack)
1. Build the binary: `make build`
2. Copy `bin/server` to your server
3. Set environment variables (GEMINI_API_KEY, etc.)
4. Run: `./server`
5. Serve on port 8080 or configure reverse proxy

## Documentation

- `plan.md` - Future pipeline plan (2D→3D generation with Gemini Pro + Tripo)
- `README.md` - This file
- `.env.example` - Configuration template
- **Live Demo**: https://hgodalan.github.io/image-warehousing/ (frontend UI preview)
