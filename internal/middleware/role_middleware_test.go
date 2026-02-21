package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-backend/internal/middleware"
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestRequireHR_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"
	var userIDUUID pgtype.UUID
	if err := userIDUUID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}

	employeeIDStr := "02020202-0202-0202-0202-020202020202"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:           employeeIDUUID,
				EmployeeType: "HR",
			}, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireHR(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRequireHR_NotHR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("02020202-0202-0202-0202-020202020202"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:           employeeIDUUID,
				EmployeeType: "EMPLOYEE", // Not HR
			}, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireHR(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireHR_RejectsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("02020202-0202-0202-0202-020202020202"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return true, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:           employeeIDUUID,
				EmployeeType: "HR",
			}, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireHR(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireHR_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{}
	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireHR(queries)

	r := gin.New()
	// No userID set in context
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireInterviewerOrRecruiter_AllowsRecruiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"

	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("02020202-0202-0202-0202-020202020202"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: employeeIDUUID}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return employeeIDUUID, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireInterviewerOrRecruiter(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRequireInterviewerOrRecruiter_RejectsNonInterviewer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"
	var userIDUUID pgtype.UUID
	if err := userIDUUID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}

	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("02020202-0202-0202-0202-020202020202"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return pgtype.UUID{}, errors.New("not recruiter")
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID: employeeIDUUID,
			}, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireInterviewerOrRecruiter(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireInterviewerOrRecruiter_AllowsActiveInterviewer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("02020202-0202-0202-0202-020202020202"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID: employeeIDUUID,
			}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return pgtype.UUID{}, errors.New("not recruiter")
		},
		GetActiveInterviewCountFunc: func(ctx context.Context, interviewerID pgtype.UUID) (int64, error) {
			return 2, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireInterviewerOrRecruiter(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRequireInterviewerOrRecruiter_RejectsAdminWithoutEmployeeProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return true, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{}, errors.New("not found")
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireInterviewerOrRecruiter(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireInterviewerOrRecruiter_RejectsAdminEvenWithRecruiterRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "01010101-0101-0101-0101-010101010101"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("02020202-0202-0202-0202-020202020202"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return true, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: employeeIDUUID}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return employeeID, nil
		},
	}

	queries := middleware.NewQueriesAdapter(mockRepo)
	mw := middleware.RequireInterviewerOrRecruiter(queries)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
