package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yourcompany/image-warehousing/internal/models"
)

type SearchService struct {
	indexService *IndexService
	aiService    *AIService
	logger       *logrus.Logger
}

func NewSearchService(index *IndexService, ai *AIService, logger *logrus.Logger) *SearchService {
	return &SearchService{
		indexService: index,
		aiService:    ai,
		logger:       logger,
	}
}

// Search performs a semantic search using Gemini
func (s *SearchService) Search(ctx context.Context, query string, limit int) (*models.SearchResponse, error) {
	s.logger.Infof("Searching for: %s (limit: %d)", query, limit)

	// 1. Read the entire index
	indexContent, err := s.indexService.ReadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	// 2. Use Gemini to search and rank results
	results, err := s.aiService.SearchImages(ctx, indexContent, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search with AI: %w", err)
	}

	// 3. Apply limit
	if len(results) > limit {
		results = results[:limit]
	}

	response := &models.SearchResponse{
		Results: results,
		Total:   len(results),
		Query:   query,
	}

	s.logger.Infof("Found %d results for query: %s", len(results), query)

	return response, nil
}
