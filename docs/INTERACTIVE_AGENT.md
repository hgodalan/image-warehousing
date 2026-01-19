# Interactive Upload Agent - Your CLI Interface

Since we don't have a web UI yet, this interactive script acts as your **command-line interface** to the backend.

## Architecture

```
┌─────────────────┐         HTTP API         ┌─────────────────┐
│  Terminal 1     │◄──────────────────────────►  Terminal 2     │
│                 │                            │                 │
│  make run       │                            │  make           │
│  (Backend       │                            │  upload-artwork │
│   Server)       │                            │  (CLI Client/   │
│                 │                            │   Your UI)      │
│  Port 8080      │                            │                 │
└─────────────────┘                            └─────────────────┘
        │                                               │
        │                                               │
        ▼                                               ▼
   Backend Logic:                               User Interaction:
   • Gemini API calls                           • Upload images
   • Image processing                           • Search & query
   • Index generation                           • View stats
   • File organization                          • Interactive chat
```

## How to Use

### Terminal 1: Start Backend Server

```bash
# Make sure .env has GEMINI_API_KEY
make run
```

Wait for:
```
AI service initialized (model: gemini-3-flash)
Server listening on port 8080
Ready to accept requests!
```

### Terminal 2: Start Interactive Agent (Your UI)

```bash
make upload-artwork
```

Or directly:
```bash
go run scripts/interactive_upload.go
```

## What the Agent Does

### 1. **Scans Your Folders**
```
📁 Base path: C:\Users\You\Downloads\artwork_images

🔍 Scanning folders:
  • 參考圖 → 15 images (category: original-reference)
  • 產品模擬圖AI → 12 images (category: product-simulation-ai)
  • 設計稿AI → 8 images (category: design-draft-ai)

📊 Total: 35 images found
```

### 2. **Uploads All Images to Backend**
Makes API calls to `POST /api/v1/images/upload`:
```
📤 Uploading images...

📁 Processing: 參考圖
  [1] Uploading: concept_01.jpg... ✅ 550e8400
  [2] Uploading: concept_02.jpg... ✅ 660e8401
  ...

📁 Processing: 產品模擬圖AI
  [1] Uploading: product_sim_01.jpg... ✅ 770e8402
  ...
```

### 3. **Waits for Backend Processing**
```
⏳ Waiting for AI processing...
   (Background workers analyzing images with Gemini)
   Progress: 60/60 seconds
✅ Processing should be complete
```

### 4. **Enters Interactive Chat Mode**
```
🤖 Entering Interactive Knowledge Management Mode

💬 You: search AI generated product images

🔍 Searching for: "AI generated product images"
✅ Found 12 result(s):

  1. [770e8402] Score: 0.95
     Title: product_sim_01
     Category: product-simulation-ai
     Source: 產品模擬圖AI
     Reason: AI-generated product simulation image with high relevance

  2. [770e8403] Score: 0.92
     ...
```

## Interactive Commands

Once in chat mode, you can:

| Command | Description | Example |
|---------|-------------|---------|
| `search <query>` | Search for images | `search original reference images` |
| `stats` | Show upload statistics | `stats` |
| `list` | List all uploaded images | `list` |
| `help` | Show commands | `help` |
| `quit` | Exit agent | `quit` |

You can also just **type your query directly** without "search":
```
💬 You: show me all AI generated designs
💬 You: original artwork from colleague
💬 You: product simulation images
```

## Example Session

```
# Terminal 1
make run

# Terminal 2
make upload-artwork

# Interactive session:
💬 You: stats
📊 Statistics:
  Total uploaded: 35 images

  By category:
    • original-reference: 15 images
    • product-simulation-ai: 12 images
    • design-draft-ai: 8 images

💬 You: search AI generated product designs

🔍 Searching for: "AI generated product designs"
✅ Found 12 result(s):
  ...

💬 You: original reference images

🔍 Searching for: "original reference images"
✅ Found 15 result(s):
  ...

💬 You: quit
👋 Goodbye!
```

## Features

✅ **Batch Upload**: Uploads entire folders automatically
✅ **Smart Categorization**: Tags based on folder (original vs AI-generated)
✅ **Interactive Search**: Chat-like interface for querying
✅ **Knowledge Management**: Keep track of what you uploaded
✅ **Real-time Stats**: See upload progress and statistics

## Backend API Calls

The agent makes these HTTP calls to your backend:

1. **Health Check**: `GET /health`
2. **Upload Images**: `POST /api/v1/images/upload`
   - Multipart form data
   - Includes: image file, title, artist, tags
3. **Search**: `POST /api/v1/search`
   - JSON body: `{"query": "...", "limit": 10}`

## What Happens on the Backend

When you upload via the agent:

1. **Server receives upload** → saves to `data/temp/`
2. **Background worker picks up job**:
   - Calls Gemini 3 Flash for analysis
   - Generates thumbnail
   - Categorizes into folder
   - Adds to `data/index.md`
3. **Search becomes available** after ~30-60 seconds

## Troubleshooting

### Agent can't connect
- Make sure Terminal 1 (server) is running
- Check `http://localhost:8080/health` in browser

### Images not found
- Check the base path: `~/Downloads/artwork_images`
- Verify folder names match exactly (Chinese characters)

### Search returns no results
- Wait longer (processing takes 30-60 seconds)
- Check server logs for Gemini API errors
- Verify `GEMINI_API_KEY` is set in `.env`

### Upload fails
- Check file formats (jpg, png, gif supported)
- Check file sizes (default 50MB limit)
- Check server logs for errors

## Next Steps

After uploading your artwork:

1. **Check the index**: `cat data/index.md`
2. **Browse categories**: `dir data\categories`
3. **Try different searches**: Test semantic understanding
4. **View server logs**: See AI analysis in action

This interactive agent is your **temporary UI** until we build a web interface!
