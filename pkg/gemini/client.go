package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"github.com/google/generative-ai-go/genai"
)

type Client struct {
	apiKey string
	client *genai.Client
}

// Analysis2DResponse represents the JSON response for 2D image analysis
type Analysis2DResponse struct {
	Type            string   `json:"type"`
	PrimaryCategory string   `json:"primary_category"`
	SubCategory     string   `json:"sub_category"`
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
	Type                   string   `json:"type"`
	PrimaryCategory        string   `json:"primary_category"`
	SubCategory            string   `json:"sub_category"`
	Description            string   `json:"description"`
	Objects                []string `json:"objects"`
	Colors                 []string `json:"colors"`
	Style                  string   `json:"style"`
	Mood                   string   `json:"mood"`
	Lighting               string   `json:"lighting"`
	ThreeDCharacteristics  string   `json:"three_d_characteristics"`
	Features               []string `json:"features"`
	Symmetry               string   `json:"symmetry"`
	Complexity             string   `json:"complexity"`
}

func NewClient(apiKey string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		apiKey: apiKey,
		client: client,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// AnalyzeImage2D analyzes a single 2D image
func (c *Client) AnalyzeImage2D(ctx context.Context, imagePath string) (*Analysis2DResponse, error) {
	// Read image file
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	prompt := `Analyze this 2D image and provide categorization + detailed analysis.

Return as JSON with this structure:
{
  "type": "2D",
  "primary_category": "animals|landscapes|portraits|3d-renders|abstract|architecture|uncategorized",
  "sub_category": "specific subcategory (e.g., cats, mountains, headshot)",
  "description": "2-3 sentence detailed description",
  "objects": ["object1", "object2"],
  "colors": ["color1", "color2"],
  "scene_type": "indoor|outdoor|studio",
  "mood": "calm|dark|energetic|mysterious|etc",
  "style": "photorealistic|cartoon|3D|painting",
  "features": ["at least 10 descriptive tags"]
}

IMPORTANT: Return ONLY valid JSON, no other text.`

	model := c.client.GenerativeModel("gemini-2.0-flash-exp")
	model.SetTemperature(0.4)

	resp, err := model.GenerateContent(ctx,
		genai.Text(prompt),
		genai.ImageData("image/jpeg", imgData),
	)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// Extract text from response
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Parse JSON response
	var analysis Analysis2DResponse
	if err := json.Unmarshal([]byte(responseText), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w\nResponse: %s", err, responseText)
	}

	return &analysis, nil
}

// AnalyzeImage3D analyzes a 3D object from 6 views
func (c *Client) AnalyzeImage3D(ctx context.Context, viewPaths map[string]string) (*Analysis3DResponse, error) {
	// Read all 6 view images
	views := []string{"front", "back", "left", "right", "top", "bottom"}
	imageParts := []genai.Part{}

	for _, view := range views {
		path, ok := viewPaths[view]
		if !ok {
			return nil, fmt.Errorf("missing view: %s", view)
		}

		imgData, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s view: %w", view, err)
		}

		imageParts = append(imageParts, genai.ImageData("image/jpeg", imgData))
	}

	prompt := `Analyze this 3D object from 6 different views (front, back, left, right, top, bottom).
These 6 images show the same 3D object from all angles.

Images provided:
1. Front view
2. Back view
3. Left view
4. Right view
5. Top view
6. Bottom view

Analyze all 6 views together to understand the complete 3D object.

Return as JSON with this structure:
{
  "type": "3D",
  "primary_category": "3d-renders|products|characters|environments|architecture|vehicles|uncategorized",
  "sub_category": "specific subcategory (e.g., characters, furniture, weapons)",
  "description": "2-3 sentence detailed description of the 3D object",
  "objects": ["primary objects identified"],
  "colors": ["dominant colors across all views"],
  "style": "photorealistic-3d|stylized|low-poly|high-poly|cartoon-3d|pbr",
  "mood": "futuristic|organic|mechanical|fantasy|realistic|etc",
  "lighting": "studio|natural|dramatic|neutral",
  "three_d_characteristics": "describe topology, modeling style, material type",
  "features": ["at least 10 descriptive tags"],
  "symmetry": "symmetrical|asymmetrical",
  "complexity": "simple|moderate|complex|highly-detailed"
}

IMPORTANT: Return ONLY valid JSON, no other text.`

	model := c.client.GenerativeModel("gemini-2.0-flash-exp")
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

	model := c.client.GenerativeModel("gemini-2.0-flash-exp")
	model.SetTemperature(0.2)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return responseText, nil
}
