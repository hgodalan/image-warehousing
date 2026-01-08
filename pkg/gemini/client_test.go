package gemini

import (
	"context"
	"os"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// TestGeminiModelConnection tests the connection to Gemini API with the configured model
func TestGeminiModelConnection(t *testing.T) {
	// Load .env file
	_ = godotenv.Load("../../.env")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping test")
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-3-flash-preview" // default
	}

	ctx := context.Background()

	// Create client using the sample code pattern
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer client.Close()

	// Get the model
	model := client.GenerativeModel(modelName)
	if model == nil {
		t.Fatalf("Failed to get model: %s", modelName)
	}

	t.Logf("Successfully created client with model: %s", modelName)

	// Send a simple text request to verify the model works
	resp, err := model.GenerateContent(ctx, genai.Text("Say hello in one word"))
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		t.Fatal("Empty response from Gemini")
	}

	// Print the response
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				t.Logf("Response: %v", part)
			}
		}
	}

	t.Logf("✅ Model %s is working correctly!", modelName)
}

// TestGeminiImageAnalysis tests image analysis with Gemini
func TestGeminiImageAnalysis(t *testing.T) {
	// Load .env file
	_ = godotenv.Load("../../.env")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping test")
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-3-flash-preview" // default
	}

	// Check if test image exists (try .png first, then .jpg)
	testImagePath := "../../test_images/sample.png"
	imageData, err := os.ReadFile(testImagePath)
	imageFormat := "png"

	if err != nil {
		// Try .jpg if .png doesn't exist
		testImagePath = "../../test_images/sample.jpg"
		imageData, err = os.ReadFile(testImagePath)
		imageFormat = "jpeg"

		if err != nil {
			t.Skipf("Test image not found at ../../test_images/sample.png or sample.jpg, skipping test")
		}
	}

	ctx := context.Background()

	// Create client using the sample code pattern
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer client.Close()

	// Get the model
	model := client.GenerativeModel(modelName)

	t.Logf("Testing image analysis with model: %s", modelName)

	// Send the image for analysis (using the sample code pattern)
	resp, err := model.GenerateContent(ctx,
		genai.ImageData(imageFormat, imageData),
		genai.Text("Analyze this image and describe what you see in one sentence."))
	if err != nil {
		t.Fatalf("Failed to analyze image: %v", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		t.Fatal("Empty response from Gemini")
	}

	// Print the analysis
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				t.Logf("Analysis: %v", part)
			}
		}
	}

	t.Logf("✅ Image analysis with model %s successful!", modelName)
}
