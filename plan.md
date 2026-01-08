# AI-Powered 3D Model Generation Pipeline

> **⚠️ FUTURE PLAN - NOT YET IMPLEMENTED**
>
> This document outlines a future enhancement to transform the current image warehousing
> system into a full AI pipeline for 2D-to-3D model generation. This is planned for
> implementation after the core features (upload, indexing, search) are complete and tested.
>
> **Current Status**: Working on basic image upload, indexing, and search features
> **Future Work**: Multi-view generation (Gemini Pro) + 3D reconstruction (Tripo API)

## Overview
An automated pipeline that converts 2D concept art/sketches into searchable 3D models using AI. Users upload a single 2D image, and the system automatically generates multi-view product photos and creates a 3D model, making everything searchable for the community.

**Core Value Proposition:**
Upload a 2D sketch → Get a production-ready 3D model + searchable archive

## Use Case & Workflow

### User Journey
1. **User A** uploads a 2D sketch/concept art
2. System analyzes the image and generates professional product photography (4 views)
3. System sends generated views to Tripo for 3D reconstruction
4. System archives all assets (original, views, 3D model) with searchable tags
5. **Other users** can search by tags and discover/download 3D models

### Real-World Applications
- **Game Asset Creation**: Upload concept art → Get game-ready 3D models
- **E-commerce**: Product photos → 3D viewer for online stores
- **3D Printing**: Sketch → Printable 3D file
- **Asset Marketplaces**: Like Sketchfab, but with automated 3D generation

## Pipeline Architecture

### Stage-by-Stage Processing

```
┌─────────────────────────────────────────────────────────────────┐
│ Stage 1: Upload & Analysis (Gemini 3 Flash)                    │
│ - User uploads 2D image                                         │
│ - Analyze content, extract tags, determine object type          │
│ - Fast, cost-effective model                                    │
└──────────────────────┬──────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────────────┐
│ Stage 2: Multi-View Generation (Gemini 3 Pro + Imagen)         │
│ - Generate 4 photorealistic studio images:                     │
│   1. Front view                                                 │
│   2. 45-degree Profile                                          │
│   3. Back view                                                  │
│   4. Top view                                                   │
│ - Neutral grey background, flat lighting                       │
│ - Optimized for 3D reconstruction                              │
└──────────────────────┬──────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────────────┐
│ Stage 3: 3D Model Generation (Tripo API)                       │
│ - Send 4 generated views to Tripo                              │
│ - Wait for 3D reconstruction (GLB/OBJ format)                  │
│ - Download and store 3D model                                  │
└──────────────────────┬──────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────────────────────┐
│ Stage 4: Archival & Indexing                                   │
│ - Store all assets in organized structure                      │
│ - Create searchable tags from AI analysis                      │
│ - Build markdown index for search                              │
│ - Make available to other users                                │
└─────────────────────────────────────────────────────────────────┘
```

## Technology Stack

### AI Services
- **Gemini 3 Flash**: Fast analysis and tagging (cost-effective)
- **Gemini 3 Pro + Imagen**: High-quality multi-view image generation
- **Tripo AI**: Multi-view to 3D model reconstruction

### Backend
- **Language**: Go 1.22+
- **HTTP Framework**: gorilla/mux
- **Job Queue**: In-memory worker queue (can scale to Redis later)
- **Storage**: Filesystem-based with markdown index
- **Image Processing**: disintegration/imaging (thumbnails)

### Configuration
- Environment-based (12-factor app)
- Configurable AI models
- API key management for Gemini + Tripo

## Gemini Prompts

### Analysis Prompt (Gemini 3 Flash)
```
Analyze this 2D image and extract detailed metadata for cataloging.

Return as JSON:
{
  "object_type": "character|product|vehicle|environment|other",
  "description": "detailed 2-3 sentence description",
  "tags": ["tag1", "tag2", "tag3", ...],  // at least 10 tags
  "primary_category": "category name",
  "colors": ["color1", "color2"],
  "style": "realistic|cartoon|anime|sketch|technical",
  "complexity": "simple|moderate|complex",
  "suitability_for_3d": "high|medium|low"
}

IMPORTANT: Return ONLY valid JSON, no other text.
```

### Multi-View Generation Prompt (Gemini 3 Pro)
```
Act as a professional product photographer. Generate 4 photorealistic studio
images of this object from the following angles: Front, 45-degree Profile,
Back, and Top. Use a neutral grey background and flat lighting to highlight
the geometry. This is for 3D reconstruction.

Requirements:
- Consistent lighting across all views
- Same scale and framing for all angles
- Clean neutral grey background (#808080)
- No shadows or reflections
- Flat, even studio lighting
- Focus on geometric accuracy

Generate 4 separate images labeled: front, profile, back, top
```

## Data Models

### ProcessingJob
```go
type ProcessingJob struct {
    ID              string                 // Unique job ID
    UserID          string                 // Creator's ID
    Status          JobStatus              // Current status
    CurrentStage    string                 // "analyzing", "generating_views", "creating_3d", "indexing"

    // Assets
    OriginalImage   string                 // User's upload path
    GeneratedViews  map[string]string      // view -> path (front, profile, back, top)
    Model3DPath     string                 // Tripo GLB/OBJ file path
    ThumbnailPath   string                 // Preview thumbnail

    // Metadata
    Analysis        *AIAnalysis            // From Gemini Flash
    Tags            []string               // Searchable tags
    Category        string                 // Primary category

    // Timing
    CreatedAt       time.Time
    UpdatedAt       time.Time
    CompletedAt     *time.Time

    // Error handling
    Error           string                 // Error message if failed
    RetryCount      int                    // Number of retries
}

type JobStatus string
const (
    JobStatusPending        JobStatus = "pending"
    JobStatusAnalyzing      JobStatus = "analyzing"
    JobStatusGeneratingViews JobStatus = "generating_views"
    JobStatusCreating3D     JobStatus = "creating_3d"
    JobStatusIndexing       JobStatus = "indexing"
    JobStatusCompleted      JobStatus = "completed"
    JobStatusFailed         JobStatus = "failed"
)
```

### AIAnalysis
```go
type AIAnalysis struct {
    ObjectType          string   // character, product, vehicle, etc.
    Description         string
    Tags                []string
    PrimaryCategory     string
    Colors              []string
    Style               string
    Complexity          string
    SuitabilityFor3D    string   // high, medium, low
    RawResponse         string   // Full Gemini response
}
```

## Storage Structure

```
data/
├── jobs/
│   └── {job-id}/
│       ├── original.jpg              # User's original upload
│       ├── views/
│       │   ├── front.jpg             # Gemini-generated
│       │   ├── profile.jpg           # Gemini-generated (45°)
│       │   ├── back.jpg              # Gemini-generated
│       │   └── top.jpg               # Gemini-generated
│       ├── model.glb                 # Tripo-generated 3D model
│       ├── thumbnail.jpg             # Preview thumbnail
│       └── metadata.json             # Job metadata
├── categories/                       # Organized by tags
│   ├── characters/
│   ├── vehicles/
│   ├── products/
│   └── environments/
└── index.md                          # Searchable index
```

## API Endpoints

**Base URL**: `http://localhost:8080/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| **Job Management** |
| POST | `/jobs/create` | Upload 2D image, start pipeline |
| GET | `/jobs/{id}` | Get job status and progress |
| GET | `/jobs/{id}/download` | Download completed 3D model |
| DELETE | `/jobs/{id}` | Cancel/delete job |
| **Search & Discovery** |
| POST | `/search` | Search by tags/description |
| GET | `/models` | List all completed models (paginated) |
| GET | `/models/{id}` | Get model metadata + assets |
| GET | `/tags` | List popular tags |
| **Assets** |
| GET | `/jobs/{id}/original` | Get original 2D image |
| GET | `/jobs/{id}/views/{view}` | Get generated view (front/profile/back/top) |
| GET | `/jobs/{id}/thumbnail` | Get thumbnail |
| **System** |
| GET | `/health` | Health check |
| GET | `/stats` | System statistics |

## API Request/Response Examples

### Create Job
```bash
POST /api/v1/jobs/create
Content-Type: multipart/form-data

{
  "image": <file>,
  "user_id": "user123",
  "title": "Sci-fi Character Concept",
  "tags": ["scifi", "character", "armor"]  // optional manual tags
}

Response:
{
  "job_id": "job-abc123",
  "status": "pending",
  "estimated_time": "3-5 minutes"
}
```

### Get Job Status
```bash
GET /api/v1/jobs/job-abc123

Response:
{
  "id": "job-abc123",
  "status": "generating_views",
  "current_stage": "generating_views",
  "progress": 50,
  "created_at": "2025-01-06T10:30:00Z",
  "estimated_completion": "2025-01-06T10:35:00Z",
  "assets": {
    "original": "/api/v1/jobs/job-abc123/original",
    "thumbnail": "/api/v1/jobs/job-abc123/thumbnail",
    "views": {
      "front": "/api/v1/jobs/job-abc123/views/front",
      "profile": "/api/v1/jobs/job-abc123/views/profile"
    }
  }
}
```

### Search Models
```bash
POST /api/v1/search
{
  "query": "sci-fi character with armor",
  "limit": 10
}

Response:
{
  "results": [
    {
      "job_id": "job-abc123",
      "title": "Sci-fi Character Concept",
      "tags": ["scifi", "character", "armor", "robot"],
      "thumbnail": "/api/v1/jobs/job-abc123/thumbnail",
      "model_url": "/api/v1/jobs/job-abc123/download",
      "created_at": "2025-01-06T10:30:00Z",
      "user_id": "user123"
    }
  ],
  "total": 1
}
```

## Service Architecture

### New Services Required

**1. JobService** (`internal/service/job_service.go`)
- Create and manage processing jobs
- Track job status through pipeline stages
- Handle retries and error recovery
- Coordinate between AI, Storage, and Index services

**2. ViewGenerationService** (`internal/service/view_generation_service.go`)
- Use Gemini 3 Pro + Imagen to generate multi-view images
- Apply the product photography prompt
- Validate generated images
- Store views to filesystem

**3. TripoService** (`pkg/tripo/client.go`)
- Send multi-view images to Tripo API
- Poll for 3D model completion
- Download GLB/OBJ files
- Handle Tripo-specific errors

**4. Enhanced AIService** (`internal/service/ai_service.go`)
- Support multiple Gemini models (Flash + Pro)
- Separate analysis vs generation methods
- Image generation via Imagen integration

### Updated Project Structure
```
image-warehousing/
├── cmd/server/main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── job_handler.go           # NEW: Job creation & status
│   │   │   ├── search_handler.go        # UPDATED: Search by tags
│   │   │   ├── model_handler.go         # NEW: Model download
│   │   │   └── health.go
│   │   ├── middleware/
│   │   │   ├── logger.go
│   │   │   ├── cors.go
│   │   │   └── auth.go                  # NEW: Optional user auth
│   │   └── router.go
│   ├── service/
│   │   ├── job_service.go               # NEW: Job orchestration
│   │   ├── view_generation_service.go   # NEW: Gemini image gen
│   │   ├── ai_service.go                # UPDATED: Multi-model support
│   │   ├── search_service.go
│   │   ├── storage_service.go
│   │   └── index_service.go
│   ├── models/
│   │   ├── job.go                       # NEW: ProcessingJob model
│   │   ├── image.go
│   │   ├── search.go
│   │   └── ai_analysis.go
│   ├── worker/
│   │   └── pipeline_worker.go           # NEW: Background job processor
│   └── config/
│       └── config.go                    # UPDATED: Tripo API key
├── pkg/
│   ├── gemini/
│   │   └── client.go                    # UPDATED: Image generation
│   └── tripo/
│       └── client.go                    # NEW: Tripo API client
└── data/
    └── jobs/                            # NEW: Job-based storage
```

## Environment Configuration

### .env.example
```bash
# Server Configuration
SERVER_PORT=8080

# Gemini AI Configuration
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_ANALYSIS_MODEL=gemini-3-flash      # Fast analysis
GEMINI_GENERATION_MODEL=gemini-3-pro      # High-quality image generation

# Tripo AI Configuration
TRIPO_API_KEY=your_tripo_api_key_here
TRIPO_API_URL=https://api.tripo3d.ai/v1

# Storage Configuration
DATA_DIR=./data
MAX_UPLOAD_SIZE=52428800                  # 50MB
JOB_RETENTION_DAYS=30                     # Auto-cleanup old jobs

# Worker Configuration
WORKER_COUNT=3                             # Concurrent job processing
MAX_RETRIES=3                              # Retry failed stages

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000
```

## Pipeline Implementation Phases

### Phase 1: Foundation (1-2 weeks)
- [ ] Job model and storage structure
- [ ] Job creation API endpoint
- [ ] Job status tracking API
- [ ] Background worker system
- [ ] Update Gemini integration for dual-model support

### Phase 2: View Generation (1 week)
- [ ] Gemini 3 Pro + Imagen integration
- [ ] Multi-view image generation service
- [ ] Product photography prompt implementation
- [ ] View validation and storage

### Phase 3: Tripo Integration (1 week)
- [ ] Tripo API client
- [ ] Multi-view to 3D pipeline stage
- [ ] 3D model download and storage
- [ ] Error handling and retries

### Phase 4: Search & Discovery (1 week)
- [ ] Enhanced tag-based search
- [ ] Model listing and pagination
- [ ] Thumbnail generation
- [ ] User-facing model viewer API

### Phase 5: Polish & Production (1 week)
- [ ] Error recovery and retry logic
- [ ] Job cleanup and retention policies
- [ ] Monitoring and logging
- [ ] Performance optimization
- [ ] Documentation and deployment guide

## Cost Estimation (Per Job)

| Service | Operation | Cost (Approx) |
|---------|-----------|---------------|
| Gemini 3 Flash | Image analysis | $0.001 |
| Gemini 3 Pro | 4 image generations | $0.10 - $0.20 |
| Tripo AI | 3D model generation | $0.50 - $2.00 |
| **Total per job** | | **~$0.60 - $2.20** |

**Optimization tips:**
- Cache analysis results
- Batch jobs when possible
- Use Flash for preview generations
- Implement job queuing to manage API rate limits

## Success Metrics

- **Processing Success Rate**: >95% jobs complete successfully
- **Average Processing Time**: 3-5 minutes per job
- **Search Relevance**: Users find models via tags >80% of the time
- **3D Model Quality**: Tripo reconstruction success rate >90%
- **User Engagement**: Models reused/downloaded by other users

## Security Considerations

- **API Key Management**: Never commit keys, use environment variables
- **File Upload Validation**: Check file types, sizes, prevent malicious uploads
- **Rate Limiting**: Prevent abuse of expensive AI operations
- **User Isolation**: Jobs should be owned by users (future: auth)
- **Content Moderation**: Optional AI-based NSFW detection

## Future Enhancements

- **User Authentication**: OAuth, user accounts, portfolios
- **WebSocket Updates**: Real-time job progress updates
- **Batch Processing**: Upload multiple sketches at once
- **Custom Tripo Parameters**: User control over 3D quality/resolution
- **Model Versioning**: Regenerate with different settings
- **Community Features**: Likes, comments, collections
- **Marketplace**: Buy/sell generated 3D models
- **API for Integrations**: Allow external tools to use the pipeline

---

**Status**: Updated 2025-01-06
**Next Steps**: Implement Phase 1 (Job foundation and worker system)
