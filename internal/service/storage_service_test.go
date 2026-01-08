package service

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorageService(t *testing.T) {
	dataDir := "/test/data"
	svc := NewStorageService(dataDir)

	if svc.dataDir != dataDir {
		t.Errorf("expected dataDir %s, got %s", dataDir, svc.dataDir)
	}

	expectedTempDir := filepath.Join(dataDir, "temp")
	if svc.tempDir != expectedTempDir {
		t.Errorf("expected tempDir %s, got %s", expectedTempDir, svc.tempDir)
	}
}

func TestInitialize(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	err := svc.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Check that directories were created
	expectedDirs := []string{
		filepath.Join(tempDir, "temp"),
		filepath.Join(tempDir, "categories"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %s was not created: %v", dir, err)
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Hello, World!")
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	size, err := svc.GetFileSize(testFile)
	if err != nil {
		t.Fatalf("GetFileSize failed: %v", err)
	}

	expectedSize := int64(len(content))
	if size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, size)
	}
}

func TestGetFileSize_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	_, err := svc.GetFileSize(filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestGetImageDimensions(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	// Create a test image
	width, height := 100, 200
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	testImagePath := filepath.Join(tempDir, "test.png")
	f, err := os.Create(testImagePath)
	if err != nil {
		t.Fatalf("failed to create test image file: %v", err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
	f.Close()

	// Test GetImageDimensions
	gotWidth, gotHeight, err := svc.GetImageDimensions(testImagePath)
	if err != nil {
		t.Fatalf("GetImageDimensions failed: %v", err)
	}

	if gotWidth != width {
		t.Errorf("expected width %d, got %d", width, gotWidth)
	}
	if gotHeight != height {
		t.Errorf("expected height %d, got %d", height, gotHeight)
	}
}

func TestGetImageDimensions_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	// Create a non-image file
	testFile := filepath.Join(tempDir, "notanimage.txt")
	err := os.WriteFile(testFile, []byte("not an image"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, _, err = svc.GetImageDimensions(testFile)
	if err == nil {
		t.Error("expected error for non-image file, got nil")
	}
}

func TestGetThumbnailPath(t *testing.T) {
	svc := NewStorageService("/test/data")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "png file",
			input:    "/path/to/image.png",
			expected: "/path/to/image_thumb.jpg",
		},
		{
			name:     "jpg file",
			input:    "/path/to/photo.jpg",
			expected: "/path/to/photo_thumb.jpg",
		},
		{
			name:     "file with multiple dots",
			input:    "/path/to/my.image.png",
			expected: "/path/to/my.image_thumb.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.getThumbnailPath(tt.input)
			// Normalize paths for cross-platform comparison
			gotNorm := filepath.ToSlash(got)
			if gotNorm != tt.expected {
				t.Errorf("getThumbnailPath(%s) = %s, want %s", tt.input, gotNorm, tt.expected)
			}
		})
	}
}

func TestCreateCategoryDir(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	category := "animals/cats"
	err := svc.CreateCategoryDir(category)
	if err != nil {
		t.Fatalf("CreateCategoryDir failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "categories", category)
	info, err := os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("category directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}

func TestCreateCategoryDir_NestedCategory(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewStorageService(tempDir)

	category := "art/digital/fantasy/characters"
	err := svc.CreateCategoryDir(category)
	if err != nil {
		t.Fatalf("CreateCategoryDir failed for nested category: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "categories", category)
	info, err := os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("nested category directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}
