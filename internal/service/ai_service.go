package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yourcompany/image-warehousing/internal/models"
	"github.com/yourcompany/image-warehousing/pkg/gemini"
)

type AIService struct {
	geminiClient *gemini.Client
}

func NewAIService(apiKey, model string) (*AIService, error) {
	client, err := gemini.NewClient(apiKey, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &AIService{
		geminiClient: client,
	}, nil
}

func (s *AIService) Close() error {
	return s.geminiClient.Close()
}

// Analyze2DImage analyzes a single 2D image
func (s *AIService) Analyze2DImage(ctx context.Context, imagePath string) (*models.AIAnalysis, error) {
	resp, err := s.geminiClient.AnalyzeImage2D(ctx, imagePath)
	if err != nil {
		return nil, err
	}

	// Convert to our model
	analysis := &models.AIAnalysis{
		Type:            resp.Type,
		PrimaryCategory: resp.PrimaryCategory,
		Description:     resp.Description,
		Objects:         resp.Objects,
		Colors:          resp.Colors,
		SceneType:       resp.SceneType,
		Mood:            resp.Mood,
		Style:           resp.Style,
		Features:        s.parseFeatures(resp.Features),
	}

	// Store raw response
	rawJSON, _ := json.Marshal(resp)
	analysis.RawResponse = string(rawJSON)

	return analysis, nil
}

// Analyze3DObject analyzes a 3D object from 6 views
func (s *AIService) Analyze3DObject(ctx context.Context, viewPaths map[string]string) (*models.AIAnalysis, error) {
	resp, err := s.geminiClient.AnalyzeImage3D(ctx, viewPaths)
	if err != nil {
		return nil, err
	}

	// Convert to our model
	analysis := &models.AIAnalysis{
		Type:                  resp.Type,
		PrimaryCategory:       resp.PrimaryCategory,
		Description:           resp.Description,
		Objects:               resp.Objects,
		Colors:                resp.Colors,
		Style:                 resp.Style,
		Mood:                  resp.Mood,
		Lighting:              resp.Lighting,
		ThreeDCharacteristics: resp.ThreeDCharacteristics,
		Symmetry:              resp.Symmetry,
		Complexity:            resp.Complexity,
		Features:              s.parseFeatures(resp.Features),
	}

	// Store raw response
	rawJSON, _ := json.Marshal(resp)
	analysis.RawResponse = string(rawJSON)

	return analysis, nil
}

// SearchImages searches the index using Gemini
func (s *AIService) SearchImages(ctx context.Context, indexContent, query string) ([]models.SearchResult, error) {
	responseText, err := s.geminiClient.SearchImages(ctx, indexContent, query)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var results []struct {
		ImageID        string  `json:"image_id"`
		RelevanceScore float64 `json:"relevance_score"`
		Reason         string  `json:"reason"`
	}

	if err := json.Unmarshal([]byte(responseText), &results); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	// Convert to our model
	searchResults := make([]models.SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = models.SearchResult{
			ImageID:        r.ImageID,
			RelevanceScore: r.RelevanceScore,
			Reason:         r.Reason,
		}
	}

	return searchResults, nil
}

// parseFeatures converts string features to Feature objects with confidence
// Assumes features might be in format "tag (0.95)" or just "tag"
func (s *AIService) parseFeatures(features []string) []models.Feature {
	result := make([]models.Feature, 0, len(features))

	for _, f := range features {
		var name string
		var confidence float64 = 1.0

		// Try to parse confidence score if present
		if strings.Contains(f, "(") && strings.Contains(f, ")") {
			parts := strings.Split(f, "(")
			name = strings.TrimSpace(parts[0])
			confStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")
			fmt.Sscanf(confStr, "%f", &confidence)
		} else {
			name = strings.TrimSpace(f)
		}

		result = append(result, models.Feature{
			Name:       name,
			Confidence: confidence,
		})
	}

	return result
}

// GetCategoryPath constructs the category path from analysis
// Now returns only primary category for flat structure
func (s *AIService) GetCategoryPath(analysis *models.AIAnalysis) string {
	return s.normalizeCategoryName(analysis.PrimaryCategory)
}

// normalizeCategoryName converts category names to filesystem-safe paths
func (s *AIService) normalizeCategoryName(category string) string {
	// Convert to lowercase
	category = strings.ToLower(category)
	// Replace spaces with hyphens
	category = strings.ReplaceAll(category, " ", "-")
	// Remove special characters
	category = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, category)

	return category
}
