// handlers/render.go
package handlers

import (
	"go-pubsub-api/internal/views"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

func Render(c *gin.Context, status int, component templ.Component) {
	c.Status(status)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.AbortWithError(500, err)
	}
}

func DashboardHandler(c *gin.Context) {
	Render(c, http.StatusOK, views.Dashboard())
}
