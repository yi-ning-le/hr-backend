package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"hr-backend/internal/handler"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAssignReviewerHandler_realUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cIDStr := "00000000-0000-0000-0000-000000000001"
	rIDStr := "00000000-0000-0000-0000-000000000002"

	mockRepo := &mocks.MockQuerier{
		AssignReviewerFunc: func(ctx context.Context, arg repository.AssignReviewerParams) (repository.AssignReviewerRow, error) {
			// verification logic
			return repository.AssignReviewerRow{
				ID:              pgtype.UUID{Bytes: [16]byte{15: 1}, Valid: true}, // approximate match
				ReviewerID:      pgtype.UUID{Bytes: [16]byte{15: 2}, Valid: true},
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AppliedJobTitle: "Software Engineer",
				Name:            "Test Candidate",
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         pgtype.UUID{Bytes: [16]byte{15: 1}, Valid: true},
				ReviewerID: pgtype.UUID{Bytes: [16]byte{15: 2}, Valid: true},
				AppliedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", "00000000-0000-0000-0000-00000000000a")
		c.Next()
	})
	r.POST("/candidates/:id/assign-reviewer", h.AssignReviewer)

	reqBody := map[string]string{
		"reviewerId": rIDStr,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/candidates/"+cIDStr+"/assign-reviewer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestSubmitReviewHandler_InvalidReviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	submitCalled := false
	mockRepo := &mocks.MockQuerier{
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "unexpected",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/00000000-0000-0000-0000-000000000001/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if submitCalled {
		t.Error("expected SubmitReview not to be called for invalid reviewStatus")
	}
}

func TestSubmitReviewHandler_NormalizesReviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000003"
	var candidateID pgtype.UUID
	if err := candidateID.Scan(candidateIDStr); err != nil {
		t.Fatalf("failed to scan candidate id: %v", err)
	}
	var reviewerEmployeeID pgtype.UUID
	if err := reviewerEmployeeID.Scan("00000000-0000-0000-0000-000000000002"); err != nil {
		t.Fatalf("failed to scan reviewer employee id: %v", err)
	}

	submitCalled := false
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: reviewerEmployeeID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         candidateID,
				ReviewerID: reviewerEmployeeID,
			}, nil
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			if arg.ReviewStatus.String != "suitable" {
				t.Errorf("expected normalized status suitable, got %s", arg.ReviewStatus.String)
			}
			return repository.SubmitReviewRow{
				ID: candidateID,

				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AppliedJobTitle: "Software Engineer",
				Name:            "Test Candidate",
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "  SUITABLE ",
		"reviewNote":   "strong fit",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/"+candidateIDStr+"/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !submitCalled {
		t.Error("expected SubmitReview to be called")
	}
}

func TestSubmitReviewHandler_ReturnsForbiddenWhenNoEmployeeProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000003"
	submitCalled := false

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{}, pgx.ErrNoRows
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "suitable",
		"reviewNote":   "strong fit",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/"+candidateIDStr+"/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
	if submitCalled {
		t.Error("expected SubmitReview not to be called")
	}
}

func TestSubmitReviewHandler_ReturnsNotFoundWhenCandidateMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000003"
	submitCalled := false

	var reviewerEmployeeID pgtype.UUID
	if err := reviewerEmployeeID.Scan("00000000-0000-0000-0000-000000000002"); err != nil {
		t.Fatalf("failed to scan reviewer employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: reviewerEmployeeID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{}, pgx.ErrNoRows
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "suitable",
		"reviewNote":   "strong fit",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/"+candidateIDStr+"/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
	if submitCalled {
		t.Error("expected SubmitReview not to be called")
	}
}

func TestRevertReviewerHandler_InvalidCandidateID_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewCandidateService(&mocks.MockQuerier{})
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/revert-reviewer", h.RevertReviewer)

	req, _ := http.NewRequest("POST", "/candidates/invalid-id/revert-reviewer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRevertReviewerHandler_NoReviewerToRevert_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, candidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{}, pgx.ErrNoRows
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/revert-reviewer", h.RevertReviewer)

	req, _ := http.NewRequest("POST", "/candidates/00000000-0000-0000-0000-000000000001/revert-reviewer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRevertReviewerHandler_ReviewAlreadySubmitted_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, candidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{ReviewStatus: "suitable"}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/revert-reviewer", h.RevertReviewer)

	req, _ := http.NewRequest("POST", "/candidates/00000000-0000-0000-0000-000000000001/revert-reviewer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRevertReviewerHandler_UnexpectedError_ReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, candidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{}, errors.New("database unavailable")
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/revert-reviewer", h.RevertReviewer)

	req, _ := http.NewRequest("POST", "/candidates/00000000-0000-0000-0000-000000000001/revert-reviewer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRevertReviewerHandler_Success_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateID := mustScanCandidateHandlerUUID(t, "00000000-0000-0000-0000-000000000001")
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{ReviewStatus: "pending"}, nil
		},
		RemoveCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (int64, error) {
			return 1, nil
		},
		ClearCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) error {
			return nil
		},
		GetCandidateFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:              candidateID,
				Name:            "John Doe",
				AppliedJobTitle: "Software Engineer",
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/revert-reviewer", h.RevertReviewer)

	req, _ := http.NewRequest("POST", "/candidates/00000000-0000-0000-0000-000000000001/revert-reviewer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func mustScanCandidateHandlerUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("failed to scan uuid %s: %v", raw, err)
	}
	return id
}

func TestCreateCandidateHandler_MultipartSuccess(t *testing.T) {
	// Don't pollute the handler directory with an uploads folder during the test
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CreateCandidateFunc: func(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error) {
			return repository.Candidate{
				ID: mustScanCandidateHandlerUUID(t, "00000000-0000-0000-0000-000000000001"),
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID: id,
			}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates", h.CreateCandidate)

	// Create multipart form data
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add file
	fileWriter, err := writer.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}

	// Add data
	data := map[string]interface{}{
		"name":            "John Doe",
		"email":           "johndoe@example.com",
		"phone":           "1234567890",
		"experienceYears": 3,
		"education":       "BSc Computer Science",
		"appliedJobId":    "00000000-0000-0000-0000-000000000001",
		"channel":         "LinkedIn",
		"appliedAt":       time.Now().Format(time.RFC3339),
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	if err := writer.WriteField("data", string(dataBytes)); err != nil {
		t.Fatalf("failed to write data field: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", "/candidates", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// We expect 201 Created
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCreateCandidateHandler_MultipartValidationError(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CreateCandidateFunc: func(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error) {
			t.Fatalf("CreateCandidate should not be called when validation fails")
			return repository.Candidate{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates", h.CreateCandidate)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	fileWriter, err := writer.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}

	// Invalid email should fail model validation.
	data := map[string]interface{}{
		"name":            "John Doe",
		"email":           "invalid-email",
		"phone":           "1234567890",
		"experienceYears": 3,
		"education":       "BSc Computer Science",
		"appliedJobId":    "00000000-0000-0000-0000-000000000001",
		"channel":         "LinkedIn",
		"appliedAt":       time.Now().Format(time.RFC3339),
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	if err := writer.WriteField("data", string(dataBytes)); err != nil {
		t.Fatalf("failed to write data field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", "/candidates", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest, got %d body=%s", w.Code, w.Body.String())
	}

	files, err := os.ReadDir("uploads")
	if err != nil {
		t.Fatalf("failed to read uploads directory: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected uploaded file to be cleaned up, found %d files", len(files))
	}
}

func TestCreateCandidateHandler_RejectsNonPDFResume(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CreateCandidateFunc: func(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error) {
			t.Fatalf("CreateCandidate should not be called for unsupported file types")
			return repository.Candidate{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates", h.CreateCandidate)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	fileWriter, err := writer.CreateFormFile("file", "resume.docx")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy docx content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}

	data := map[string]interface{}{
		"name":            "John Doe",
		"email":           "john@example.com",
		"phone":           "1234567890",
		"experienceYears": 3,
		"education":       "BSc Computer Science",
		"appliedJobId":    "00000000-0000-0000-0000-000000000001",
		"channel":         "LinkedIn",
		"appliedAt":       time.Now().Format(time.RFC3339),
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	if err := writer.WriteField("data", string(dataBytes)); err != nil {
		t.Fatalf("failed to write data field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", "/candidates", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateResumeHandler_Success(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	candidateID := mustScanCandidateHandlerUUID(t, "00000000-0000-0000-0000-000000000001")
	if err := os.MkdirAll("./uploads", 0o755); err != nil {
		t.Fatalf("failed to create uploads dir: %v", err)
	}
	if err := os.WriteFile("./uploads/old-resume.pdf", []byte("old"), 0o600); err != nil {
		t.Fatalf("failed to create old resume file: %v", err)
	}
	getCandidateCalls := 0
	mockRepo := &mocks.MockQuerier{
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			return repository.Candidate{ID: arg.ID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			getCandidateCalls++
			resumeURL := "/static/resumes/old-resume.pdf"
			if getCandidateCalls > 1 {
				resumeURL = "/static/resumes/new-resume.pdf"
			}
			return repository.GetCandidateRow{
				ID:              candidateID,
				Name:            "John Doe",
				ResumeUrl:       resumeURL,
				AppliedJobTitle: "Software Engineer",
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	fileWriter, err := writer.CreateFormFile("file", "new_resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("PATCH", "/candidates/00000000-0000-0000-0000-000000000001/resume", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	if _, err := os.Stat("./uploads/old-resume.pdf"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected old resume file to be removed, err=%v", err)
	}

	files, err := os.ReadDir("./uploads")
	if err != nil {
		t.Fatalf("failed to read uploads dir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected exactly one resume file after replacement, got %d", len(files))
	}
}

func TestUpdateResumeHandler_MissingFile_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			t.Fatal("UpdateCandidateResume should not be called when file is missing")
			return repository.Candidate{}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	req, _ := http.NewRequest("PATCH", "/candidates/00000000-0000-0000-0000-000000000001/resume", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateResumeHandler_NonPDF_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			t.Fatal("UpdateCandidateResume should not be called for non-PDF files")
			return repository.Candidate{}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fileWriter, err := writer.CreateFormFile("file", "resume.docx")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy docx content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("PATCH", "/candidates/00000000-0000-0000-0000-000000000001/resume", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUpdateResumeHandler_ServiceError_ReturnsInternalServerError(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:        id,
				Name:      "John Doe",
				ResumeUrl: "/static/resumes/old-resume.pdf",
			}, nil
		},
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			return repository.Candidate{}, errors.New("database error")
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fileWriter, err := writer.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("PATCH", "/candidates/00000000-0000-0000-0000-000000000001/resume", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d body=%s", w.Code, w.Body.String())
	}

	files, err := os.ReadDir("./uploads")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed to read uploads dir: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected uploaded file rollback on failure, got %d files", len(files))
	}
}

func TestUpdateResumeHandler_CandidateNotFound_ReturnsNotFound(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{}, pgx.ErrNoRows
		},
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			return repository.Candidate{}, pgx.ErrNoRows
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fileWriter, err := writer.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("PATCH", "/candidates/00000000-0000-0000-0000-000000000001/resume", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}

	files, err := os.ReadDir("./uploads")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed to read uploads dir: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected uploaded file rollback when candidate missing, got %d files", len(files))
	}
}

func TestUpdateResumeHandler_InvalidCandidateID_ReturnsBadRequestAndRollsBackFile(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			t.Fatal("GetCandidate should not be called for invalid candidate id")
			return repository.GetCandidateRow{}, nil
		},
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			t.Fatal("UpdateCandidateResume should not be called for invalid candidate id")
			return repository.Candidate{}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fileWriter, err := writer.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("PATCH", "/candidates/not-a-uuid/resume", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}

	files, err := os.ReadDir("./uploads")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed to read uploads dir: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected uploaded file rollback on invalid candidate id, got %d files", len(files))
	}
}

func TestUpdateResumeHandler_ExternalOldURL_DoesNotDeleteLocalFiles(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	gin.SetMode(gin.TestMode)

	if err := os.MkdirAll("./uploads", 0o755); err != nil {
		t.Fatalf("failed to create uploads dir: %v", err)
	}
	localKeepFile := filepath.Join("./uploads", "keep.pdf")
	if err := os.WriteFile(localKeepFile, []byte("keep"), 0o600); err != nil {
		t.Fatalf("failed to write keep file: %v", err)
	}

	candidateID := mustScanCandidateHandlerUUID(t, "00000000-0000-0000-0000-000000000001")
	getCandidateCalls := 0
	mockRepo := &mocks.MockQuerier{
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			return repository.Candidate{ID: arg.ID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			getCandidateCalls++
			resumeURL := "https://cdn.example.com/resume.pdf"
			if getCandidateCalls > 1 {
				resumeURL = "/static/resumes/new-resume.pdf"
			}
			return repository.GetCandidateRow{
				ID:              candidateID,
				Name:            "John Doe",
				ResumeUrl:       resumeURL,
				AppliedJobTitle: "Software Engineer",
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.PATCH("/candidates/:id/resume", h.UpdateResume)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fileWriter, err := writer.CreateFormFile("file", "new_resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("dummy pdf content")); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("PATCH", "/candidates/00000000-0000-0000-0000-000000000001/resume", &b)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	if _, err := os.Stat(localKeepFile); err != nil {
		t.Fatalf("expected local keep file to stay untouched, err=%v", err)
	}
}
