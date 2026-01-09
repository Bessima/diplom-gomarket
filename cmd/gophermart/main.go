package main

import (
	"context"
	"github.com/Bessima/diplom-gomarket/internal/config"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/handlers"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/server"
	"github.com/Bessima/diplom-gomarket/internal/service"
	"go.uber.org/zap"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	err := initLogger()
	if err != nil {
		logger.Log.Warn(err.Error())
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelCtx()

	conf := config.InitConfig()

	dbObj, errDB := db.NewDB(ctx, conf.DatabaseDNS)
	if errDB != nil {
		logger.Log.Error(
			"Unable to connect to database",
			zap.String("path", conf.DatabaseDNS),
			zap.String("error", errDB.Error()),
		)
	}
	defer dbObj.Close()

	ordersForProcessing := make(chan models.Order, 10)
	defer close(ordersForProcessing)

	orderService := service.NewOrderService(dbObj, conf.GetAccrualAddressWithProtocol())

	go orderService.AddNotProcessedOrders(ordersForProcessing)
	go orderService.GetAccrualForOrder(ordersForProcessing)

	serverService := server.NewServerService(ctx, conf.Address, dbObj)

	// Конфигурация JWT
	jwtConfig := &handlers.JWTConfig{
		SecretKey:       conf.SecretKey,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // 7 дней
	}
	serverService.SetRouter(jwtConfig, ordersForProcessing)

	serverErr := make(chan error, 1)
	logger.Log.Info("Running Server on", zap.String("address", conf.Address))
	go serverService.RunServer(&serverErr)

	// Ждем сигнал завершения или ошибку сервера
	var err error
	select {
	case <-ctx.Done():
		logger.Log.Info("Received shutdown signal, shutting down.")
	case err = <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
		return err
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
