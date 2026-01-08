#!/bin/bash
# Manual Testing Helper Script

set -e

echo "🧪 Manual Testing Helper for Image Warehousing"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server is running
echo -e "${BLUE}Checking if server is running...${NC}"
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✓ Server is running${NC}"
else
    echo -e "${YELLOW}⚠ Server is not running. Start it with: make run${NC}"
    exit 1
fi

echo ""
echo "Available test commands:"
echo "------------------------"
echo ""

# 1. Upload test
echo -e "${BLUE}1. Upload a test image:${NC}"
echo "   curl -X POST http://localhost:8080/api/v1/images/upload \\"
echo "     -F \"image=@your_image.jpg\" \\"
echo "     -F \"title=Test Image\" \\"
echo "     -F \"artist=Your Name\" \\"
echo "     -F 'tags=[\"test\",\"sample\"]'"
echo ""

# 2. Search test
echo -e "${BLUE}2. Search for images:${NC}"
echo "   curl -X POST http://localhost:8080/api/v1/search \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d '{\"query\": \"your search query\", \"limit\": 10}'"
echo ""

# 3. Health check
echo -e "${BLUE}3. Check server health:${NC}"
echo "   curl http://localhost:8080/health"
echo ""

# 4. View index
echo -e "${BLUE}4. View the index file:${NC}"
echo "   cat data/index.md"
echo ""

# 5. Quick test if image provided
if [ -n "$1" ]; then
    echo -e "${GREEN}Running quick upload test with: $1${NC}"

    IMAGE_FILE="$1"
    TITLE="${2:-Test Upload}"
    ARTIST="${3:-Test User}"

    echo ""
    echo "Uploading $IMAGE_FILE..."

    RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/images/upload \
        -F "image=@$IMAGE_FILE" \
        -F "title=$TITLE" \
        -F "artist=$ARTIST" \
        -F 'tags=["test","manual"]')

    echo "Response:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

    echo ""
    echo -e "${GREEN}✓ Upload complete!${NC}"
    echo "Wait 30 seconds for AI processing, then try searching."
    echo ""
    echo "Search example:"
    echo "  curl -X POST http://localhost:8080/api/v1/search \\"
    echo "    -H \"Content-Type: application/json\" \\"
    echo "    -d '{\"query\": \"$TITLE\", \"limit\": 10}'"
else
    echo -e "${YELLOW}Usage: $0 [image_file] [title] [artist]${NC}"
    echo ""
    echo "Example:"
    echo "  $0 photo.jpg \"Beach Sunset\" \"John Doe\""
    echo ""
    echo "Or just run the automated test:"
    echo "  make test-e2e"
fi
