package menu

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.GET("/restaurants/:id/menus", h.GetByRestaurant)
	r.GET("/restaurants/:id/menu-items", h.GetItemsByRestaurant)

	protected := r.Group("", auth)
	protected.POST("/restaurants/:id/menus", h.Create)
	protected.POST("/menus/:menuId/items", h.CreateItem)
	protected.DELETE("/menu-items/:itemId", h.DeleteItem)
}
