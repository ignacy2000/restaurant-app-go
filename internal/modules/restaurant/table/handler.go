package table

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"table-service.pl/pkg/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetByRestaurant(c *gin.Context) {
	resp, err := h.svc.GetByRestaurant(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), c.Param("id"), userID, req)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) UpdateCapacity(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.UpdateCapacity(c.Request.Context(), c.Param("id"), c.Param("tableId"), userID, req)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	if err := h.svc.Delete(c.Request.Context(), c.Param("id"), c.Param("tableId"), userID); err != nil {
		h.handleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "not the restaurant owner"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
