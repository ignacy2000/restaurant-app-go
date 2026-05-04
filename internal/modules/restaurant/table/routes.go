package table

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.GET("/restaurants/:id/tables", h.GetByRestaurant)

	protected := r.Group("", auth)
	protected.POST("/restaurants/:id/tables", h.Create)
	protected.PATCH("/restaurants/:id/tables/:tableId", h.UpdateCapacity)
	protected.DELETE("/restaurants/:id/tables/:tableId", h.Delete)
}
