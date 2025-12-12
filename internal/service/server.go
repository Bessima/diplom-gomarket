package service

import (
	"context"
	"github.com/Bessima/diplom-gomarket/internal/handlers"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"github.com/go-chi/chi/v5"
	"net"
	"net/http"
	"time"
)

type ServerService struct {
	Server  *http.Server
	storage repository.UserStorageRepositoryI
}

func NewServerService(rootContext context.Context, address string, storage repository.UserStorageRepositoryI) ServerService {
	server := &http.Server{
		Addr: address,
		BaseContext: func(_ net.Listener) context.Context {
			return rootContext
		},
	}
	return ServerService{Server: server, storage: storage}
}

func (serverService *ServerService) SetRouter(jwtConfig *handlers.JWTConfig) {
	var router chi.Router
	router = serverService.getRouter(jwtConfig)

	serverService.Server.Handler = router
}

func (serverService *ServerService) getRouter(jwtConfig *handlers.JWTConfig) chi.Router {
	router := chi.NewRouter()

	router.Use(logger.RequestLogger)

	authHandler := handlers.NewAuthHandler(jwtConfig, serverService.storage)
	router.Post("/api/user/register/", authHandler.RegisterHandler)
	router.Post("/api/user/login/", authHandler.LoginHandler)
	router.Post("/api/user/refresh/", authHandler.RefreshHandler)

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
