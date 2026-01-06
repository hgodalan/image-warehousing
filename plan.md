# Image Warehousing Tool - Implementation Plan

## Overview
Build an HTTP-based image warehousing system for artists with AI-powered analysis and natural language search using Google Gemini.

**Key Features:**
- **2D Image Support**: Single image uploads (photos, paintings, concept art, etc.)
- **3D Object Support**: Multi-view uploads (6 images per object - front, back, left, right, top, bottom)
- **AI-Powered Categorization**: Gemini automatically analyzes and categorizes all uploads
- **Semantic Search**: Natural language queries return relevant images/objects
- **Filesystem-Based**: No database required - everything stored in organized folders + markdown index

## Architecture Decisions

### Technology Stack
- **Language**: Go 1.22+
- **HTTP Framework**: gorilla/mux
- **Index Storage**: Markdown files (human-readable, version-controllable)
- **AI Provider**: Google Gemini 2.0 Flash (multimodal vision, supports multi-image analysis)
- **Image Processing**: disintegration/imaging (thumbnails)
- **Storage**: Local filesystem with category-based organization

**Note**: This plan is maintained in `plan.md` at the project root and will be updated as implementation progresses.

### Core Features
1. **Upload API**:
   - **2D Images**: Artists upload single image with optional manual tags
   - **3D Images**: Artists upload 6 images at once (6 surfaces/views of one 3D object)
2. **AI Analysis**: Gemini 2.0 Flash analyzes images and extracts features/categories automatically
   - For 3D: Analyzes all 6 views together to understand the complete 3D object
3. **Filesystem Organization**: Images organized by AI-detected categories in folder structure
   - 3D images stored as collections (6 images per object in dedicated folder)
4. **Markdown Index**: Central index.md file stores all image metadata and features
5. **Search API**: Natural language queries ("dark cat image") → LLM reviews markdown index → returns matching images

## Project Structure

```
image-warehousing/
├── cmd/server/main.go                    # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── upload.go                # 2D image upload endpoint
│   │   │   ├── upload_3d.go             # 3D object upload endpoint (6 views)
│   │   │   ├── search.go                # Search endpoint
│   │   │   └── health.go                # Health check
│   │   ├── middleware/
│   │   │   ├── logger.go                # Request logging
│   │   │   └── cors.go                  # CORS handling
│   │   └── router.go                    # Route definitions
│   ├── service/
│   │   ├── image_service.go             # Core business logic
│   │   ├── ai_service.go                # Gemini integration
│   │   ├── search_service.go            # Search logic
│   │   ├── storage_service.go           # File operations
│   │   └── index_service.go             # Markdown index management
│   ├── models/
│   │   ├── image.go                     # Image data structures (2D and 3D)
│   │   ├── search.go                    # Search request/response
│   │   └── ai_analysis.go               # AI analysis results
│   ├── config/
│   │   └── config.go                    # Configuration
│   └── utils/
│       ├── validator.go                 # Input validation
│       └── errors.go                    # Error types
├── pkg/gemini/
│   └── client.go                        # Gemini client wrapper
├── data/
│   ├── categories/                      # AI-organized by category
│   │   ├── animals/
│   │   │   ├── cats/
│   │   │   │   ├── {uuid}.jpg
│   │   │   │   └── {uuid}_thumb.jpg
│   │   │   └── dogs/
│   │   ├── landscapes/
│   │   │   ├── mountains/
│   │   │   └── ocean/
│   │   ├── portraits/
│   │   ├── 3d-renders/
│   │   └── uncategorized/               # Fallback category
│   └── index.md                         # Master index file
├── .env.example
├── .gitignore
├── go.mod
└── Makefile
```

## API Endpoints

**Base URL**: `http://localhost:8080/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/images/upload` | Upload 2D image with optional tags |
| POST | `/images/upload-3d` | Upload 3D object (6 images: front, back, left, right, top, bottom) |
| GET | `/images/{id}` | Get image metadata (2D or 3D) |
| GET | `/images/{id}/file` | Download original image (2D) |
| GET | `/images/{id}/view/{view}` | Get specific view of 3D object (front/back/left/right/top/bottom) |
| GET | `/images/{id}/thumbnail` | Get thumbnail |
| POST | `/search` | Search images by natural language |
| GET | `/images` | List all images (paginated) |
| GET | `/health` | Health check |

See full plan for detailed implementation steps, Gemini prompts, and deployment checklist.
