# Testing Search Functionality

## Current Status

✅ Search endpoint implemented: `POST /api/v1/search`
✅ Search service uses Gemini for semantic search
⚠️ **Requires Gemini API key to work**

## How Search Works

1. Client sends query to `/api/v1/search`
2. Server reads the markdown index (`data/index.md`)
3. Server sends index + query to Gemini
4. Gemini performs semantic search and returns ranked results
5. Server returns results to client

## Testing Search (Manual)

### Prerequisites
1. Gemini API key configured in `.env`
2. At least one image indexed in `data/index.md`

### Step 1: Setup Environment
```bash
# Copy example config
cp .env.example .env

# Edit .env and add your real API key
# GEMINI_API_KEY=your_actual_key_here
```

### Step 2: Start Server
```bash
make run
```

### Step 3: Upload Test Image (to create index data)
```bash
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "image=@test.jpg" \
  -F "title=Dark Forest Scene" \
  -F "artist=Test Artist" \
  -F "tags=[\"forest\",\"dark\",\"nature\"]"
```

Wait for background processing to complete (check logs).

### Step 4: Test Search
```bash
# Search for "forest"
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "dark forest images",
    "limit": 10
  }'
```

### Expected Response
```json
{
  "results": [
    {
      "image_id": "550e8400-e29b-41d4-a716-446655440000",
      "relevance_score": 0.95,
      "reason": "Matches query for dark forest with high confidence",
      "image": null
    }
  ],
  "total": 1,
  "query": "dark forest images"
}
```

## Why Search Isn't Tested Automatically

The search feature requires:
1. **Real Gemini API calls** - Can't mock in unit tests
2. **API key** - Not available in CI/CD
3. **Network connectivity** - May not be available in test environment
4. **Cost** - Each search costs ~$0.001-0.01

## Search Prompt (Sent to Gemini)

```
Search this image index for images matching the query.

Index content:
{markdown index content}

Query: "{user query}"

Return JSON array of matching images ranked by relevance:
[
  {
    "image_id": "id from index",
    "relevance_score": 0.0-1.0,
    "reason": "why it matches"
  }
]

Return empty array [] if no matches found.
```

## Alternative: Mock Search for Testing

If you want to test without a real API key, you can:

1. **Create a mock Gemini client** for testing
2. **Use simple keyword matching** instead of AI (fallback mode)
3. **Pre-record API responses** for specific queries

## Troubleshooting

### "GEMINI_API_KEY is required"
- Add your API key to `.env` file
- Make sure `.env` is in project root

### "failed to read index"
- Make sure `data/index.md` exists
- Upload at least one image first
- Check that initialization ran: `make init-data`

### Empty results
- Index might be empty (no uploaded images)
- Query might not match any indexed content
- Check server logs for Gemini errors

### "Search failed: context deadline exceeded"
- Gemini API request timed out
- Check internet connectivity
- Try again (might be temporary)

## Future Improvements

- [ ] Add fallback keyword search (no API key needed)
- [ ] Cache search results for common queries
- [ ] Add search filters (by category, date, artist)
- [ ] Support multi-language queries
- [ ] Add search suggestions/autocomplete
