package order

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.POST("/restaurants/:id/tables/:tableId/orders", h.Create)

	protected := r.Group("", auth)
	protected.GET("/restaurants/:id/orders", h.GetByRestaurant)
	protected.PATCH("/restaurants/:id/orders/:orderId/status", h.UpdateStatus)
}
