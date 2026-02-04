package handler

import (
	"net/http"
	"strconv"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	service *service.EmployeeService
}

func NewEmployeeHandler(s *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: s}
}

// ListEmployees godoc
// GET /employees
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	status := c.Query("status")
	department := c.Query("department")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.service.ListEmployees(c.Request.Context(), status, department, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetEmployee godoc
// GET /employees/:id
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	employee, err := h.service.GetEmployee(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}
	c.JSON(http.StatusOK, employee)
}

// CreateEmployee godoc
// POST /employees
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var input model.EmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employee, err := h.service.CreateEmployee(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, employee)
}

// UpdateEmployee godoc
// PUT /employees/:id
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var input model.EmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employee, err := h.service.UpdateEmployee(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, employee)
}

// DeleteEmployee godoc
// DELETE /employees/:id
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteEmployee(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
