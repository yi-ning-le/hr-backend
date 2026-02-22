package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"hr-backend/internal/config"
	"hr-backend/internal/handler"
	"hr-backend/internal/middleware"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg        *config.Config
	db         *database.Database
	router     *gin.Engine
	httpServer *http.Server

	// Background tasks control
	bgCtx       context.Context
	bgCancel    context.CancelFunc
	authService *service.AuthService
}

func NewServer(cfg *config.Config, db *database.Database) *Server {
	// Initialize Repository
	repo := repository.New(db.Pool)

	// Initialize Services
	authService := service.NewAuthService(repo, cfg.JWTSecret, db.Pool)
	jobService := service.NewJobService(repo)
	candidateService := service.NewCandidateService(repo)
	employeeService := service.NewEmployeeService(repo, db.Pool)
	candidateStatusService := service.NewCandidateStatusService(repo)
	candidateCommentService := service.NewCandidateCommentService(repo)

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(authService)
	jobHandler := handler.NewJobHandler(jobService)
	candidateHandler := handler.NewCandidateHandler(candidateService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	candidateStatusHandler := handler.NewCandidateStatusHandler(candidateStatusService)
	recruitmentHandler := handler.NewRecruitmentHandler(repo)
	candidateCommentHandler := handler.NewCandidateCommentHandler(candidateCommentService)
	notificationService := service.NewNotificationService(repo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// Background Context
	bgCtx, bgCancel := context.WithCancel(context.Background())

	srv := &Server{
		cfg:         cfg,
		db:          db,
		bgCtx:       bgCtx,
		bgCancel:    bgCancel,
		authService: authService,
	}

	// Setup Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Static Files
	// Immutable cache for resumes (UUID based)
	r.Group("/static/resumes").
		Use(middleware.ImmutableCache()).
		StaticFS("/", gin.Dir("./uploads", false))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --- Routes Grouping ---

	// Public Auth
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh-token", authHandler.RefreshToken)
	}

	// Protected API
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	{
		// Auth
		api.GET("/auth/sessions", authHandler.ListSessions)
		api.DELETE("/auth/sessions/:id", authHandler.DeleteSession)
		api.GET("/auth/ping", authHandler.Ping)
		api.POST("/auth/logout", authHandler.Logout)

		// Read-Only Resources
		api.GET("/jobs", jobHandler.ListJobs)
		api.GET("/candidates", candidateHandler.ListCandidates)
		api.GET("/candidates/counts", candidateHandler.GetCandidateCounts)
		api.GET("/candidates/pending", recruitmentHandler.GetPendingReviewCandidates)
		api.GET("/candidates/reviewed", recruitmentHandler.GetPastReviewedCandidates)
		api.GET("/candidates/:id", candidateHandler.GetCandidate)
		api.GET("/candidates/:id/history", recruitmentHandler.GetCandidateHistory)
		api.GET("/candidate-statuses", candidateStatusHandler.ListStatuses)
		api.GET("/employees/me", employeeHandler.GetCurrentEmployee)
		api.GET("/employees", employeeHandler.ListEmployees)
		api.GET("/employees/:id", employeeHandler.GetEmployee)
		api.GET("/recruitment/role", recruitmentHandler.GetMyRole)
		api.GET("/recruitment/interviewers", recruitmentHandler.GetInterviewers)
		api.GET("/recruitment/recruiters", recruitmentHandler.GetRecruiters)
		api.GET("/recruitment/hrs", recruitmentHandler.GetHRs)
		api.DELETE("/comments/:commentId", candidateCommentHandler.DeleteComment) // Uses service-level auth

		// Notifications
		api.GET("/notifications", notificationHandler.GetUserNotifications)
		api.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
		api.PUT("/notifications/:id/read", notificationHandler.MarkAsRead)
		api.PUT("/notifications/read-all", notificationHandler.MarkAllAsRead)
		api.DELETE("/notifications/:id", notificationHandler.DeleteNotification)
	}

	// Recruiter Write Access
	recruiterWrite := r.Group("/")
	recruiterWrite.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	recruiterWrite.Use(middleware.RequireRecruiter(repo))
	{
		// Jobs
		recruiterWrite.POST("/jobs", jobHandler.CreateJob)
		recruiterWrite.PUT("/jobs/:id", jobHandler.UpdateJob)
		recruiterWrite.DELETE("/jobs/:id", jobHandler.DeleteJob)
		recruiterWrite.PATCH("/jobs/:id/status", jobHandler.ToggleStatus)

		// Candidates
		recruiterWrite.POST("/candidates", candidateHandler.CreateCandidate)
		recruiterWrite.PUT("/candidates/:id", candidateHandler.UpdateCandidate)
		recruiterWrite.DELETE("/candidates/:id", candidateHandler.DeleteCandidate)
		recruiterWrite.PATCH("/candidates/:id/status", candidateHandler.UpdateStatus)
		recruiterWrite.POST("/candidates/:id/resume", candidateHandler.UploadResume)
		recruiterWrite.POST("/candidates/:id/assign-reviewer", candidateHandler.AssignReviewer)

		// Statuses
		recruiterWrite.POST("/candidate-statuses", candidateStatusHandler.CreateStatus)
		recruiterWrite.PUT("/candidate-statuses/:id", candidateStatusHandler.UpdateStatus)
		recruiterWrite.DELETE("/candidate-statuses/:id", candidateStatusHandler.DeleteStatus)
		recruiterWrite.PATCH("/candidate-statuses/reorder", candidateStatusHandler.ReorderStatuses)

		// Interview Management
		recruiterWrite.GET("/recruitment/interviews", recruitmentHandler.GetAllInterviews)
		recruiterWrite.POST("/recruitment/interviews", recruitmentHandler.CreateInterview)
		recruiterWrite.PUT("/recruitment/interviews/:id", recruitmentHandler.UpdateInterview)
		recruiterWrite.POST("/recruitment/interviews/:id/transfer", recruitmentHandler.TransferInterview)
	}

	// HR Admin Access
	hrQuerier := middleware.NewQueriesAdapter(repo)
	hrApi := r.Group("/")
	hrApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	hrApi.Use(middleware.RequireHR(hrQuerier))
	{
		hrApi.POST("/employees", employeeHandler.CreateEmployee)
		hrApi.PUT("/employees/:id", employeeHandler.UpdateEmployee)
		hrApi.DELETE("/employees/:id", employeeHandler.DeleteEmployee)
	}

	// Reviewer Access
	reviewAPI := r.Group("/")
	reviewAPI.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	reviewAPI.Use(middleware.RequireInterviewerOrRecruiter(hrQuerier))
	reviewAPI.Use(middleware.RequireCandidateReviewer(hrQuerier))
	{
		reviewAPI.POST("/candidates/:id/review", candidateHandler.SubmitReview)
		reviewAPI.GET("/candidates/:id/comments", candidateCommentHandler.ListComments)
		reviewAPI.POST("/candidates/:id/comments", candidateCommentHandler.CreateComment)
	}

	// Admin Access
	adminApi := r.Group("/recruitment/admin")
	adminApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	adminApi.Use(middleware.RequireAdmin(repo))
	{
		adminApi.POST("/recruiters", recruitmentHandler.AssignRecruiter)
		adminApi.DELETE("/recruiters", recruitmentHandler.RevokeRecruiter)
		adminApi.POST("/hrs", recruitmentHandler.AssignHR)
		adminApi.DELETE("/hrs", recruitmentHandler.RevokeHR)
		adminApi.POST("/interviewers", recruitmentHandler.AssignInterviewer)
		adminApi.DELETE("/interviewers", recruitmentHandler.RevokeInterviewer)
	}

	// Interviewer Access
	interviewerQueries := middleware.NewQueriesAdapter(repo)
	interviewApi := r.Group("/recruitment")
	interviewApi.Use(middleware.AuthMiddleware(cfg.JWTSecret, repo))
	interviewApi.Use(middleware.RequireInterviewerOrRecruiter(interviewerQueries))
	{
		interviewApi.GET("/interviews/me", recruitmentHandler.GetMyInterviews)
		interviewApi.GET("/interviews/:id", recruitmentHandler.GetInterview)
		interviewApi.PATCH("/interviews/:id/status", recruitmentHandler.UpdateInterviewStatus)
	}

	srv.router = r
	return srv
}

func (s *Server) Start() error {
	// Start background tasks
	go s.authService.StartCleanupTask(s.bgCtx)

	s.httpServer = &http.Server{
		Addr:              ":" + s.cfg.ServerPort,
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server starting on port %s", s.cfg.ServerPort)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Stopping background tasks...")
	s.bgCancel()

	log.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
