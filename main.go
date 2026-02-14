package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	authService := service.NewAuthService(repo, cfg.JWTSecret, db.Pool)
	employeeService := service.NewEmployeeService(repo, db.Pool)
	candidateStatusService := service.NewCandidateStatusService(repo)
	candidateCommentService := service.NewCandidateCommentService(repo)

	// 4.1 Start Background Tasks
	backgroundCtx, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()

	go authService.StartCleanupTask(backgroundCtx)

	// 5. Initialize Handlers
	jobHandler := handler.NewJobHandler(jobService)
	candidateHandler := handler.NewCandidateHandler(candidateService)
	authHandler := handler.NewAuthHandler(authService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	candidateStatusHandler := handler.NewCandidateStatusHandler(candidateStatusService)
	recruitmentHandler := handler.NewRecruitmentHandler(repo)
	candidateCommentHandler := handler.NewCandidateCommentHandler(candidateCommentService)

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
		auth.POST("/refresh-token", authHandler.RefreshToken)
	}

	// Protected Routes (API)
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	{
		// Auth Routes
		api.GET("/auth/sessions", authHandler.ListSessions)
		api.DELETE("/auth/sessions/:id", authHandler.DeleteSession)
		api.GET("/auth/ping", authHandler.Ping)
		api.POST("/auth/logout", authHandler.Logout)

		// Job Routes
		api.GET("/jobs", jobHandler.ListJobs)

		// Candidate Routes
		api.GET("/candidates", candidateHandler.ListCandidates)
		api.GET("/candidates/counts", candidateHandler.GetCandidateCounts)
		api.GET("/candidates/:id", candidateHandler.GetCandidate)

		// Candidate Status Routes
		api.GET("/candidate-statuses", candidateStatusHandler.ListStatuses)

		// Employee Routes (Read - All authenticated users)
		api.GET("/employees/me", employeeHandler.GetCurrentEmployee)
		api.GET("/employees", employeeHandler.ListEmployees)
		api.GET("/employees/:id", employeeHandler.GetEmployee)

		// Recruitment Role Routes
		api.GET("/recruitment/role", recruitmentHandler.GetMyRole)
	}

	recruitmentWriteAPI := r.Group("/")
	recruitmentWriteAPI.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	recruitmentWriteAPI.Use(middleware.RequireRecruiter(repo))
	{
		// Job write routes
		recruitmentWriteAPI.POST("/jobs", jobHandler.CreateJob)
		recruitmentWriteAPI.PUT("/jobs/:id", jobHandler.UpdateJob)
		recruitmentWriteAPI.DELETE("/jobs/:id", jobHandler.DeleteJob)
		recruitmentWriteAPI.PATCH("/jobs/:id/status", jobHandler.ToggleStatus)

		// Candidate write routes (recruiter only; admin strictly isolated)
		recruitmentWriteAPI.POST("/candidates", candidateHandler.CreateCandidate)
		recruitmentWriteAPI.PUT("/candidates/:id", candidateHandler.UpdateCandidate)
		recruitmentWriteAPI.DELETE("/candidates/:id", candidateHandler.DeleteCandidate)
		recruitmentWriteAPI.PATCH("/candidates/:id/status", candidateHandler.UpdateStatus)
		recruitmentWriteAPI.POST("/candidates/:id/resume", candidateHandler.UploadResume)
		recruitmentWriteAPI.POST(
			"/candidates/:id/assign-reviewer",
			candidateHandler.AssignReviewer,
		)

		// Candidate status write routes
		recruitmentWriteAPI.POST("/candidate-statuses", candidateStatusHandler.CreateStatus)
		recruitmentWriteAPI.PUT("/candidate-statuses/:id", candidateStatusHandler.UpdateStatus)
		recruitmentWriteAPI.DELETE("/candidate-statuses/:id", candidateStatusHandler.DeleteStatus)
		recruitmentWriteAPI.PATCH(
			"/candidate-statuses/reorder",
			candidateStatusHandler.ReorderStatuses,
		)
	}

	// HR-only Employee Routes (Create, Update, Delete)
	hrQuerier := middleware.NewQueriesAdapter(repo)
	hrApi := r.Group("/")
	hrApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	hrApi.Use(middleware.RequireHR(hrQuerier))
	{
		hrApi.POST("/employees", employeeHandler.CreateEmployee)
		hrApi.PUT("/employees/:id", employeeHandler.UpdateEmployee)
		hrApi.DELETE("/employees/:id", employeeHandler.DeleteEmployee)
	}

	// Reviewer-only action (resolved by candidate assignment check in handler/service)
	reviewAPI := r.Group("/")
	reviewAPI.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	reviewAPI.Use(middleware.RequireInterviewerOrRecruiter(hrQuerier))
	reviewAPI.Use(middleware.RequireCandidateReviewer(hrQuerier))
	{
		reviewAPI.POST("/candidates/:id/review", candidateHandler.SubmitReview)

		// Candidate Comments
		reviewAPI.GET("/candidates/:id/comments", candidateCommentHandler.ListComments)
		reviewAPI.POST("/candidates/:id/comments", candidateCommentHandler.CreateComment)
	}

	// Comment delete uses service-level author/HR policy.
	commentAPI := r.Group("/")
	commentAPI.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	{
		commentAPI.DELETE("/comments/:commentId", candidateCommentHandler.DeleteComment)
	}

	// Admin only Recruitment Routes
	adminApi := r.Group("/recruitment/admin")
	adminApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
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
	recruiterApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	recruiterApi.Use(middleware.RequireRecruiter(repo))
	{
		recruiterApi.GET("/interviews", recruitmentHandler.GetAllInterviews)
		recruiterApi.POST("/interviews", recruitmentHandler.CreateInterview)
		recruiterApi.PUT("/interviews/:id", recruitmentHandler.UpdateInterview)
		recruiterApi.POST("/interviews/:id/transfer", recruitmentHandler.TransferInterview)
	}

	// Interviewer Routes (Employee access)
	interviewerQueries := middleware.NewQueriesAdapter(repo)
	interviewApi := r.Group("/recruitment")
	interviewApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	interviewApi.Use(middleware.RequireInterviewerOrRecruiter(interviewerQueries))
	{
		interviewApi.GET("/interviews/me", recruitmentHandler.GetMyInterviews)
		interviewApi.GET("/interviews/:id", recruitmentHandler.GetInterview)
		interviewApi.PATCH("/interviews/:id/status", recruitmentHandler.UpdateInterviewStatus)
	}

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serverErr:
		log.Fatalf("Server failed to start: %v", err)
	case sig := <-quit:
		log.Printf("Received signal: %s", sig.String())
	}

	log.Println("Shutting down server...")
	cancelBackground()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server exited")
}
