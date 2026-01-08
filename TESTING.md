# Testing Guide

This guide covers both **manual testing** (download images and test yourself) and **automated testing** (run scripts).

## Prerequisites

1. **Server running**: `make run` in one terminal
2. **Gemini API key** configured in `.env`
3. **Test images** (either download manually or use the automation)

## Option 1: Manual Testing (Recommended First Time)

### Step 1: Get Test Images

Download some images to test with. Here are free sources:

- **Unsplash**: https://unsplash.com (free high-quality photos)
- **Pexels**: https://www.pexels.com (free stock photos)
- **Pixabay**: https://pixabay.com (free images)

Download a few different types:
- 1-2 landscape/nature photos (beach, mountain, forest)
- 1-2 urban/city photos
- 1-2 other (animals, people, objects)

Save them in a folder like `D:\test_images\`

### Step 2: Start the Server

```bash
# Terminal 1
make run
```

Wait for the startup message:
```
AI service initialized (model: gemini-3-flash)
Server listening on port 8080
Ready to accept requests!
```

### Step 3: Upload Your First Image

**Windows (PowerShell/CMD):**
```batch
curl -X POST http://localhost:8080/api/v1/images/upload ^
  -F "image=@D:\test_images\beach.jpg" ^
  -F "title=Beautiful Beach Sunset" ^
  -F "artist=Test User" ^
  -F "tags=[\"beach\",\"sunset\",\"ocean\"]"
```

**Expected Response:**
```json
{
  "image_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "message": "Image uploaded successfully. Processing in background."
}
```

**Check the server logs** - you should see:
```
Generating thumbnail for 550e8400...
Analyzing 2D image 550e8400 with Gemini
Image 550e8400 categorized as: landscapes/ocean
Adding image 550e8400 to index
```

### Step 4: Upload More Images

Upload 2-3 more images with different tags:

```batch
REM Mountain image
curl -X POST http://localhost:8080/api/v1/images/upload ^
  -F "image=@D:\test_images\mountain.jpg" ^
  -F "title=Snow Mountain Peak" ^
  -F "artist=Test User" ^
  -F "tags=[\"mountain\",\"snow\",\"nature\"]"

REM City image
curl -X POST http://localhost:8080/api/v1/images/upload ^
  -F "image=@D:\test_images\city.jpg" ^
  -F "title=City Lights at Night" ^
  -F "artist=Test User" ^
  -F "tags=[\"city\",\"night\",\"urban\"]"
```

### Step 5: Wait for Processing

⏳ **Wait about 30-60 seconds** for the background workers to:
- Analyze images with Gemini
- Generate thumbnails
- Move files to category folders
- Update the index

Watch the server logs to see when processing completes.

### Step 6: Check the Index

View the generated markdown index:

```batch
type data\index.md
```

You should see entries like:
```markdown
## Image: 550e8400-e29b-41d4-a716-446655440000

**Title:** Beautiful Beach Sunset
**Artist:** Test User
**Type:** 2D
**Category:** landscapes/ocean
**File Path:** categories/landscapes/ocean/550e8400-e29b-41d4-a716-446655440000.jpg
**Thumbnail:** categories/landscapes/ocean/550e8400-e29b-41d4-a716-446655440000_thumb.jpg
**Dimensions:** 1920x1080
**File Size:** 2.5 MB

**Manual Tags:** beach, sunset, ocean

**AI Analysis:**
- **Description:** A stunning beach sunset with golden light reflecting on calm ocean waves
- **Primary Category:** landscapes
- **Sub-category:** ocean
- **Objects Detected:** beach, ocean, sunset, sky, sand
- **Dominant Colors:** orange, blue, gold
- **Scene Type:** outdoor
- **Mood:** calm
- **Style:** photorealistic
```

### Step 7: Test Search

Now search for your images:

```batch
REM Search for beach
curl -X POST http://localhost:8080/api/v1/search ^
  -H "Content-Type: application/json" ^
  -d "{\"query\": \"beach sunset ocean\", \"limit\": 10}"

REM Search for mountains
curl -X POST http://localhost:8080/api/v1/search ^
  -H "Content-Type: application/json" ^
  -d "{\"query\": \"mountain landscape\", \"limit\": 10}"

REM Semantic search (test AI understanding)
curl -X POST http://localhost:8080/api/v1/search ^
  -H "Content-Type: application/json" ^
  -d "{\"query\": \"peaceful nature scenes\", \"limit\": 10}"
```

**Expected Search Response:**
```json
{
  "results": [
    {
      "image_id": "550e8400-e29b-41d4-a716-446655440000",
      "relevance_score": 0.95,
      "reason": "Matches beach, sunset, and ocean tags with high confidence"
    }
  ],
  "total": 1,
  "query": "beach sunset ocean"
}
```

### Step 8: Verify File Organization

Check that files were organized correctly:

```batch
dir /s data\categories
```

You should see structure like:
```
data/categories/
├── landscapes/
│   ├── ocean/
│   │   ├── 550e8400...jpg
│   │   └── 550e8400..._thumb.jpg
│   └── mountains/
│       ├── 660e8400...jpg
│       └── 660e8400..._thumb.jpg
└── urban/
    └── night/
        ├── 770e8400...jpg
        └── 770e8400..._thumb.jpg
```

## Option 2: Automated Testing

### Quick Automated Test

1. **Place images in `test_images/` folder**
2. **Run the script:**

```bash
# Windows
go run scripts/test_e2e.go

# Or use make
make test-e2e
```

The script will:
- ✅ Check server health
- ✅ Upload all images from `test_images/`
- ✅ Wait for processing
- ✅ Test multiple search queries
- ✅ Display results

### Using the Helper Script

**Windows:**
```batch
scripts\test_manual.bat D:\test_images\beach.jpg "Beach Sunset" "John Doe"
```

**Linux/Mac:**
```bash
bash scripts/test_manual.sh ~/test_images/beach.jpg "Beach Sunset" "John Doe"
```

## Troubleshooting

### "Connection refused"
- Server isn't running
- Run `make run` first

### "GEMINI_API_KEY is required"
- Add your API key to `.env`
- Get one from: https://aistudio.google.com/app/apikey

### "Upload successful but no results in search"
- Wait longer for processing (30-60 seconds)
- Check server logs for errors
- Make sure Gemini API key is valid

### "Empty search results"
- Index might be empty (no uploaded images processed yet)
- Check `data/index.md` exists and has content
- Try uploading images again

### Processing never completes
- Check server logs for Gemini API errors
- Verify API key is correct
- Check internet connectivity
- Try simpler images (smaller size, common formats)

## What to Test

### Basic Functionality
- [x] Upload 2D images
- [x] Thumbnails generated
- [x] Files moved to category folders
- [x] Index updated with metadata
- [x] Search returns relevant results

### Edge Cases
- [x] Upload very large images (test MAX_UPLOAD_SIZE)
- [x] Upload unsupported formats (should fail gracefully)
- [x] Search with no results
- [x] Search with empty index
- [x] Multiple uploads simultaneously

### Search Quality
- [x] Exact tag matching ("beach" finds beach images)
- [x] Semantic search ("peaceful nature" finds calm landscapes)
- [x] Multi-word queries ("dark city night")
- [x] Typo tolerance (Gemini may handle minor typos)

## Next Steps

After successful testing:

1. **Explore the index**: `cat data/index.md`
2. **Check categories**: Browse `data/categories/`
3. **Try more search queries**: Test semantic understanding
4. **Upload different image types**: Test AI categorization
5. **Read the logs**: Understand the processing pipeline

## Performance Benchmarks

Expected timing (depends on image size and Gemini API latency):
- Upload: < 1 second
- Thumbnail generation: < 5 seconds
- Gemini analysis: 5-15 seconds
- Total processing: 10-30 seconds per image
- Search: 2-5 seconds

## Cost Estimates

Per image uploaded (with Gemini 3 Flash):
- Analysis: ~$0.001
- Search (per query): ~$0.001-0.01

For 100 images + 50 searches: ~$0.60 total
