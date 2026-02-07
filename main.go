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
	authService := service.NewAuthService(repo, cfg.JWTSecret)
	employeeService := service.NewEmployeeService(repo)
	candidateStatusService := service.NewCandidateStatusService(repo)

	// 5. Initialize Handlers
	jobHandler := handler.NewJobHandler(jobService)
	candidateHandler := handler.NewCandidateHandler(candidateService)
	authHandler := handler.NewAuthHandler(authService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	candidateStatusHandler := handler.NewCandidateStatusHandler(candidateStatusService)
	recruitmentHandler := handler.NewRecruitmentHandler(repo)

	// 6. Setup Router
	r := gin.Default()

	// Global Middleware
	r.Use(middleware.CORSMiddleware())

	// Static File Serving
	r.Static("/static/resumes", "./uploads")

	// --- Routes ---

	// Public Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected Routes (API)
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
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

		// Candidate Status Routes
		api.GET("/candidate-statuses", candidateStatusHandler.ListStatuses)
		api.POST("/candidate-statuses", candidateStatusHandler.CreateStatus)
		api.PUT("/candidate-statuses/:id", candidateStatusHandler.UpdateStatus)
		api.DELETE("/candidate-statuses/:id", candidateStatusHandler.DeleteStatus)
		api.PATCH("/candidate-statuses/reorder", candidateStatusHandler.ReorderStatuses)

		// Employee Routes
		api.GET("/employees", employeeHandler.ListEmployees)
		api.POST("/employees", employeeHandler.CreateEmployee)
		api.GET("/employees/:id", employeeHandler.GetEmployee)
		api.PUT("/employees/:id", employeeHandler.UpdateEmployee)
		api.DELETE("/employees/:id", employeeHandler.DeleteEmployee)

		// Recruitment Role Routes
		api.GET("/recruitment/role", recruitmentHandler.GetMyRole)
	}

	// Admin only Recruitment Routes
	adminApi := r.Group("/recruitment/admin")
	adminApi.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	adminApi.Use(middleware.RequireAdmin(repo))
	{
		adminApi.GET("/recruiters", recruitmentHandler.GetRecruiters)
		adminApi.POST("/recruiters", recruitmentHandler.AssignRecruiter)
		adminApi.DELETE("/recruiters", recruitmentHandler.RevokeRecruiter)
	}

	// Recruiter only Routes
	recruiterApi := r.Group("/recruitment")
	recruiterApi.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	recruiterApi.Use(middleware.RequireRecruiter(repo))
	{
		recruiterApi.POST("/interviews/:id/transfer", recruitmentHandler.TransferInterview)
	}

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
