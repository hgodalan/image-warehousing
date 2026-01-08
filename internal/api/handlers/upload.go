package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourcompany/image-warehousing/internal/models"
	"github.com/yourcompany/image-warehousing/internal/service"
)

type UploadHandler struct {
	storageService *service.StorageService
	imageService   *service.ImageService
	maxUploadSize  int64
}

func NewUploadHandler(storage *service.StorageService, image *service.ImageService, maxSize int64) *UploadHandler {
	return &UploadHandler{
		storageService: storage,
		imageService:   image,
		maxUploadSize:  maxSize,
	}
}

func (h *UploadHandler) Handle2DUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get metadata
	title := r.FormValue("title")
	artist := r.FormValue("artist")

	// Parse tags (JSON array)
	var tags []string
	tagsStr := r.FormValue("tags")
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			http.Error(w, "Invalid tags format", http.StatusBadRequest)
			return
		}
	}

	// Validate required fields
	if title == "" || artist == "" {
		http.Error(w, "Title and artist are required", http.StatusBadRequest)
		return
	}

	// Save to temp
	imageID, tempPath, err := h.storageService.SaveImageToTemp(file, header.Filename)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	// Queue job for processing
	job := &models.UploadJob{
		ImageID:    imageID,
		Type:       models.ImageType2D,
		FilePath:   tempPath,
		Title:      title,
		Artist:     artist,
		ManualTags: tags,
	}

	if err := h.imageService.QueueJob(job); err != nil {
		http.Error(w, "Failed to queue job", http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"id":     imageID,
		"status": "processing",
		"message": "Image uploaded successfully and is being processed",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
