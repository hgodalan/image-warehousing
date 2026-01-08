package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/yourcompany/image-warehousing/internal/models"
)

type IndexService struct {
	indexPath string
	lock      *flock.Flock
}

func NewIndexService(dataDir string) *IndexService {
	indexPath := filepath.Join(dataDir, "index.md")
	return &IndexService{
		indexPath: indexPath,
		lock:      flock.New(indexPath + ".lock"),
	}
}

// InitializeIndex creates the index file if it doesn't exist
func (s *IndexService) InitializeIndex() error {
	if _, err := os.Stat(s.indexPath); os.IsNotExist(err) {
		initialContent := fmt.Sprintf(`# Image Warehouse Index
Last Updated: %s

---
`, time.Now().Format("2006-01-02 15:04:05"))

		return os.WriteFile(s.indexPath, []byte(initialContent), 0644)
	}
	return nil
}

// AppendToIndex adds a new image entry to the index
func (s *IndexService) AppendToIndex(image *models.Image) error {
	// Acquire file lock
	if err := s.lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer s.lock.Unlock()

	// Build the markdown entry
	entry := s.buildMarkdownEntry(image)

	// Append to file
	f, err := os.OpenFile(s.indexPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to index: %w", err)
	}

	return nil
}

// ReadIndex returns the entire index content
func (s *IndexService) ReadIndex() (string, error) {
	content, err := os.ReadFile(s.indexPath)
	if err != nil {
		return "", fmt.Errorf("failed to read index: %w", err)
	}
	return string(content), nil
}

// buildMarkdownEntry creates a markdown entry for an image
func (s *IndexService) buildMarkdownEntry(img *models.Image) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n## Image: %s\n\n", img.ID))
	sb.WriteString(fmt.Sprintf("**Title:** %s\n", img.Title))
	sb.WriteString(fmt.Sprintf("**Artist:** %s\n", img.Artist))
	sb.WriteString(fmt.Sprintf("**Uploaded:** %s\n", img.UploadedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", img.Type))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n", img.Category))

	if img.Type == models.ImageType2D {
		sb.WriteString(fmt.Sprintf("**File Path:** %s\n", img.FilePath))
		sb.WriteString(fmt.Sprintf("**Thumbnail:** %s\n", img.ThumbnailPath))
		sb.WriteString(fmt.Sprintf("**Dimensions:** %dx%d\n", img.Width, img.Height))
		sb.WriteString(fmt.Sprintf("**File Size:** %.1f MB\n", float64(img.FileSize)/(1024*1024)))
	} else if img.Type == models.ImageType3D {
		sb.WriteString(fmt.Sprintf("**Folder Path:** %s\n", img.FolderPath))
		sb.WriteString("**Views:**\n")
		for view, path := range img.Views {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", view, path))
		}
		sb.WriteString(fmt.Sprintf("**Total File Size:** %.1f MB (6 images)\n", float64(img.TotalFileSize)/(1024*1024)))
	}

	if len(img.ManualTags) > 0 {
		sb.WriteString(fmt.Sprintf("\n**Manual Tags:** %s\n", strings.Join(img.ManualTags, ", ")))
	}

	if img.AIAnalysis != nil {
		s.writeAIAnalysis(&sb, img.AIAnalysis)
	}

	sb.WriteString("\n---\n")

	return sb.String()
}

// writeAIAnalysis writes the AI analysis section
func (s *IndexService) writeAIAnalysis(sb *strings.Builder, ai *models.AIAnalysis) {
	sb.WriteString("\n**AI Analysis:**\n")
	sb.WriteString(fmt.Sprintf("- **Description:** %s\n", ai.Description))
	sb.WriteString(fmt.Sprintf("- **Primary Category:** %s\n", ai.PrimaryCategory))
	sb.WriteString(fmt.Sprintf("- **Sub-category:** %s\n", ai.SubCategory))

	if len(ai.Objects) > 0 {
		sb.WriteString(fmt.Sprintf("- **Objects Detected:** %s\n", strings.Join(ai.Objects, ", ")))
	}

	if len(ai.Colors) > 0 {
		sb.WriteString(fmt.Sprintf("- **Dominant Colors:** %s\n", strings.Join(ai.Colors, ", ")))
	}

	if ai.SceneType != "" {
		sb.WriteString(fmt.Sprintf("- **Scene Type:** %s\n", ai.SceneType))
	}

	if ai.Mood != "" {
		sb.WriteString(fmt.Sprintf("- **Mood:** %s\n", ai.Mood))
	}

	if ai.Style != "" {
		sb.WriteString(fmt.Sprintf("- **Style:** %s\n", ai.Style))
	}

	if ai.Lighting != "" {
		sb.WriteString(fmt.Sprintf("- **Lighting:** %s\n", ai.Lighting))
	}

	if len(ai.Features) > 0 {
		featuresStr := make([]string, len(ai.Features))
		for i, f := range ai.Features {
			featuresStr[i] = fmt.Sprintf("%s (%.2f)", f.Name, f.Confidence)
		}
		sb.WriteString(fmt.Sprintf("- **AI Features:** %s\n", strings.Join(featuresStr, ", ")))
	}

	if ai.ThreeDCharacteristics != "" {
		sb.WriteString(fmt.Sprintf("- **3D Characteristics:** %s\n", ai.ThreeDCharacteristics))
	}
}

// ImageMetadata represents simplified image metadata for listing
type ImageMetadata struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Artist        string   `json:"artist"`
	Category      string   `json:"category"`
	ThumbnailPath string   `json:"thumbnail_path,omitempty"`
	FilePath      string   `json:"file_path,omitempty"`
	Description   string   `json:"description,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	UploadedAt    string   `json:"uploaded_at"`
}

// GetAllImages parses the index and returns all images
func (s *IndexService) GetAllImages() ([]*ImageMetadata, error) {
	content, err := s.ReadIndex()
	if err != nil {
		return nil, err
	}

	var images []*ImageMetadata

	// Split by image entries
	imageRegex := regexp.MustCompile(`(?m)^## Image: (.+)$`)
	matches := imageRegex.FindAllStringSubmatchIndex(content, -1)

	for i, match := range matches {
		start := match[0]
		var end int
		if i < len(matches)-1 {
			end = matches[i+1][0]
		} else {
			end = len(content)
		}

		section := content[start:end]
		imageID := content[match[2]:match[3]]

		img := &ImageMetadata{
			ID: imageID,
		}

		// Parse fields
		img.Title = extractField(section, "Title")
		img.Artist = extractField(section, "Artist")
		img.Category = extractField(section, "Category")
		img.ThumbnailPath = normalizePath(extractField(section, "Thumbnail"))
		img.FilePath = normalizePath(extractField(section, "File Path"))
		img.Description = extractField(section, "Description")
		img.UploadedAt = extractField(section, "Uploaded")

		// Extract tags
		if tagsStr := extractField(section, "Manual Tags"); tagsStr != "" {
			img.Tags = strings.Split(tagsStr, ", ")
		}

		images = append(images, img)
	}

	return images, nil
}

// GetImageByID finds a specific image in the index
func (s *IndexService) GetImageByID(imageID string) (*ImageMetadata, error) {
	images, err := s.GetAllImages()
	if err != nil {
		return nil, err
	}

	for _, img := range images {
		if img.ID == imageID {
			return img, nil
		}
	}

	return nil, fmt.Errorf("image not found: %s", imageID)
}

// extractField extracts a field value from markdown content
func extractField(content, fieldName string) string {
	pattern := fmt.Sprintf(`\*\*%s:\*\*\s*(.+)`, regexp.QuoteMeta(fieldName))
	re := regexp.MustCompile(pattern)
	if matches := re.FindStringSubmatch(content); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// normalizePath converts Windows backslashes to forward slashes for web URLs
func normalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
