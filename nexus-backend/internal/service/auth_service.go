package service

import (
"context"
"fmt"
"time"

"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
"github.com/luisforni/nexus-logistics/backend/internal/domain"
"github.com/rs/zerolog/log"
"golang.org/x/crypto/bcrypt"
)

type userReader interface {
FindByEmail(ctx context.Context, email string) (*domain.User, error)
FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type AuthService struct {
repo             userReader
jwtSecret        []byte
jwtExpiry        time.Duration
jwtRefreshExpiry time.Duration
}

func NewAuthService(repo userReader, secret string, expiry, refreshExpiry time.Duration) *AuthService {
return &AuthService{
repo:             repo,
jwtSecret:        []byte(secret),
jwtExpiry:        expiry,
jwtRefreshExpiry: refreshExpiry,
}
}

func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.TokenPair, error) {
user, err := s.repo.FindByEmail(ctx, req.Email)
if err != nil {

_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$dummyhashfortimingprotect"), []byte(req.Password))
log.Warn().
Str("email_domain", emailDomain(req.Email)).
Msg("login attempt: user not found")
return nil, fmt.Errorf("invalid credentials")
}

if !user.Active {
log.Warn().
Str("user_id", user.ID.String()).
Msg("login attempt: account inactive")
return nil, fmt.Errorf("account is inactive")
}

if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
log.Warn().
Str("user_id", user.ID.String()).
Msg("login attempt: invalid password")
return nil, fmt.Errorf("invalid credentials")
}

log.Info().
Str("user_id", user.ID.String()).
Str("role", string(user.Role)).
Msg("login success")

_ = s.repo.UpdateLastLogin(ctx, user.ID)
return s.generatePair(user)
}

func (s *AuthService) ValidateToken(tokenStr string) (*domain.Claims, error) {
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
}
return s.jwtSecret, nil
}, jwt.WithExpirationRequired())
if err != nil {
return nil, fmt.Errorf("invalid token: %w", err)
}

claims, ok := token.Claims.(jwt.MapClaims)
if !ok || !token.Valid {
return nil, fmt.Errorf("invalid token claims")
}

if claims["type"] != "access" {
return nil, fmt.Errorf("token type mismatch: expected access")
}

return &domain.Claims{
UserID: claims["sub"].(string),
Email:  claims["email"].(string),
Role:   domain.Role(claims["role"].(string)),
}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, fmt.Errorf("unexpected signing method")
}
return s.jwtSecret, nil
}, jwt.WithExpirationRequired())
if err != nil {
return nil, fmt.Errorf("invalid refresh token")
}

claims, ok := token.Claims.(jwt.MapClaims)
if !ok || !token.Valid {
return nil, fmt.Errorf("invalid token claims")
}

if claims["type"] != "refresh" {
return nil, fmt.Errorf("not a refresh token")
}

uid, err := uuid.Parse(claims["sub"].(string))
if err != nil {
return nil, fmt.Errorf("invalid subject")
}

user, err := s.repo.FindByID(ctx, uid)
if err != nil || !user.Active {
return nil, fmt.Errorf("user not found or inactive")
}

log.Info().Str("user_id", uid.String()).Msg("token refreshed")
return s.generatePair(user)
}

func (s *AuthService) generatePair(user *domain.User) (*domain.TokenPair, error) {
now := time.Now()

accessClaims := jwt.MapClaims{
"sub":   user.ID.String(),
"email": user.Email,
"role":  string(user.Role),
"type":  "access",
"iat":   now.Unix(),
"exp":   now.Add(s.jwtExpiry).Unix(),
}
accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
if err != nil {
return nil, err
}

refreshClaims := jwt.MapClaims{
"sub":  user.ID.String(),
"type": "refresh",
"iat":  now.Unix(),
"exp":  now.Add(s.jwtRefreshExpiry).Unix(),
}
refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
if err != nil {
return nil, err
}

return &domain.TokenPair{
AccessToken:  accessToken,
RefreshToken: refreshToken,
ExpiresIn:    int64(s.jwtExpiry.Seconds()),
}, nil
}

func emailDomain(email string) string {
for i := len(email) - 1; i >= 0; i-- {
if email[i] == '@' {
return email[i:]
}
}
return "unknown"
}
