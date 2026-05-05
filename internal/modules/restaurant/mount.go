package restaurant

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	callModule "table-service.pl/internal/modules/restaurant/call"
	menuModule "table-service.pl/internal/modules/restaurant/menu"
	orderModule "table-service.pl/internal/modules/restaurant/order"
	tableModule "table-service.pl/internal/modules/restaurant/table"
	internalWS "table-service.pl/internal/ws"
	"table-service.pl/pkg/mailer"
	"table-service.pl/pkg/ws"
)

// MountAll wires and registers all restaurant sub-modules
// (restaurant, menu, table, order, call, websocket) under the given api group.
func MountAll(api *gin.RouterGroup, authMw gin.HandlerFunc, db *sql.DB, hub *ws.Hub, jwtSecret string, mail *mailer.Mailer, frontendURL string) {
	restaurantRepo := NewRepository(db)
	menuRepo := menuModule.NewRepository(db)
	tableRepo := tableModule.NewRepository(db)
	orderRepo := orderModule.NewRepository(db)
	callRepo := callModule.NewRepository(db)

	restaurantSvc := NewService(restaurantRepo)
	menuSvc := menuModule.NewService(menuRepo, restaurantRepo)
	tableSvc := tableModule.NewService(tableRepo, restaurantRepo)
	orderSvc := orderModule.NewService(orderRepo, tableRepo, hub, restaurantRepo, mail, frontendURL)
	callSvc := callModule.NewService(callRepo, tableRepo, hub)

	Mount(api, authMw, NewHandler(restaurantSvc))
	menuModule.Mount(api, authMw, menuModule.NewHandler(menuSvc))
	tableModule.Mount(api, authMw, tableModule.NewHandler(tableSvc))
	orderModule.Mount(api, authMw, orderModule.NewHandler(orderSvc))
	callModule.Mount(api, authMw, callModule.NewHandler(callSvc))

	internalWS.Mount(api, internalWS.NewHandler(hub, jwtSecret, restaurantRepo, tableRepo))
}
