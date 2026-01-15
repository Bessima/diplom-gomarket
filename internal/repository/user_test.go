package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateUser_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	username := "testuser"
	passwordHash := "$2a$10$hashedpassword"
	userID := 1

	rows := pgxmock.NewRows([]string{"id", "name", "password"}).
		AddRow(userID, username, passwordHash)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(username, passwordHash).
		WillReturnRows(rows)

	// Act
	user, err := repo.CreateUser(username, passwordHash)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, username, user.Login)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CreateUser_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	username := "testuser"
	passwordHash := "$2a$10$hashedpassword"
	expectedError := errors.New("unique violation")

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(username, passwordHash).
		WillReturnError(expectedError)

	// Act
	user, err := repo.CreateUser(username, passwordHash)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user) // Функция всегда возвращает указатель, даже при ошибке
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CreateUser_ScanError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	username := "testuser"
	passwordHash := "$2a$10$hashedpassword"

	rows := pgxmock.NewRows([]string{"id", "name", "password"}).
		AddRow("invalid", username, passwordHash).
		RowError(0, errors.New("scan error"))

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(username, passwordHash).
		WillReturnRows(rows)

	// Act
	user, err := repo.CreateUser(username, passwordHash)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByLogin_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	username := "testuser"
	passwordHash := "$2a$10$hashedpassword"
	userID := 1

	rows := pgxmock.NewRows([]string{"id", "name", "password"}).
		AddRow(userID, username, passwordHash)

	mock.ExpectQuery("SELECT id, name, password FROM users WHERE name").
		WithArgs(username).
		WillReturnRows(rows)

	// Act
	user, err := repo.GetUserByLogin(username)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, username, user.Login)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByLogin_NotFound(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	username := "nonexistent"

	mock.ExpectQuery("SELECT id, name, password FROM users WHERE name").
		WithArgs(username).
		WillReturnError(pgx.ErrNoRows)

	// Act
	user, err := repo.GetUserByLogin(username)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByLogin_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	username := "testuser"
	expectedError := errors.New("database connection error")

	mock.ExpectQuery("SELECT id, name, password FROM users WHERE name").
		WithArgs(username).
		WillReturnError(expectedError)

	// Act
	user, err := repo.GetUserByLogin(username)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_Success(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	userID := 1
	username := "testuser"
	passwordHash := "$2a$10$hashedpassword"

	rows := pgxmock.NewRows([]string{"id", "name", "password"}).
		AddRow(userID, username, passwordHash)

	mock.ExpectQuery("SELECT id, name, password FROM users WHERE id").
		WithArgs(userID).
		WillReturnRows(rows)

	// Act
	user, err := repo.GetUserByID(userID)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, username, user.Login)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	userID := 999

	mock.ExpectQuery("SELECT id, name, password FROM users WHERE id").
		WithArgs(userID).
		WillReturnError(pgx.ErrNoRows)

	// Act
	user, err := repo.GetUserByID(userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_DatabaseError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	dbObj := NewTestDB(mock)
	repo := NewUserRepository(dbObj)

	userID := 1
	expectedError := errors.New("database connection error")

	mock.ExpectQuery("SELECT id, name, password FROM users WHERE id").
		WithArgs(userID).
		WillReturnError(expectedError)

	// Act
	user, err := repo.GetUserByID(userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CreateUser_DifferentUsers(t *testing.T) {
	testCases := []struct {
		name         string
		username     string
		passwordHash string
		userID       int
	}{
		{
			name:         "user1",
			username:     "alice",
			passwordHash: "$2a$10$hash1",
			userID:       1,
		},
		{
			name:         "user2",
			username:     "bob",
			passwordHash: "$2a$10$hash2",
			userID:       2,
		},
		{
			name:         "user with special chars",
			username:     "user@example.com",
			passwordHash: "$2a$10$hash3",
			userID:       3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			dbObj := NewTestDB(mock)
			repo := NewUserRepository(dbObj)

			rows := pgxmock.NewRows([]string{"id", "name", "password"}).
				AddRow(tc.userID, tc.username, tc.passwordHash)

			mock.ExpectQuery("INSERT INTO users").
				WithArgs(tc.username, tc.passwordHash).
				WillReturnRows(rows)

			// Act
			user, err := repo.CreateUser(tc.username, tc.passwordHash)

			// Assert
			assert.NoError(t, err)
			require.NotNil(t, user)
			assert.Equal(t, tc.userID, user.ID)
			assert.Equal(t, tc.username, user.Login)
			assert.Equal(t, tc.passwordHash, user.PasswordHash)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
