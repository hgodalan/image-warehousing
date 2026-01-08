# Image Warehouse - Demo UI

A simple single-page web interface for the Image Warehouse system.

## Features

### 📤 Upload Tab
- Drag & drop or browse to select images
- Batch upload multiple images at once
- Add title, artist, and tags
- Real-time upload progress tracking
- Automatic AI processing in background

### 🏛️ Warehouse Tab
- View all uploaded images in a grid
- Filter by category
- Click any image to view details
- Shows thumbnails, titles, and tags
- Displays total image count

### 🔍 Search Tab
- Semantic AI-powered search
- Search by descriptions, features, mood, style, etc.
- Example queries provided
- Shows relevance scores and match reasons
- Click results to view full details

## How to Use

1. **Start the server:**
   ```bash
   cd bin
   ./server
   ```

2. **Open in browser:**
   ```
   http://localhost:8080
   ```

3. **Upload images:**
   - Go to Upload tab
   - Drag & drop images or click to browse
   - Optionally add metadata
   - Click "Upload All"

4. **Browse warehouse:**
   - Go to Warehouse tab
   - View all uploaded images
   - Filter by category
   - Click images for details

5. **Search:**
   - Go to Search tab
   - Enter search query (e.g., "anthropomorphic rabbit art")
   - View results with relevance scores

## API Endpoints Used

- `POST /api/v1/images/upload` - Upload images
- `GET /api/v1/images` - List all images
- `GET /api/v1/images/:id` - Get image details
- `POST /api/v1/search` - Search images

## File Structure

```
frontend/
├── index.html      # Main HTML page
├── style.css       # Styling
├── app.js          # JavaScript logic
└── README.md       # This file
```

## Notes

- Images are processed in the background by AI workers
- Search requires images to be fully processed (30-60 seconds after upload)
- Thumbnails are automatically generated
- All data stored in `data/` directory
