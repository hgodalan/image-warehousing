package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourcompany/image-warehousing/internal/service"
)

type ImagesHandler struct {
	imageService *service.ImageService
	indexService *service.IndexService
}

func NewImagesHandler(image *service.ImageService, index *service.IndexService) *ImagesHandler {
	return &ImagesHandler{
		imageService: image,
		indexService: index,
	}
}

// HandleListImages lists all images from the index
func (h *ImagesHandler) HandleListImages(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	category := r.URL.Query().Get("category")

	// Get all images from index
	images, err := h.indexService.GetAllImages()
	if err != nil {
		http.Error(w, "Failed to load images", http.StatusInternalServerError)
		return
	}

	// Filter by category if specified
	if category != "" {
		var filtered []*service.ImageMetadata
		for _, img := range images {
			if img.Category == category {
				filtered = append(filtered, img)
			}
		}
		images = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"images": images,
		"total":  len(images),
	})
}

// HandleGetImage gets a single image by ID
func (h *ImagesHandler) HandleGetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	if imageID == "" {
		http.Error(w, "Image ID required", http.StatusBadRequest)
		return
	}

	// First try in-memory status
	image, err := h.imageService.GetStatus(imageID)
	if err != nil {
		// If not in memory, try to get from index
		metadata, err := h.indexService.GetImageByID(imageID)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}
