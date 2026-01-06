package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourcompany/image-warehousing/internal/models"
	"github.com/yourcompany/image-warehousing/internal/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(search *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: search,
	}
}

func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Perform search
	results, err := h.searchService.Search(r.Context(), req.Query, req.Limit)
	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
