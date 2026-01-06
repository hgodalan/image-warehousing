package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/yourcompany/image-warehousing/internal/api/handlers"
	"github.com/yourcompany/image-warehousing/internal/api/middleware"
	"github.com/yourcompany/image-warehousing/internal/config"
	"github.com/yourcompany/image-warehousing/internal/service"
)

type Router struct {
	router         *mux.Router
	uploadHandler  *handlers.UploadHandler
	upload3DHandler *handlers.Upload3DHandler
	searchHandler  *handlers.SearchHandler
	healthHandler  *handlers.HealthHandler
}

func NewRouter(
	cfg *config.Config,
	storageService *service.StorageService,
	imageService   *service.ImageService,
	searchService  *service.SearchService,
	logger         *logrus.Logger,
) *Router {
	r := mux.NewRouter()

	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(storageService, imageService, cfg.MaxUploadSize)
	upload3DHandler := handlers.NewUpload3DHandler(storageService, imageService, cfg.MaxUploadSize)
	searchHandler := handlers.NewSearchHandler(searchService)
	healthHandler := handlers.NewHealthHandler()

	// Apply global middleware
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Upload endpoints
	api.HandleFunc("/images/upload", uploadHandler.Handle2DUpload).Methods("POST")
	api.HandleFunc("/images/upload-3d", upload3DHandler.Handle3DUpload).Methods("POST")

	// Search endpoint
	api.HandleFunc("/search", searchHandler.HandleSearch).Methods("POST")

	// Health check
	api.HandleFunc("/health", healthHandler.HandleHealth).Methods("GET")
	r.HandleFunc("/health", healthHandler.HandleHealth).Methods("GET") // Also at root

	return &Router{
		router:          r,
		uploadHandler:   uploadHandler,
		upload3DHandler: upload3DHandler,
		searchHandler:   searchHandler,
		healthHandler:   healthHandler,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
