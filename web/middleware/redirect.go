package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RedirectMiddleware returns a Gin middleware that handles URL redirections.
// It normalizes legacy panel path casing to the canonical lowercase routes.
func RedirectMiddleware(basePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		redirects := map[string]string{
			"panel/API": "panel/api",
		}

		path := c.Request.URL.Path
		for from, to := range redirects {
			from, to = basePath+from, basePath+to

			if strings.HasPrefix(path, from) {
				newPath := to + path[len(from):]

				c.Redirect(http.StatusMovedPermanently, newPath)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
