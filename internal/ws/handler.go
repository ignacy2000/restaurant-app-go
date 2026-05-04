package ws

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	pkgws "table-service.pl/pkg/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type restaurantRepo interface {
	FindOwnerID(ctx context.Context, id string) (string, error)
}

type tableRepo interface {
	FindRestaurantID(ctx context.Context, id string) (string, error)
}

type Handler struct {
	hub            *pkgws.Hub
	jwtSecret      string
	restaurantRepo restaurantRepo
	tableRepo      tableRepo
}

func NewHandler(hub *pkgws.Hub, jwtSecret string, restaurantRepo restaurantRepo, tableRepo tableRepo) *Handler {
	return &Handler{
		hub:            hub,
		jwtSecret:      jwtSecret,
		restaurantRepo: restaurantRepo,
		tableRepo:      tableRepo,
	}
}

// OwnerConnect — właściciel restauracji łączy się przez WS.
// Token JWT przekazywany jako query param: ?token=<jwt>
// (przeglądarki nie obsługują custom headerów w WS API)
func (h *Handler) OwnerConnect(c *gin.Context) {
	restaurantID := c.Param("id")

	userID, err := h.parseToken(c.Query("token"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	ownerID, err := h.restaurantRepo.FindOwnerID(c.Request.Context(), restaurantID)
	if err != nil || ownerID == "" || ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not the restaurant owner"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.ServeOwner(conn, restaurantID)
}

// TableConnect — klient przy stoliku łączy się przez WS (bez autoryzacji).
func (h *Handler) TableConnect(c *gin.Context) {
	restaurantID := c.Param("id")
	tableID := c.Param("tableId")

	tableRestaurantID, err := h.tableRepo.FindRestaurantID(c.Request.Context(), tableID)
	if err != nil || tableRestaurantID == "" || tableRestaurantID != restaurantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.ServeTable(conn, tableID)
}

func Mount(r *gin.RouterGroup, h *Handler) {
	r.GET("/ws/restaurants/:id", h.OwnerConnect)
	r.GET("/ws/restaurants/:id/tables/:tableId", h.TableConnect)
}

func (h *Handler) parseToken(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", errors.New("missing token")
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("missing sub")
	}
	return sub, nil
}
