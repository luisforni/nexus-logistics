package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleOperator Role = "OPERATOR"
	RoleViewer   Role = "VIEWER"
)

type User struct {
	ID           uuid.UUID  `json:"id"         gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string     `json:"email"      gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-"          gorm:"not null"`
	FirstName    string     `json:"first_name" gorm:"not null"`
	LastName     string     `json:"last_name"  gorm:"not null"`
	Role         Role       `json:"role"       gorm:"not null;default:'VIEWER'"`
	Active       bool       `json:"active"     gorm:"not null;default:true"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
}
