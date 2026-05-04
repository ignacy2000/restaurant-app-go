package restaurant

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.GET("/restaurants/:id", h.GetByID)

	protected := r.Group("", auth)
	protected.POST("/restaurants", h.Create)
	protected.GET("/restaurants/my", h.GetMy)
}
