package main

import (
	"context"
	"github.com/Bessima/diplom-gomarket/internal/config"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"github.com/Bessima/diplom-gomarket/internal/service"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

func main() {
	err := initLogger()
	if err != nil {
		logger.Log.Warn(err.Error())
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf := config.InitConfig()
	storage := repository.NewDBRepository(rootCtx, conf.DatabaseDNS)

	serverService := service.NewServerService(rootCtx, conf.Address, storage)

	serverErr := make(chan error, 1)
	logger.Log.Info("Running Server on", zap.String("address", conf.Address))
	go serverService.RunServer(&serverErr)

	// Ждем сигнал завершения или ошибку сервера
	var err error
	select {
	case <-rootCtx.Done():
		logger.Log.Info("Received shutdown signal, shutting down.")
	case err = <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
	}

	if shutdownErr := serverService.Shutdown(); shutdownErr != nil {
		logger.Log.Error("Server shutdown error", zap.Error(shutdownErr))
	}

	logger.Log.Info("Received shutdown signal, shutting down.")

	return err
}

func initLogger() error {
	if err := logger.Initialize("debug"); err != nil {
		return err
	}
	return nil
}
