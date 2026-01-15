package db

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
)

type DB struct {
	Pool PgxPoolInterface
}

func NewDB(ctx context.Context, dns string) (*DB, error) {
	dbPool, err := pgxpool.New(ctx, dns)
	if err != nil {
		return nil, err
	}

	if err := dbPool.Ping(ctx); err != nil {
		return &DB{Pool: dbPool}, err
	}
	obj := DB{Pool: dbPool}

	err = obj.runMigrations()
	if err != nil {
		return &DB{Pool: dbPool}, err
	}

	return &obj, nil
}

func (db *DB) runMigrations() error {
	// Получаем конфиг из пула
	config := db.Pool.Config()

	// Создаем стандартное sql.DB подключение через pgx
	sqlDB := stdlib.OpenDB(*config.ConnConfig)
	defer sqlDB.Close()

	// Создаем драйвер для миграций
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
