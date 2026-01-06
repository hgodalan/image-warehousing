package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourcompany/image-warehousing/internal/models"
)

func TestNewIndexService(t *testing.T) {
	dataDir := "/test/data"
	svc := NewIndexService(dataDir)

	expectedPath := filepath.Join(dataDir, "index.md")
	if svc.indexPath != expectedPath {
		t.Errorf("expected indexPath %s, got %s", expectedPath, svc.indexPath)
	}

	if svc.lock == nil {
		t.Error("lock should not be nil")
	}
}

func TestInitializeIndex(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewIndexService(tempDir)

	err := svc.InitializeIndex()
	if err != nil {
		t.Fatalf("InitializeIndex failed: %v", err)
	}

	// Check that index file was created
	content, err := os.ReadFile(svc.indexPath)
	if err != nil {
		t.Fatalf("failed to read index file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Image Warehouse Index") {
		t.Error("index file does not contain expected header")
	}

	if !strings.Contains(contentStr, "Last Updated:") {
		t.Error("index file does not contain 'Last Updated' field")
	}
}

func TestInitializeIndex_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewIndexService(tempDir)

	// Create index first time
	err := svc.InitializeIndex()
	if err != nil {
		t.Fatalf("InitializeIndex failed: %v", err)
	}

	// Write custom content
	customContent := "# Custom Content"
	err = os.WriteFile(svc.indexPath, []byte(customContent), 0644)
	if err != nil {
		t.Fatalf("failed to write custom content: %v", err)
	}

	// Initialize again - should not overwrite
	err = svc.InitializeIndex()
	if err != nil {
		t.Fatalf("InitializeIndex failed on second call: %v", err)
	}

	// Verify content is unchanged
	content, err := os.ReadFile(svc.indexPath)
	if err != nil {
		t.Fatalf("failed to read index file: %v", err)
	}

	if string(content) != customContent {
		t.Error("InitializeIndex overwrote existing content")
	}
}

func TestAppendToIndex_2DImage(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewIndexService(tempDir)

	// Initialize index
	err := svc.InitializeIndex()
	if err != nil {
		t.Fatalf("InitializeIndex failed: %v", err)
	}

	// Create test image
	now := time.Now()
	image := &models.Image{
		ID:            "test-id-123",
		Title:         "Test Image",
		Artist:        "Test Artist",
		Type:          models.ImageType2D,
		UploadedAt:    now,
		ProcessedAt:   &now,
		Status:        "completed",
		FilePath:      "categories/test/image.jpg",
		ThumbnailPath: "categories/test/image_thumb.jpg",
		FileSize:      1024000,
		Width:         1920,
		Height:        1080,
		Category:      "test/category",
		ManualTags:    []string{"tag1", "tag2"},
		AIAnalysis: &models.AIAnalysis{
			Description:     "A test image",
			PrimaryCategory: "test",
			SubCategory:     "category",
			Objects:         []string{"object1", "object2"},
			Colors:          []string{"red", "blue"},
			SceneType:       "indoor",
			Mood:            "calm",
			Style:           "realistic",
			Lighting:        "natural",
		},
	}

	err = svc.AppendToIndex(image)
	if err != nil {
		t.Fatalf("AppendToIndex failed: %v", err)
	}

	// Read and verify content
	content, err := svc.ReadIndex()
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}

	// Check for key fields
	expectedFields := []string{
		"## Image: test-id-123",
		"**Title:** Test Image",
		"**Artist:** Test Artist",
		"**Type:** 2D",
		"**Category:** test/category",
		"**File Path:** categories/test/image.jpg",
		"**Thumbnail:** categories/test/image_thumb.jpg",
		"**Dimensions:** 1920x1080",
		"**Manual Tags:** tag1, tag2",
		"**AI Analysis:**",
		"**Description:** A test image",
		"**Primary Category:** test",
		"**Objects Detected:** object1, object2",
		"**Dominant Colors:** red, blue",
		"**Scene Type:** indoor",
		"**Mood:** calm",
		"**Style:** realistic",
		"**Lighting:** natural",
	}

	for _, field := range expectedFields {
		if !strings.Contains(content, field) {
			t.Errorf("index does not contain expected field: %s", field)
		}
	}
}

func TestAppendToIndex_3DObject(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewIndexService(tempDir)

	// Initialize index
	err := svc.InitializeIndex()
	if err != nil {
		t.Fatalf("InitializeIndex failed: %v", err)
	}

	// Create test 3D object
	now := time.Now()
	image := &models.Image{
		ID:         "test-3d-123",
		Title:      "Test 3D Object",
		Artist:     "Test Artist",
		Type:       models.ImageType3D,
		UploadedAt: now,
		ProcessedAt: &now,
		Status:     "completed",
		FolderPath: "categories/3d/test-3d-123",
		Views: map[string]string{
			"front":  "categories/3d/test-3d-123/front.jpg",
			"back":   "categories/3d/test-3d-123/back.jpg",
			"left":   "categories/3d/test-3d-123/left.jpg",
			"right":  "categories/3d/test-3d-123/right.jpg",
			"top":    "categories/3d/test-3d-123/top.jpg",
			"bottom": "categories/3d/test-3d-123/bottom.jpg",
		},
		TotalFileSize: 6144000,
		Category:      "3d/objects",
		ManualTags:    []string{"3d", "model"},
		AIAnalysis: &models.AIAnalysis{
			Description:            "A 3D test object",
			PrimaryCategory:        "3d",
			SubCategory:            "objects",
			ThreeDCharacteristics:  "cube-like shape",
		},
	}

	err = svc.AppendToIndex(image)
	if err != nil {
		t.Fatalf("AppendToIndex failed: %v", err)
	}

	// Read and verify content
	content, err := svc.ReadIndex()
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}

	// Check for key fields
	expectedFields := []string{
		"## Image: test-3d-123",
		"**Title:** Test 3D Object",
		"**Type:** 3D",
		"**Folder Path:** categories/3d/test-3d-123",
		"**Views:**",
		"front: categories/3d/test-3d-123/front.jpg",
		"**Total File Size:**",
		"(6 images)",
		"**3D Characteristics:** cube-like shape",
	}

	for _, field := range expectedFields {
		if !strings.Contains(content, field) {
			t.Errorf("index does not contain expected field: %s", field)
		}
	}
}

func TestAppendToIndex_Multiple(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewIndexService(tempDir)

	// Initialize index
	err := svc.InitializeIndex()
	if err != nil {
		t.Fatalf("InitializeIndex failed: %v", err)
	}

	// Append multiple images
	now := time.Now()
	for i := 0; i < 3; i++ {
		image := &models.Image{
			ID:            string(rune('a' + i)),
			Title:         "Image " + string(rune('A' + i)),
			Artist:        "Artist",
			Type:          models.ImageType2D,
			UploadedAt:    now,
			ProcessedAt:   &now,
			Status:        "completed",
			FilePath:      "test.jpg",
			ThumbnailPath: "test_thumb.jpg",
			FileSize:      1024,
			Width:         100,
			Height:        100,
			Category:      "test",
		}

		err = svc.AppendToIndex(image)
		if err != nil {
			t.Fatalf("AppendToIndex failed for image %d: %v", i, err)
		}
	}

	// Read and verify all entries exist
	content, err := svc.ReadIndex()
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		expectedID := "## Image: " + string(rune('a' + i))
		expectedTitle := "**Title:** Image " + string(rune('A' + i))

		if !strings.Contains(content, expectedID) {
			t.Errorf("index does not contain image ID: %s", expectedID)
		}
		if !strings.Contains(content, expectedTitle) {
			t.Errorf("index does not contain image title: %s", expectedTitle)
		}
	}
}

func TestReadIndex_NonExistent(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewIndexService(tempDir)

	// Try to read before initializing
	_, err := svc.ReadIndex()
	if err == nil {
		t.Error("expected error when reading non-existent index, got nil")
	}
}

func TestBuildMarkdownEntry_NoAIAnalysis(t *testing.T) {
	svc := NewIndexService("/test")

	now := time.Now()
	image := &models.Image{
		ID:            "test-123",
		Title:         "Test",
		Artist:        "Artist",
		Type:          models.ImageType2D,
		UploadedAt:    now,
		FilePath:      "test.jpg",
		ThumbnailPath: "test_thumb.jpg",
		FileSize:      1024,
		Width:         100,
		Height:        100,
		Category:      "test",
		AIAnalysis:    nil, // No AI analysis
	}

	entry := svc.buildMarkdownEntry(image)

	// Should not contain AI Analysis section
	if strings.Contains(entry, "**AI Analysis:**") {
		t.Error("entry should not contain AI Analysis section when AIAnalysis is nil")
	}

	// Should contain basic fields
	if !strings.Contains(entry, "## Image: test-123") {
		t.Error("entry does not contain image ID")
	}
}

func TestBuildMarkdownEntry_NoManualTags(t *testing.T) {
	svc := NewIndexService("/test")

	now := time.Now()
	image := &models.Image{
		ID:            "test-123",
		Title:         "Test",
		Artist:        "Artist",
		Type:          models.ImageType2D,
		UploadedAt:    now,
		FilePath:      "test.jpg",
		ThumbnailPath: "test_thumb.jpg",
		FileSize:      1024,
		Width:         100,
		Height:        100,
		Category:      "test",
		ManualTags:    []string{}, // No manual tags
	}

	entry := svc.buildMarkdownEntry(image)

	// Should not contain Manual Tags section
	if strings.Contains(entry, "**Manual Tags:**") {
		t.Error("entry should not contain Manual Tags section when ManualTags is empty")
	}
}
