package repository

import (
	"context"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"go.uber.org/zap"
)

type DBRepository struct {
	db *db.DB
}

func NewDBRepository(rootContext context.Context, databaseDNS string) *DBRepository {
	dbObj, errDB := db.NewDB(rootContext, databaseDNS)

	if errDB != nil {

		logger.Log.Error(
			"Unable to connect to database",
			zap.String("path", databaseDNS),
			zap.String("error", errDB.Error()),
		)
	}

	return &DBRepository{db: dbObj}
}

func (repository *DBRepository) Close() error {
	repository.db.Close()
	return nil
}
