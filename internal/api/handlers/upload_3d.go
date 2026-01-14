package handlers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/yourcompany/image-warehousing/internal/models"
	"github.com/yourcompany/image-warehousing/internal/service"
)

type Upload3DHandler struct {
	storageService *service.StorageService
	imageService   *service.ImageService
	maxUploadSize  int64
}

func NewUpload3DHandler(storage *service.StorageService, image *service.ImageService, maxSize int64) *Upload3DHandler {
	return &Upload3DHandler{
		storageService: storage,
		imageService:   image,
		maxUploadSize:  maxSize * 6, // 6 images
	}
}

func (h *Upload3DHandler) Handle3DUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (larger size for 6 images)
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		http.Error(w, "Files too large or invalid form", http.StatusBadRequest)
		return
	}

	// Get the 3D model file (required)
	modelFile, modelHeader, err := r.FormFile("model")
	if err != nil {
		http.Error(w, "Missing 3D model file", http.StatusBadRequest)
		return
	}
	defer modelFile.Close()

	// Check if 4-surface or 6-surface mode
	mode := r.FormValue("mode") // "4" or "6"
	var views []string

	if mode == "4" {
		views = []string{"front", "back", "left", "right"}
	} else {
		// Default to 6-surface mode
		views = []string{"front", "back", "left", "right", "top", "bottom"}
	}

	viewFiles := make(map[string]multipart.File)
	viewFilenames := make(map[string]string)

	for _, view := range views {
		file, header, err := r.FormFile(view)
		if err != nil {
			http.Error(w, "Missing view: "+view, http.StatusBadRequest)
			return
		}
		viewFiles[view] = file
		viewFilenames[view] = header.Filename
	}

	// Defer closing all files
	defer func() {
		for _, file := range viewFiles {
			file.Close()
		}
	}()

	// Get metadata
	title := r.FormValue("title")
	artist := r.FormValue("artist")

	// Parse tags
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

	// Save to temp (including model file)
	imageID, tempPaths, modelPath, err := h.storageService.Save3DObjectToTemp(modelFile, modelHeader.Filename, viewFiles, viewFilenames)
	if err != nil {
		http.Error(w, "Failed to save 3D object: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Queue job for processing
	job := &models.UploadJob{
		ImageID:        imageID,
		Type:           models.ImageType3D,
		FilePaths:      tempPaths,
		ModelFilePath:  modelPath,
		ModelFilename:  modelHeader.Filename,
		Title:          title,
		Artist:         artist,
		ManualTags:     tags,
	}

	if err := h.imageService.QueueJob(job); err != nil {
		http.Error(w, "Failed to queue job", http.StatusInternalServerError)
		return
	}

	// Return response
	viewCount := len(views)
	var viewsText string
	if viewCount == 4 {
		viewsText = "4 views"
	} else {
		viewsText = "6 views"
	}
	response := map[string]interface{}{
		"id":      imageID,
		"status":  "processing",
		"message": "3D object uploaded successfully (" + viewsText + ") and is being processed",
		"views":   viewCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
