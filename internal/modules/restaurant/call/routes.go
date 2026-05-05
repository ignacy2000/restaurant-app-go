package call

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.POST("/restaurants/:id/tables/:tableId/calls", h.Create)

	protected := r.Group("", auth)
	protected.GET("/restaurants/:id/calls", h.GetByRestaurant)
	protected.GET("/restaurants/:id/calls/active", h.GetActiveByRestaurant)
	protected.PATCH("/restaurants/:id/calls/:callId/status", h.UpdateStatus)
}
