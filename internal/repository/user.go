package repository

import (
	"context"
	"errors"
	"github.com/Bessima/diplom-gomarket/internal/config/db"
	"github.com/Bessima/diplom-gomarket/internal/models"
	"github.com/Bessima/diplom-gomarket/internal/retry"
)

type UserRepository struct {
	db *db.DB
}

type UserStorageRepositoryI interface {
	CreateUser(username, passwordHash string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
}

func NewUserRepository(dbObj *db.DB) *UserRepository {
	return &UserRepository{db: dbObj}
}

func (repository *UserRepository) CreateUser(username, passwordHash string) (*models.User, error) {
	query := `INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id, name, password`

	return retry.DoRetryWithResult(context.Background(), func() (*models.User, error) {
		row := repository.db.Pool.QueryRow(context.Background(), query, username, passwordHash)
		if row == nil {
			return nil, errors.New("user was not created")
		}
		user := models.User{}
		err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
		return &user, err
	})
}

func (repository *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, name, password FROM users WHERE name = $1`
	return retry.DoRetryWithResult(context.Background(), func() (*models.User, error) {
		row := repository.db.Pool.QueryRow(
			context.Background(),
			query,
			username,
		)

		elem := models.User{}
		err := row.Scan(&elem.ID, &elem.Username, &elem.PasswordHash)
		return &elem, err
	})
}

func (repository *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, name, password FROM users WHERE id = $1`
	return retry.DoRetryWithResult(context.Background(), func() (*models.User, error) {
		row := repository.db.Pool.QueryRow(
			context.Background(),
			query,
			id,
		)

		elem := models.User{}
		err := row.Scan(&elem.ID, &elem.Username, &elem.PasswordHash)
		return &elem, err
	})
}
