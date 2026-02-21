package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPagination parses limit and page from query parameters.
// It returns limit and offset calculated from page.
// Default limit is used if limit query param is missing or invalid.
// Default page is used if page query param is missing or invalid.
func GetPagination(c *gin.Context, defaultLimit, defaultPage int) (limit, offset int) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	l, err := strconv.Atoi(limitStr)
	if err != nil || l <= 0 {
		l = defaultLimit
	}

	p, err := strconv.Atoi(pageStr)
	if err != nil || p <= 0 {
		p = defaultPage
	}

	return l, (p - 1) * l
}

// ParseLimitOffset parses limit and offset directly from query parameters.
// It returns limit and offset as int32.
// Default limit and offset are used if query params are missing or invalid.
func ParseLimitOffset(c *gin.Context, defaultLimit, defaultOffset int) (limit, offset int32) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	l, err := strconv.Atoi(limitStr)
	if err != nil || l <= 0 {
		l = defaultLimit
	}

	o, err := strconv.Atoi(offsetStr)
	if err != nil || o < 0 {
		o = defaultOffset
	}

	return int32(l), int32(o)
}
