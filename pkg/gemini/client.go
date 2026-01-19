package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	apiKey string
	client *genai.Client
	model  string
}

// Analysis2DResponse represents the JSON response for 2D image analysis
type Analysis2DResponse struct {
	Type            string   `json:"type"`
	PrimaryCategory string   `json:"primary_category"`
	Description     string   `json:"description"`
	Objects         []string `json:"objects"`
	Colors          []string `json:"colors"`
	SceneType       string   `json:"scene_type"`
	Mood            string   `json:"mood"`
	Style           string   `json:"style"`
	Features        []string `json:"features"`
}

// Analysis3DResponse represents the JSON response for 3D object analysis
type Analysis3DResponse struct {
	Type                  string   `json:"type"`
	PrimaryCategory       string   `json:"primary_category"`
	Description           string   `json:"description"`
	Objects               []string `json:"objects"`
	Colors                []string `json:"colors"`
	Style                 string   `json:"style"`
	Mood                  string   `json:"mood"`
	Lighting              string   `json:"lighting"`
	ThreeDCharacteristics string   `json:"three_d_characteristics"`
	Features              []string `json:"features"`
	Symmetry              string   `json:"symmetry"`
	Complexity            string   `json:"complexity"`
}

func NewClient(apiKey, model string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Default to gemini-3-flash-preview if no model specified
	if model == "" {
		model = "gemini-3-flash-preview"
	}

	return &Client{
		apiKey: apiKey,
		client: client,
		model:  model,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// detectImageFormat detects the image format from file extension
// Returns format string compatible with genai.ImageData (e.g., "jpeg", "png", "gif", "webp")
func detectImageFormat(imagePath string) string {
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".webp":
		return "webp"
	default:
		return "jpeg" // fallback to jpeg
	}
}

// cleanMarkdownJSON removes markdown code fences from JSON responses
// Gemini sometimes wraps JSON in ```json ... ``` blocks
func cleanMarkdownJSON(text string) string {
	text = strings.TrimSpace(text)

	// Remove ```json and ``` markers
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
	}
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
	}
	if strings.HasSuffix(text, "```") {
		text = strings.TrimSuffix(text, "```")
	}

	return strings.TrimSpace(text)
}

// AnalyzeImage2D analyzes a single 2D image
func (c *Client) AnalyzeImage2D(ctx context.Context, imagePath string) (*Analysis2DResponse, error) {
	// Read image file
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	// Detect image format from file extension
	format := detectImageFormat(imagePath)

	prompt := `Analyze this 2D image and provide categorization + detailed analysis.

Return as JSON with this structure:
{
  "type": "2D",
  "primary_category": "artwork|conceptual-art|surrealism|figurines|character-design|sculpture|performance-art|animals|landscapes|portraits|3d-renders|abstract|architecture|products|uncategorized",
  "description": "2-3 sentence detailed description",
  "objects": ["object1", "object2"],
  "colors": ["color1", "color2"],
  "scene_type": "indoor|outdoor|studio",
  "mood": "calm|dark|energetic|mysterious|whimsical|etc",
  "style": "photorealistic|cartoon|3D|painting|sketch|sculpture",
  "features": ["at least 10 descriptive tags"]
}

IMPORTANT: Return ONLY valid JSON, no other text.`

	model := c.client.GenerativeModel(c.model)
	model.SetTemperature(0.4)

	resp, err := model.GenerateContent(ctx,
		genai.Text(prompt),
		genai.ImageData(format, imgData),
	)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// Extract text from response
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean markdown code fences if present
	responseText = cleanMarkdownJSON(responseText)

	// Parse JSON response
	var analysis Analysis2DResponse
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w\nResponse: %s", err, responseText)
	}

	return &analysis, nil
}

// AnalyzeImage3D analyzes a 3D object from multiple surface views (4 or 6)
func (c *Client) AnalyzeImage3D(ctx context.Context, viewPaths map[string]string) (*Analysis3DResponse, error) {
	// Read all surface view images (dynamically handles 4 or 6 views)
	possibleViews := []string{"front", "back", "left", "right", "top", "bottom"}
	views := []string{}
	imageParts := []genai.Part{}

	// Only include views that are present
	for _, view := range possibleViews {
		if path, ok := viewPaths[view]; ok {
			views = append(views, view)

			imgData, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s view: %w", view, err)
			}

			// Detect format from file extension
			format := detectImageFormat(path)
			imageParts = append(imageParts, genai.ImageData(format, imgData))
		}
	}

	if len(views) == 0 {
		return nil, fmt.Errorf("no surface views provided")
	}

	// Build dynamic prompt based on views present
	viewsList := ""
	for i, view := range views {
		viewsList += fmt.Sprintf("%d. %s view\n", i+1, view)
	}

	viewsJoined := strings.Join(views, ", ")
	prompt := fmt.Sprintf(`Analyze this 3D object from %d different views (%s).
These %d images show the same 3D object from multiple angles.

Images provided:
%s
Analyze all %d views together to understand the complete 3D object.

Return as JSON with this structure:`, len(views), viewsJoined, len(views), viewsList, len(views)) + `
{
  "type": "3D",
  "primary_category": "sculpture|figurines|character-design|3d-renders|products|characters|environments|architecture|vehicles|artwork|uncategorized",
  "description": "2-3 sentence detailed description of the 3D object",
  "objects": ["primary objects identified"],
  "colors": ["dominant colors across all views"],
  "style": "photorealistic-3d|stylized|low-poly|high-poly|cartoon-3d|pbr|ceramic|sculpted",
  "mood": "futuristic|organic|mechanical|fantasy|realistic|whimsical|surreal|etc",
  "lighting": "studio|natural|dramatic|neutral",
  "three_d_characteristics": "describe topology, modeling style, material type",
  "features": ["at least 10 descriptive tags"],
  "symmetry": "symmetrical|asymmetrical",
  "complexity": "simple|moderate|complex|highly-detailed"
}

IMPORTANT: Return ONLY valid JSON, no other text.`

	model := c.client.GenerativeModel(c.model)
	model.SetTemperature(0.4)

	// Build parts array: prompt first, then all images
	parts := []genai.Part{genai.Text(prompt)}
	parts = append(parts, imageParts...)

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// Extract text from response
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean markdown code fences if present
	responseText = cleanMarkdownJSON(responseText)

	// Parse JSON response
	var analysis Analysis3DResponse
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w\nResponse: %s", err, responseText)
	}

	return &analysis, nil
}

// SearchImages uses Gemini to search through an index
func (c *Client) SearchImages(ctx context.Context, indexContent, query string) (string, error) {
	prompt := fmt.Sprintf(`Given the following image index and a user search query, find all relevant images.

Image Index:
%s

User Query: "%s"

Analyze the index and return a JSON array of matching image IDs ranked by relevance:
[
  {"image_id": "uuid", "relevance_score": 0.95, "reason": "why it matches"},
  ...
]

Consider:
- Semantic similarity (e.g., "dark cat" matches "black cat at night")
- AI features and confidence scores
- Manual tags
- Scene type and mood
- Object detection results

IMPORTANT: Return ONLY valid JSON array, no other text.`, indexContent, query)

	model := c.client.GenerativeModel(c.model)
	model.SetTemperature(0.2)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean markdown code fences if present
	responseText = cleanMarkdownJSON(responseText)

	return responseText, nil
}
