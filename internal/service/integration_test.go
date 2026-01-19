package service

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourcompany/image-warehousing/internal/models"
)

// TestIndexingWorkflow tests the complete indexing workflow
func TestIndexingWorkflow(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	// Create services
	indexService := NewIndexService(tempDir)

	// Initialize index
	if err := indexService.InitializeIndex(); err != nil {
		t.Fatalf("Failed to initialize index: %v", err)
	}

	// Create test images with rich metadata
	now := time.Now()
	testImages := []*models.Image{
		{
			ID:            "img-001",
			Title:         "Sunset Beach",
			Artist:        "Alice",
			Type:          models.ImageType2D,
			UploadedAt:    now,
			ProcessedAt:   &now,
			Status:        "completed",
			FilePath:      "categories/landscapes/ocean/sunset_beach.jpg",
			ThumbnailPath: "categories/landscapes/ocean/sunset_beach_thumb.jpg",
			FileSize:      2048000,
			Width:         1920,
			Height:        1080,
			Category:      "landscapes",
			ManualTags:    []string{"beach", "sunset", "ocean", "nature"},
			AIAnalysis: &models.AIAnalysis{
				Description:     "A beautiful sunset over a calm ocean beach with golden sand",
				PrimaryCategory: "landscapes",
				Objects:         []string{"beach", "ocean", "sunset", "sky", "sand"},
				Colors:          []string{"orange", "blue", "gold", "purple"},
				SceneType:       "outdoor",
				Mood:            "calm",
				Style:           "photorealistic",
			},
		},
		{
			ID:            "img-002",
			Title:         "Mountain Peak",
			Artist:        "Bob",
			Type:          models.ImageType2D,
			UploadedAt:    now,
			ProcessedAt:   &now,
			Status:        "completed",
			FilePath:      "categories/landscapes/mountains/peak.jpg",
			ThumbnailPath: "categories/landscapes/mountains/peak_thumb.jpg",
			FileSize:      3072000,
			Width:         2560,
			Height:        1440,
			Category:      "landscapes",
			ManualTags:    []string{"mountain", "peak", "snow", "hiking"},
			AIAnalysis: &models.AIAnalysis{
				Description:     "A majestic snow-capped mountain peak against a clear blue sky",
				PrimaryCategory: "landscapes",
				Objects:         []string{"mountain", "snow", "sky", "rocks"},
				Colors:          []string{"white", "blue", "grey"},
				SceneType:       "outdoor",
				Mood:            "majestic",
				Style:           "photorealistic",
			},
		},
		{
			ID:            "img-003",
			Title:         "City Street at Night",
			Artist:        "Charlie",
			Type:          models.ImageType2D,
			UploadedAt:    now,
			ProcessedAt:   &now,
			Status:        "completed",
			FilePath:      "categories/urban/night_street.jpg",
			ThumbnailPath: "categories/urban/night_street_thumb.jpg",
			FileSize:      1536000,
			Width:         1280,
			Height:        720,
			Category:      "urban",
			ManualTags:    []string{"city", "night", "lights", "street"},
			AIAnalysis: &models.AIAnalysis{
				Description:     "A bustling city street illuminated by neon lights and street lamps",
				PrimaryCategory: "urban",
				Objects:         []string{"buildings", "cars", "lights", "street"},
				Colors:          []string{"blue", "red", "yellow", "black"},
				SceneType:       "outdoor",
				Mood:            "energetic",
				Style:           "photorealistic",
			},
		},
	}

	// Add images to index
	for _, img := range testImages {
		if err := indexService.AppendToIndex(img); err != nil {
			t.Fatalf("Failed to append image %s to index: %v", img.ID, err)
		}
	}

	// Read back the index
	indexContent, err := indexService.ReadIndex()
	if err != nil {
		t.Fatalf("Failed to read index: %v", err)
	}

	// Verify index contains all images
	for _, img := range testImages {
		if !contains(indexContent, "## Image: "+img.ID) {
			t.Errorf("Index does not contain image ID: %s", img.ID)
		}
		if !contains(indexContent, "**Title:** "+img.Title) {
			t.Errorf("Index does not contain image title: %s", img.Title)
		}
		if !contains(indexContent, "**Artist:** "+img.Artist) {
			t.Errorf("Index does not contain artist: %s", img.Artist)
		}
	}

	// Verify all tags are indexed
	expectedTags := []string{
		"beach", "sunset", "ocean", "nature",
		"mountain", "peak", "snow", "hiking",
		"city", "night", "lights", "street",
	}

	for _, tag := range expectedTags {
		if !contains(indexContent, tag) {
			t.Errorf("Index does not contain expected tag: %s", tag)
		}
	}

	// Verify AI analysis is included
	if !contains(indexContent, "**AI Analysis:**") {
		t.Error("Index should contain AI Analysis section")
	}

	if !contains(indexContent, "calm ocean beach") {
		t.Error("Index should contain AI description for beach image")
	}

	t.Logf("Successfully indexed %d images", len(testImages))
	t.Logf("Index file size: %d bytes", len(indexContent))
}

// TestSearchServiceIntegration tests search with populated index
func TestSearchServiceIntegration(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetOutput(os.Stderr)

	// Create services
	indexService := NewIndexService(tempDir)

	// Initialize and populate index
	if err := indexService.InitializeIndex(); err != nil {
		t.Fatalf("Failed to initialize index: %v", err)
	}

	// Add test data
	now := time.Now()
	testImage := &models.Image{
		ID:            "test-search-001",
		Title:         "Dark Forest",
		Artist:        "TestArtist",
		Type:          models.ImageType2D,
		UploadedAt:    now,
		ProcessedAt:   &now,
		Status:        "completed",
		FilePath:      "categories/nature/dark_forest.jpg",
		ThumbnailPath: "categories/nature/dark_forest_thumb.jpg",
		FileSize:      1024000,
		Width:         1600,
		Height:        900,
		Category:      "nature",
		ManualTags:    []string{"forest", "dark", "trees", "mysterious"},
		AIAnalysis: &models.AIAnalysis{
			Description:     "A dark, mysterious forest with tall trees",
			PrimaryCategory: "nature",
			Objects:         []string{"trees", "forest", "shadows"},
			Colors:          []string{"dark green", "brown", "black"},
			SceneType:       "outdoor",
			Mood:            "mysterious",
			Style:           "photorealistic",
		},
	}

	if err := indexService.AppendToIndex(testImage); err != nil {
		t.Fatalf("Failed to append test image: %v", err)
	}

	// Read index to verify
	indexContent, err := indexService.ReadIndex()
	if err != nil {
		t.Fatalf("Failed to read index: %v", err)
	}

	// Verify the index is searchable
	searchTerms := []string{"forest", "dark", "mysterious", "trees"}
	for _, term := range searchTerms {
		if !contains(indexContent, term) {
			t.Errorf("Index should contain searchable term: %s", term)
		}
	}

	t.Logf("Index is ready for search with %d searchable terms", len(searchTerms))
	t.Logf("Index content preview:\n%s", indexContent[:min(500, len(indexContent))])
}

// TestIndexPersistence tests that index survives service restarts
func TestIndexPersistence(t *testing.T) {
	tempDir := t.TempDir()

	// First service instance
	indexService1 := NewIndexService(tempDir)
	if err := indexService1.InitializeIndex(); err != nil {
		t.Fatalf("Failed to initialize index: %v", err)
	}

	// Add an image
	now := time.Now()
	testImage := &models.Image{
		ID:         "persist-001",
		Title:      "Test Persistence",
		Artist:     "Tester",
		Type:       models.ImageType2D,
		UploadedAt: now,
		Category:   "test",
		FilePath:   "test.jpg",
		FileSize:   1024,
		Width:      100,
		Height:     100,
	}

	if err := indexService1.AppendToIndex(testImage); err != nil {
		t.Fatalf("Failed to append image: %v", err)
	}

	// Create new service instance (simulates restart)
	indexService2 := NewIndexService(tempDir)

	// Index should already exist, InitializeIndex should not overwrite
	if err := indexService2.InitializeIndex(); err != nil {
		t.Fatalf("Failed to initialize index on restart: %v", err)
	}

	// Read and verify content persisted
	content, err := indexService2.ReadIndex()
	if err != nil {
		t.Fatalf("Failed to read index after restart: %v", err)
	}

	if !contains(content, "persist-001") {
		t.Error("Index did not persist image ID across service restart")
	}

	if !contains(content, "Test Persistence") {
		t.Error("Index did not persist image title across service restart")
	}

	t.Log("Index successfully persisted across service restart")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark index operations
func BenchmarkAppendToIndex(b *testing.B) {
	tempDir := b.TempDir()
	indexService := NewIndexService(tempDir)
	indexService.InitializeIndex()

	now := time.Now()
	testImage := &models.Image{
		ID:         "bench-001",
		Title:      "Benchmark Image",
		Artist:     "Bencher",
		Type:       models.ImageType2D,
		UploadedAt: now,
		Category:   "test",
		FilePath:   "test.jpg",
		FileSize:   1024,
		Width:      100,
		Height:     100,
		AIAnalysis: &models.AIAnalysis{
			Description:     "Test description",
			PrimaryCategory: "test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		indexService.AppendToIndex(testImage)
	}
}

func BenchmarkReadIndex(b *testing.B) {
	tempDir := b.TempDir()
	indexService := NewIndexService(tempDir)
	indexService.InitializeIndex()

	// Add some test data
	now := time.Now()
	for i := 0; i < 10; i++ {
		testImage := &models.Image{
			ID:         "bench-" + string(rune('a'+i)),
			Title:      "Image " + string(rune('A'+i)),
			Type:       models.ImageType2D,
			UploadedAt: now,
			Category:   "test",
			FilePath:   "test.jpg",
			FileSize:   1024,
		}
		indexService.AppendToIndex(testImage)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		indexService.ReadIndex()
	}
}
