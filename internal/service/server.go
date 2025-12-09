package service

import (
	"context"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"github.com/go-chi/chi/v5"
	"net"
	"net/http"
	"time"
)

type ServerService struct {
	Server  *http.Server
	storage repository.StorageRepositoryI
}

func NewServerService(rootContext context.Context, address string, storage repository.StorageRepositoryI) ServerService {
	server := &http.Server{
		Addr: address,
		BaseContext: func(_ net.Listener) context.Context {
			return rootContext
		},
	}
	return ServerService{Server: server, storage: storage}
}

func (serverService *ServerService) SetRouter() {
	var router chi.Router

	serverService.Server.Handler = router
}

func (serverService *ServerService) getRouter() chi.Router {
	router := chi.NewRouter()

	router.Use(logger.RequestLogger)

	//router.Get("/", handler.MainHandler(serverService.storage, templates))

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
