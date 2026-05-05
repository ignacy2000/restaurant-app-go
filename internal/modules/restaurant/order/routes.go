package order

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.POST("/restaurants/:id/tables/:tableId/orders", h.Create)
	r.GET("/restaurants/:id/tables/:tableId/orders", h.GetByTable)
	r.GET("/orders/confirm", h.Confirm)

	protected := r.Group("", auth)
	protected.GET("/restaurants/:id/orders", h.GetByRestaurant)
	protected.GET("/restaurants/:id/orders/active", h.GetActiveByRestaurant)
	protected.PATCH("/restaurants/:id/orders/:orderId/status", h.UpdateStatus)
}
