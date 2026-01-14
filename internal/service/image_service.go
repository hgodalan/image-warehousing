package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourcompany/image-warehousing/internal/models"
)

type ImageService struct {
	storageService *StorageService
	aiService      *AIService
	indexService   *IndexService
	jobQueue       chan *models.UploadJob
	statusMap      map[string]*models.Image
	statusMutex    sync.RWMutex
	logger         *logrus.Logger
}

func NewImageService(storage *StorageService, ai *AIService, index *IndexService, logger *logrus.Logger) *ImageService {
	return &ImageService{
		storageService: storage,
		aiService:      ai,
		indexService:   index,
		jobQueue:       make(chan *models.UploadJob, 100),
		statusMap:      make(map[string]*models.Image),
		logger:         logger,
	}
}

// StartWorkers starts the background workers
func (s *ImageService) StartWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go s.worker(i)
	}
	s.logger.Infof("Started %d worker goroutines", numWorkers)
}

// QueueJob adds a job to the processing queue
func (s *ImageService) QueueJob(job *models.UploadJob) error {
	// Initialize status
	s.statusMutex.Lock()
	s.statusMap[job.ImageID] = &models.Image{
		ID:         job.ImageID,
		Title:      job.Title,
		Artist:     job.Artist,
		Type:       job.Type,
		Status:     "processing",
		UploadedAt: time.Now(),
		ManualTags: job.ManualTags,
	}
	s.statusMutex.Unlock()

	// Add to queue
	select {
	case s.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetStatus returns the current status of an image
func (s *ImageService) GetStatus(imageID string) (*models.Image, error) {
	s.statusMutex.RLock()
	defer s.statusMutex.RUnlock()

	if img, ok := s.statusMap[imageID]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("image not found")
}

// worker processes jobs from the queue
func (s *ImageService) worker(id int) {
	s.logger.Infof("Worker %d started", id)

	for job := range s.jobQueue {
		s.logger.Infof("Worker %d processing job for image %s (type: %s)", id, job.ImageID, job.Type)

		var err error
		if job.Type == models.ImageType2D {
			err = s.process2DJob(job)
		} else if job.Type == models.ImageType3D {
			err = s.process3DJob(job)
		} else {
			err = fmt.Errorf("unknown job type: %s", job.Type)
		}

		if err != nil {
			s.logger.Errorf("Worker %d failed to process job %s: %v", id, job.ImageID, err)
			s.updateStatus(job.ImageID, "error")
		} else {
			s.logger.Infof("Worker %d completed job %s", id, job.ImageID)
			s.updateStatus(job.ImageID, "completed")
		}
	}
}

// process2DJob processes a 2D image job
func (s *ImageService) process2DJob(job *models.UploadJob) error {
	ctx := context.Background()

	// 1. Generate thumbnail
	s.logger.Infof("Generating thumbnail for %s", job.ImageID)
	_, err := s.storageService.GenerateThumbnail(job.FilePath)
	if err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// 2. Get image dimensions
	width, height, err := s.storageService.GetImageDimensions(job.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get dimensions: %w", err)
	}

	// 3. Get file size
	fileSize, err := s.storageService.GetFileSize(job.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// 4. Analyze with AI
	s.logger.Infof("Analyzing 2D image %s with Gemini", job.ImageID)
	analysis, err := s.aiService.Analyze2DImage(ctx, job.FilePath)
	if err != nil {
		return fmt.Errorf("failed to analyze image: %w", err)
	}

	// 5. Determine category path
	categoryPath := s.aiService.GetCategoryPath(analysis)
	s.logger.Infof("Image %s categorized as: %s", job.ImageID, categoryPath)

	// 6. Move to category folder
	filePath, thumbPathFinal, err := s.storageService.MoveToCategory(job.ImageID, job.FilePath, categoryPath)
	if err != nil {
		return fmt.Errorf("failed to move to category: %w", err)
	}

	// 7. Update image metadata
	now := time.Now()
	image := &models.Image{
		ID:            job.ImageID,
		Title:         job.Title,
		Artist:        job.Artist,
		Type:          models.ImageType2D,
		UploadedAt:    time.Now(),
		ProcessedAt:   &now,
		Status:        "completed",
		FilePath:      filePath,
		ThumbnailPath: thumbPathFinal,
		FileSize:      fileSize,
		Width:         width,
		Height:        height,
		Category:      categoryPath,
		ManualTags:    job.ManualTags,
		AIAnalysis:    analysis,
	}

	// 8. Append to index
	s.logger.Infof("Adding image %s to index", job.ImageID)
	if err := s.indexService.AppendToIndex(image); err != nil {
		return fmt.Errorf("failed to append to index: %w", err)
	}

	// 9. Update in-memory status
	s.statusMutex.Lock()
	s.statusMap[job.ImageID] = image
	s.statusMutex.Unlock()

	return nil
}

// process3DJob processes a 3D object job
func (s *ImageService) process3DJob(job *models.UploadJob) error {
	ctx := context.Background()

	// 1. Generate thumbnails for all views (4 or 6)
	s.logger.Infof("Generating thumbnails for 3D object %s", job.ImageID)
	_, err := s.storageService.GenerateThumbnails3D(job.FilePaths)
	if err != nil {
		return fmt.Errorf("failed to generate thumbnails: %w", err)
	}

	// 2. Calculate total file size (including model file)
	var totalSize int64
	for _, path := range job.FilePaths {
		size, err := s.storageService.GetFileSize(path)
		if err != nil {
			return fmt.Errorf("failed to get file size: %w", err)
		}
		totalSize += size
	}
	// Add model file size
	if job.ModelFilePath != "" {
		modelSize, err := s.storageService.GetFileSize(job.ModelFilePath)
		if err != nil {
			return fmt.Errorf("failed to get model file size: %w", err)
		}
		totalSize += modelSize
	}

	// 3. Analyze with AI (all surface views together)
	viewCount := len(job.FilePaths)
	s.logger.Infof("Analyzing 3D object %s with Gemini (%d views)", job.ImageID, viewCount)
	analysis, err := s.aiService.Analyze3DObject(ctx, job.FilePaths)
	if err != nil {
		return fmt.Errorf("failed to analyze 3D object: %w", err)
	}

	// 4. Determine category path
	categoryPath := s.aiService.GetCategoryPath(analysis)
	s.logger.Infof("3D object %s categorized as: %s", job.ImageID, categoryPath)

	// 5. Move to category folder
	folderPath, modelPath, views, err := s.storageService.Move3DToCategory(job.ImageID, "", categoryPath)
	if err != nil {
		return fmt.Errorf("failed to move to category: %w", err)
	}

	// 6. Update image metadata
	now := time.Now()
	image := &models.Image{
		ID:            job.ImageID,
		Title:         job.Title,
		Artist:        job.Artist,
		Type:          models.ImageType3D,
		UploadedAt:    time.Now(),
		ProcessedAt:   &now,
		Status:        "completed",
		FolderPath:    folderPath,
		ModelFilePath: modelPath,
		ModelFilename: job.ModelFilename,
		Views:         views,
		TotalFileSize: totalSize,
		Category:      categoryPath,
		ManualTags:    job.ManualTags,
		AIAnalysis:    analysis,
	}

	// 7. Append to index
	s.logger.Infof("Adding 3D object %s to index", job.ImageID)
	if err := s.indexService.AppendToIndex(image); err != nil {
		return fmt.Errorf("failed to append to index: %w", err)
	}

	// 8. Update in-memory status
	s.statusMutex.Lock()
	s.statusMap[job.ImageID] = image
	s.statusMutex.Unlock()

	return nil
}

// updateStatus updates the status of an image
func (s *ImageService) updateStatus(imageID, status string) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if img, ok := s.statusMap[imageID]; ok {
		img.Status = status
		if status == "completed" {
			now := time.Now()
			img.ProcessedAt = &now
		}
	}
}
