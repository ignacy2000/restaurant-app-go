package user

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, _ gin.HandlerFunc, h *Handler) {
	r.POST("/users", h.Create)
}
