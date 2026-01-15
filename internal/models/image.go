package models

import "time"

type ImageType string

const (
	ImageType2D ImageType = "2D"
	ImageType3D ImageType = "3D"
)

// Image represents both 2D and 3D images
type Image struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Artist           string    `json:"artist"`
	Type             ImageType `json:"type"`
	UploadedAt       time.Time `json:"uploaded_at"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	Status           string    `json:"status"` // pending, processing, completed, error

	// For 2D images
	OriginalFilename string `json:"original_filename,omitempty"`
	FilePath         string `json:"file_path,omitempty"`
	ThumbnailPath    string `json:"thumbnail_path,omitempty"`
	MimeType         string `json:"mime_type,omitempty"`
	FileSize         int64  `json:"file_size,omitempty"`
	Width            int    `json:"width,omitempty"`
	Height           int    `json:"height,omitempty"`

	// For 3D objects
	FolderPath       string            `json:"folder_path,omitempty"`
	ModelFilePath    string            `json:"model_file_path,omitempty"`    // Path to the 3D model file (.obj, .glb, .fbx, etc.)
	ModelFilename    string            `json:"model_filename,omitempty"`     // Original filename of the 3D model
	Views            map[string]string `json:"views,omitempty"`              // view name -> file path
	TotalFileSize    int64             `json:"total_file_size,omitempty"`

	// Common fields
	Category         string   `json:"category"`
	ManualTags       []string `json:"manual_tags,omitempty"`
	AIAnalysis       *AIAnalysis `json:"ai_analysis,omitempty"`
}

// UploadJob represents a job for the background worker
type UploadJob struct {
	ImageID        string
	Type           ImageType
	FilePath       string            // For 2D
	FilePaths      map[string]string // For 3D (view -> path)
	ModelFilePath  string            // For 3D (the actual 3D model file)
	ModelFilename  string            // For 3D (original model filename)
	Title          string
	Artist         string
	ManualTags     []string
}
