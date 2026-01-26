package repository

import (
	"context"
	"errors"
	"reforest/internal/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
    CreateUserWithRole(ctx context.Context, user *models.User) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type authRepository struct {
    db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
    return &authRepository{db: db}
}

func (r *authRepository) CreateUserWithRole(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		switch user.RoleType {
		case "SPONSOR":
			sponsor := &models.Sponsor{
				ID:      user.ID,
				Balance: 0,
			}
			if err := tx.Create(sponsor).Error; err != nil {
				return err
			}

		case "ADMIN":
			admin := &models.Admin{
				ID:          user.ID,
				EmployeeID:  "EMP-" + user.ID.String()[:8],
				AccessLevel: "STANDARD",
			}
			if err := tx.Create(admin).Error; err != nil {
				return err
			}

		default:
			return errors.New("unsupported role type")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *authRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    return &user, err
}