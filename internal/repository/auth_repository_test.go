package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"reforest/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: false,
	})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return gormDB, mock
}

func TestAuthRepository_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()
		repo := NewAuthRepository(db)

		userID := uuid.New()
		user := &models.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: "hash",
			RoleType:     "USER",
			Profile: models.Profile{
				ID:       userID,
				FullName: "Test User",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectBegin()
		// Expect Insert into users. Using AnyArg because GORM field order can vary.
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Expect Insert into profiles (association)
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "profiles"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		created, err := repo.CreateUser(context.Background(), user)
		if err != nil {
			t.Errorf("CreateUser() error = %v", err)
		}
		if created == nil {
			t.Errorf("expected user, got nil")
		}
	})

	t.Run("duplicate key error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()
		repo := NewAuthRepository(db)

		user := &models.User{ID: uuid.New(), Email: "dup@example.com"}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(gorm.ErrDuplicatedKey)
		mock.ExpectRollback()

		_, err := repo.CreateUser(context.Background(), user)
		if err != models.ErrAlreadyExists {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("generic error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()
		repo := NewAuthRepository(db)

		user := &models.User{ID: uuid.New(), Email: "err@example.com"}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		_, err := repo.CreateUser(context.Background(), user)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestAuthRepository_GetByEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()
		repo := NewAuthRepository(db)

		email := "test@example.com"
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role_type"}).
			AddRow(userID, email, "hash", "USER")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email =`)).
			WithArgs(email, 1).
			WillReturnRows(rows)

		profileRows := sqlmock.NewRows([]string{"id", "full_name"}).
			AddRow(userID, "Test User")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "profiles" WHERE "profiles"."id" =`)).
			WithArgs(userID).
			WillReturnRows(profileRows)

		user, err := repo.GetByEmail(context.Background(), email)
		if err != nil {
			t.Errorf("GetByEmail() error = %v", err)
		}
		if user.Email != email {
			t.Errorf("expected email %s, got %s", email, user.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()
		repo := NewAuthRepository(db)

		email := "missing@example.com"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email =`)).
			WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetByEmail(context.Background(), email)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("expected ErrRecordNotFound, got %v", err)
		}
	})
}