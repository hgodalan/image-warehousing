package service

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const (
	ThumbnailSize = 300
)

type StorageService struct {
	dataDir string
	tempDir string
}

func NewStorageService(dataDir string) *StorageService {
	tempDir := filepath.Join(dataDir, "temp")
	return &StorageService{
		dataDir: dataDir,
		tempDir: tempDir,
	}
}

// Initialize creates necessary directories
func (s *StorageService) Initialize() error {
	dirs := []string{
		s.tempDir,
		filepath.Join(s.dataDir, "categories"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// SaveImageToTemp saves a 2D image temporarily and returns the path
func (s *StorageService) SaveImageToTemp(file multipart.File, filename string) (string, string, error) {
	imageID := uuid.New().String()
	ext := filepath.Ext(filename)
	tempPath := filepath.Join(s.tempDir, imageID+ext)

	// Save the file
	outFile, err := os.Create(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	return imageID, tempPath, nil
}

// Save3DObjectToTemp saves 6 views of a 3D object to a temp folder
func (s *StorageService) Save3DObjectToTemp(views map[string]multipart.File, filenames map[string]string) (string, map[string]string, error) {
	imageID := uuid.New().String()
	objectDir := filepath.Join(s.tempDir, imageID)

	// Create object directory
	if err := os.MkdirAll(objectDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create object directory: %w", err)
	}

	paths := make(map[string]string)
	requiredViews := []string{"front", "back", "left", "right", "top", "bottom"}

	// Validate all views are present
	for _, view := range requiredViews {
		if _, ok := views[view]; !ok {
			return "", nil, fmt.Errorf("missing required view: %s", view)
		}
	}

	// Save each view
	for view, file := range views {
		ext := filepath.Ext(filenames[view])
		viewPath := filepath.Join(objectDir, view+ext)

		outFile, err := os.Create(viewPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create file for view %s: %w", view, err)
		}

		if _, err := io.Copy(outFile, file); err != nil {
			outFile.Close()
			return "", nil, fmt.Errorf("failed to save view %s: %w", view, err)
		}
		outFile.Close()

		paths[view] = viewPath
	}

	return imageID, paths, nil
}

// GenerateThumbnail creates a 300x300 thumbnail for a 2D image
func (s *StorageService) GenerateThumbnail(imagePath string) (string, error) {
	// Open the image
	src, err := imaging.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	// Create thumbnail
	thumb := imaging.Fit(src, ThumbnailSize, ThumbnailSize, imaging.Lanczos)

	// Save thumbnail
	thumbPath := s.getThumbnailPath(imagePath)
	if err := imaging.Save(thumb, thumbPath); err != nil {
		return "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return thumbPath, nil
}

// GenerateThumbnails3D creates thumbnails for all 6 views of a 3D object
func (s *StorageService) GenerateThumbnails3D(viewPaths map[string]string) (map[string]string, error) {
	thumbnails := make(map[string]string)

	for view, path := range viewPaths {
		thumb, err := s.GenerateThumbnail(path)
		if err != nil {
			return nil, fmt.Errorf("failed to generate thumbnail for view %s: %w", view, err)
		}
		thumbnails[view] = thumb
	}

	return thumbnails, nil
}

// MoveToCategory moves a 2D image from temp to its category folder
func (s *StorageService) MoveToCategory(imageID, tempPath, category string) (string, string, error) {
	categoryDir := filepath.Join(s.dataDir, "categories", category)
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create category directory: %w", err)
	}

	ext := filepath.Ext(tempPath)
	newPath := filepath.Join(categoryDir, imageID+ext)
	thumbPath := s.getThumbnailPath(tempPath)
	newThumbPath := filepath.Join(categoryDir, imageID+"_thumb.jpg")

	// Move main image
	if err := os.Rename(tempPath, newPath); err != nil {
		return "", "", fmt.Errorf("failed to move image: %w", err)
	}

	// Move thumbnail
	if err := os.Rename(thumbPath, newThumbPath); err != nil {
		// If thumbnail move fails, try to move image back
		os.Rename(newPath, tempPath)
		return "", "", fmt.Errorf("failed to move thumbnail: %w", err)
	}

	// Make paths relative to data dir
	relPath, _ := filepath.Rel(s.dataDir, newPath)
	relThumbPath, _ := filepath.Rel(s.dataDir, newThumbPath)

	return relPath, relThumbPath, nil
}

// Move3DToCategory moves a 3D object folder from temp to its category folder
func (s *StorageService) Move3DToCategory(imageID, tempDir, category string) (string, map[string]string, error) {
	categoryDir := filepath.Join(s.dataDir, "categories", category)
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create category directory: %w", err)
	}

	newObjectDir := filepath.Join(categoryDir, imageID)

	// Move the entire object directory
	if err := os.Rename(filepath.Join(s.tempDir, imageID), newObjectDir); err != nil {
		return "", nil, fmt.Errorf("failed to move object directory: %w", err)
	}

	// Build relative paths for all views
	views := make(map[string]string)
	requiredViews := []string{"front", "back", "left", "right", "top", "bottom"}

	for _, view := range requiredViews {
		// Find the file with this view name (regardless of extension)
		entries, err := os.ReadDir(newObjectDir)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read object directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Base(entry.Name())[:len(view)] == view {
				fullPath := filepath.Join(newObjectDir, entry.Name())
				relPath, _ := filepath.Rel(s.dataDir, fullPath)
				views[view] = relPath
			}
		}
	}

	relFolderPath, _ := filepath.Rel(s.dataDir, newObjectDir)

	return relFolderPath, views, nil
}

// GetImageDimensions returns the width and height of an image
func (s *StorageService) GetImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	return img.Width, img.Height, nil
}

// GetFileSize returns the size of a file in bytes
func (s *StorageService) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	return info.Size(), nil
}

// getThumbnailPath returns the thumbnail path for a given image path
func (s *StorageService) getThumbnailPath(imagePath string) string {
	dir := filepath.Dir(imagePath)
	base := filepath.Base(imagePath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+"_thumb.jpg")
}

// CreateCategoryDir creates a category directory if it doesn't exist
func (s *StorageService) CreateCategoryDir(category string) error {
	categoryPath := filepath.Join(s.dataDir, "categories", category)
	return os.MkdirAll(categoryPath, 0755)
}
