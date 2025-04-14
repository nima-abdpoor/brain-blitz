package repository

import (
	"BrainBlitz.com/game/services/user_app/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

type MockLogger struct{}

func (m MockLogger) Info(msg string, args ...any)  {}
func (m MockLogger) Warn(msg string, args ...any)  {}
func (m MockLogger) Debug(msg string, args ...any) {}
func (m MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// no op
}

func TestInsertUser(t *testing.T) {
	ctx := context.Background()

	user := service.User{
		ID:             12,
		Username:       "nima",
		HashedPassword: "PASSWORD",
		DisplayName:    "nima",
		CreatedAt:      uint64(time.Now().UnixMilli()),
		UpdatedAt:      uint64(time.Now().UnixMilli()),
		Role:           service.UserRole,
	}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("INSERT INTO users").
			ExpectQuery().
			WithArgs(user.Username, user.HashedPassword, user.DisplayName, user.Role.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := repo.InsertUser(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate key error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("INSERT INTO users").
			ExpectQuery().
			WithArgs(user.Username, user.HashedPassword, user.DisplayName, user.Role.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint"))

		id, err := repo.InsertUser(ctx, user)

		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "duplicate username"))
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("generic insert error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("INSERT INTO users").
			ExpectQuery().
			WithArgs(user.Username, user.HashedPassword, user.DisplayName, user.Role.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("some db error"))

		id, err := repo.InsertUser(ctx, user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert user")
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("prepare statement error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("INSERT INTO users").
			WillReturnError(errors.New("prepare failed"))

		id, err := repo.InsertUser(ctx, user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare statement")
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	username := "nima"
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE username = \\$1 LIMIT 1").
			ExpectQuery().
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username", "password", "display_name", "role", "created_at", "updated_at",
			}).AddRow(
				1, "nima", "hashed_pass", "Nima Abdpour", "user", now, now,
			))

		user, err := repo.GetUser(ctx, username)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "nima", user.Username)
		assert.Equal(t, "hashed_pass", user.HashedPassword)
		assert.Equal(t, "Nima Abdpour", user.DisplayName)
		assert.Equal(t, service.UserRoleStr, user.Role.String())
		assert.Equal(t, uint64(now.UTC().UnixMilli()), user.CreatedAt)
		assert.Equal(t, uint64(now.UTC().UnixMilli()), user.UpdatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE username = \\$1 LIMIT 1").
			ExpectQuery().
			WithArgs(username).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUser(ctx, username)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "result with username")
		assert.Empty(t, user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE username = \\$1 LIMIT 1").
			ExpectQuery().
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username",
			}).AddRow(
				1, "johndoe",
			))

		_, err = repo.GetUser(ctx, username)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error retrieving user with username")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("prepare statement error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE username = \\$1 LIMIT 1").
			WillReturnError(fmt.Errorf("prepare error"))

		_, err = repo.GetUser(ctx, username)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare find result by ID statement")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetUserById(t *testing.T) {
	ctx := context.Background()
	userID := "1"
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE id = \\$1 LIMIT 1").
			ExpectQuery().
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username", "password", "display_name", "role", "created_at", "updated_at",
			}).AddRow(
				1, "nima", "hashed_pass", "Nima Abdpour", "user", now, now,
			))

		user, err := repo.GetUserById(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "nima", user.Username)
		assert.Equal(t, "hashed_pass", user.HashedPassword)
		assert.Equal(t, "Nima Abdpour", user.DisplayName)
		assert.Equal(t, service.UserRoleStr, user.Role.String())
		assert.Equal(t, uint64(now.UTC().UnixMilli()), user.CreatedAt)
		assert.Equal(t, uint64(now.UTC().UnixMilli()), user.UpdatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE id = \\$1 LIMIT 1").
			ExpectQuery().
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserById(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "result with id")
		assert.Empty(t, user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE id = \\$1 LIMIT 1").
			ExpectQuery().
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username", // Missing columns to trigger scan error
			}).AddRow(
				1, "nima",
			))

		_, err = repo.GetUserById(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error retrieving user with id")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("prepare statement error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		repo := NewUserRepository(Config{}, db, MockLogger{})

		mock.ExpectPrepare("SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE id = \\$1 LIMIT 1").
			WillReturnError(fmt.Errorf("prepare failed"))

		_, err = repo.GetUserById(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare find result by ID")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
