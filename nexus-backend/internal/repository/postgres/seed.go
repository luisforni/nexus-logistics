package postgres

import (
	"context"
	"errors"
	"os"

	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdminUser(ctx context.Context, db *gorm.DB) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")

	if email == "" || password == "" {
		log.Info().Msg("ADMIN_EMAIL/ADMIN_PASSWORD not set – skipping admin seed")
		return nil
	}

	var count int64
	if err := db.WithContext(ctx).
		Model(&domain.User{}).
		Where("role = ?", domain.RoleAdmin).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Info().Msg("admin user already exists – skipping seed")
		return nil
	}

	var existing domain.User
	err := db.WithContext(ctx).First(&existing, "email = ?", email).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		log.Warn().Str("email", email).Msg("seed email already in use by another user – skipping")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    "Admin",
		LastName:     "Nexus",
		Role:         domain.RoleAdmin,
		Active:       true,
	}

	if err := db.WithContext(ctx).Create(admin).Error; err != nil {
		return err
	}

	log.Info().Str("email", email).Str("id", admin.ID.String()).Msg("admin user created")
	return nil
}
