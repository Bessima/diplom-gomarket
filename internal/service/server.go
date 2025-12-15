package service

import (
	"context"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/handlers"
	middleware "github.com/Bessima/diplom-gomarket/internal/middlewares"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"github.com/go-chi/chi/v5"
	"net"
	"net/http"
	"time"
)

type ServerService struct {
	Server *http.Server
	db     *db.DB
}

func NewServerService(rootContext context.Context, address string, db *db.DB) ServerService {
	server := &http.Server{
		Addr: address,
		BaseContext: func(_ net.Listener) context.Context {
			return rootContext
		},
	}
	return ServerService{Server: server, db: db}
}

func (serverService *ServerService) SetRouter(jwtConfig *handlers.JWTConfig) {
	var router chi.Router
	router = serverService.getRouter(jwtConfig)

	serverService.Server.Handler = router
}

func (serverService *ServerService) getRouter(jwtConfig *handlers.JWTConfig) chi.Router {
	router := chi.NewRouter()

	router.Use(logger.RequestLogger)
	//router.Use(compress.GZIPMiddleware)

	userRepository := repository.NewUserRepository(serverService.db)
	orderRepository := repository.NewOrderRepository(serverService.db)

	authHandler := handlers.NewAuthHandler(jwtConfig, userRepository)
	router.Post("/api/user/register/", authHandler.RegisterHandler)
	router.Post("/api/user/login/", authHandler.LoginHandler)
	router.Post("/api/user/refresh/", authHandler.RefreshHandler)

	orderHandler := handlers.NewOrderHandler(orderRepository)
	router.With(middleware.AuthMiddleware(authHandler)).Post("/api/user/logout/", authHandler.LogoutHandler)
	router.With(middleware.AuthMiddleware(authHandler)).Post("/api/user/orders/", orderHandler.Add)

	return router
}

func (serverService *ServerService) RunServer(serverErr *chan error) {
	if err := serverService.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		*serverErr <- err
	} else {
		*serverErr <- nil
	}
}

func (serverService *ServerService) Shutdown() error {
	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if shutdownErr := serverService.Server.Shutdown(shutdownCtx); shutdownErr != nil {
		return shutdownErr
	}

	return nil
}
