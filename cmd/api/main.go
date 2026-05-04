package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"table-service.pl/internal/database"
	authModule "table-service.pl/internal/modules/auth"
	restaurantModule "table-service.pl/internal/modules/restaurant"
	userModule "table-service.pl/internal/modules/user"
	"table-service.pl/pkg/config"
	"table-service.pl/pkg/logger"
	"table-service.pl/pkg/middleware"
	"table-service.pl/pkg/ws"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Error("connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.MigrateUp(db); err != nil {
		log.Error("run migrations", "error", err)
		os.Exit(1)
	}

	hub := ws.NewHub()
	go hub.Run()

	userRepo := userModule.NewRepository(db)
	authRepo := authModule.NewRepository(db)

	userSvc := userModule.NewService(userRepo)
	authSvc := authModule.NewService(authRepo, userRepo, cfg.JWTSecret)

	userHandler := userModule.NewHandler(userSvc)
	authHandler := authModule.NewHandler(authSvc)

	authMw := middleware.Auth(cfg.JWTSecret)
	originStore := middleware.NewOriginStore(time.Minute, authRepo.AllOriginValues, cfg.FrontendURL)

	r := gin.New()
	r.Use(middleware.Recovery(log))
	r.Use(middleware.DynamicCORS(originStore))

	api := r.Group("/api")
	userModule.Mount(api, authMw, userHandler)
	authModule.Mount(api, authMw, authHandler)
	restaurantModule.MountAll(api, authMw, db, hub, cfg.JWTSecret)

	log.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
