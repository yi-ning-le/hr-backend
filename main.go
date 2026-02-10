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
		api.POST("/candidates/:id/assign-reviewer", candidateHandler.AssignReviewer)
		api.POST("/candidates/:id/review", candidateHandler.SubmitReview)

		// Candidate Status Routes
		api.GET("/candidate-statuses", candidateStatusHandler.ListStatuses)
		api.POST("/candidate-statuses", candidateStatusHandler.CreateStatus)
		api.PUT("/candidate-statuses/:id", candidateStatusHandler.UpdateStatus)
		api.DELETE("/candidate-statuses/:id", candidateStatusHandler.DeleteStatus)
		api.PATCH("/candidate-statuses/reorder", candidateStatusHandler.ReorderStatuses)

		// Employee Routes (Read - All authenticated users)
		api.GET("/employees/me", employeeHandler.GetCurrentEmployee)
		api.GET("/employees", employeeHandler.ListEmployees)
		api.GET("/employees/:id", employeeHandler.GetEmployee)

		// Recruitment Role Routes
		api.GET("/recruitment/role", recruitmentHandler.GetMyRole)
	}

	// HR-only Employee Routes (Create, Update, Delete)
	hrQuerier := middleware.NewQueriesAdapter(repo)
	hrApi := r.Group("/")
	hrApi.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	hrApi.Use(middleware.RequireHR(hrQuerier))
	{
		hrApi.POST("/employees", employeeHandler.CreateEmployee)
		hrApi.PUT("/employees/:id", employeeHandler.UpdateEmployee)
		hrApi.DELETE("/employees/:id", employeeHandler.DeleteEmployee)
	}

	// Admin only Recruitment Routes
	adminApi := r.Group("/recruitment/admin")
	adminApi.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	adminApi.Use(middleware.RequireAdmin(repo))
	{
		adminApi.GET("/recruiters", recruitmentHandler.GetRecruiters)
		adminApi.POST("/recruiters", recruitmentHandler.AssignRecruiter)
		adminApi.DELETE("/recruiters", recruitmentHandler.RevokeRecruiter)
		// HR management routes
		adminApi.GET("/hrs", recruitmentHandler.GetHRs)
		adminApi.POST("/hrs", recruitmentHandler.AssignHR)
		adminApi.DELETE("/hrs", recruitmentHandler.RevokeHR)
	}

	// Recruiter only Routes
	recruiterApi := r.Group("/recruitment")
	recruiterApi.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	recruiterApi.Use(middleware.RequireRecruiter(repo))
	{
		recruiterApi.POST("/interviews", recruitmentHandler.CreateInterview)
		recruiterApi.POST("/interviews/:id/transfer", recruitmentHandler.TransferInterview)
	}

	// Interviewer Routes (Employee access)
	// Technically any employee can be an interviewer, so we just check auth.
	// We could add a RequireInterviewer middleware if we wanted to be strict,
	// but GetMyInterviews implicitly filters by user.
	interviewApi := r.Group("/recruitment")
	interviewApi.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		interviewApi.GET("/interviews/me", recruitmentHandler.GetMyInterviews)
		interviewApi.GET("/interviews/:id", recruitmentHandler.GetInterview)
		interviewApi.PATCH("/interviews/:id/notes", recruitmentHandler.UpdateInterviewNotes)
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
