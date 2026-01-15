package repository

import "github.com/Bessima/diplom-gomarket/internal/config/db"

func NewTestDB(pool db.PgxPoolInterface) *db.DB {
	return &db.DB{
		Pool: pool,
	}
}
