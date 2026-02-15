package middleware

import "github.com/gin-gonic/gin"

// ImmutableCache sets the Cache-Control header to aggressive caching for immutable files.
// Use this for files that are fingerprinted or unique by UUID (like uploaded resumes).
func ImmutableCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	}
}
