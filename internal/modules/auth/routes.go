package auth

import "github.com/gin-gonic/gin"

func Mount(r *gin.RouterGroup, auth gin.HandlerFunc, h *Handler) {
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
	r.POST("/auth/refresh", h.Refresh)
	r.POST("/auth/forgot-password", h.ForgotPassword)
	r.POST("/auth/reset-password", h.ResetPassword)

	protected := r.Group("", auth)
	protected.GET("/origins", h.ListOrigins)
	protected.POST("/origins", h.AddOrigin)
	protected.DELETE("/origins/:id", h.RemoveOrigin)
}
