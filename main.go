package main

import (
	"log"
	"net/http"

	"hr-backend/internal/config"
	"hr-backend/internal/handler"
	"hr-backend/internal/middleware"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Database Connection
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// 3. Initialize Repository (SQLC)
	repo := repository.New(db.Pool)

	// 4. Initialize Services
	jobService := service.NewJobService(repo)
	candidateService := service.NewCandidateService(repo)

	// 5. Initialize Handlers
	jobHandler := handler.NewJobHandler(jobService)
	candidateHandler := handler.NewCandidateHandler(candidateService)

	// 6. Setup Router
	r := gin.Default()

	// Global Middleware
	r.Use(middleware.CORSMiddleware())

	// Static File Serving
	r.Static("/static/resumes", "./uploads")

	// Routes
	// We'll group them under /api if we want to match the spec's likely intention,
	// or directly at root if the frontend proxy handles path rewriting.
	// Given the spec says server url is .../api, and paths are /jobs, 
	// usually this means GET .../api/jobs.
	api := r.Group("/") // Change to "/api" if needed, but often proxies strip prefix.
	// Actually, let's just support both or stick to root for simplicity unless requested otherwise.
	// User didn't specify prefix handling. I'll stick to root based on paths.
	// Wait, if I use root, then `http://localhost:8080/jobs` matches.
	
	// Job Routes
	api.GET("/jobs", jobHandler.ListJobs)
	api.POST("/jobs", jobHandler.CreateJob)
	api.PUT("/jobs/:id", jobHandler.UpdateJob)
	api.DELETE("/jobs/:id", jobHandler.DeleteJob)
	api.PATCH("/jobs/:id/status", jobHandler.ToggleStatus)

	// Candidate Routes
	api.GET("/candidates", candidateHandler.ListCandidates)
	api.POST("/candidates", candidateHandler.CreateCandidate)
	api.GET("/candidates/:id", candidateHandler.GetCandidate)
	api.PUT("/candidates/:id", candidateHandler.UpdateCandidate)
	api.DELETE("/candidates/:id", candidateHandler.DeleteCandidate)
	api.PATCH("/candidates/:id/status", candidateHandler.UpdateStatus)
	api.PATCH("/candidates/:id/note", candidateHandler.UpdateNote)
	api.POST("/candidates/:id/resume", candidateHandler.UploadResume)

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start Server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}