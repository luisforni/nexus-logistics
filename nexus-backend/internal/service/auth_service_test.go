package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/luisforni/nexus-logistics/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	findByEmailFn     func(ctx context.Context, email string) (*domain.User, error)
	findByIDFn        func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	updateLastLoginFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	if m.updateLastLoginFn != nil {
		return m.updateLastLoginFn(ctx, id)
	}
	return nil
}

const testSecret = "test-jwt-secret-that-is-at-least-32-chars!!"

func newTestAuthService(repo *mockUserRepo) *service.AuthService {
	return service.NewAuthService(repo, testSecret, 15*time.Minute, 7*24*time.Hour)
}

func hashPassword(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt failed: %v", err)
	}
	return string(h)
}

func activeUser(t *testing.T, email, password string) *domain.User {
	t.Helper()
	return &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashPassword(t, password),
		Role:         domain.RoleOperator,
		Active:       true,
	}
}

func TestAuthService_Login(t *testing.T) {
	const email = "user@example.com"
	const password = "securePass123!"

	t.Run("success returns token pair", func(t *testing.T) {
		user := activeUser(t, email, password)
		repo := &mockUserRepo{
			findByEmailFn:     func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
			updateLastLoginFn: func(_ context.Context, _ uuid.UUID) error { return nil },
		}
		svc := newTestAuthService(repo)

		pair, err := svc.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pair.AccessToken == "" {
			t.Error("access token is empty")
		}
		if pair.RefreshToken == "" {
			t.Error("refresh token is empty")
		}
	})

	t.Run("wrong password returns error", func(t *testing.T) {
		user := activeUser(t, email, password)
		repo := &mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
		}
		svc := newTestAuthService(repo)

		_, err := svc.Login(context.Background(), domain.LoginRequest{Email: email, Password: "wrongPassword"})
		if err == nil {
			t.Fatal("expected error for wrong password")
		}
	})

	t.Run("unknown email returns error without leaking existence", func(t *testing.T) {
		repo := &mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
				return nil, errors.New("not found")
			},
		}
		svc := newTestAuthService(repo)

		_, err := svc.Login(context.Background(), domain.LoginRequest{Email: "ghost@example.com", Password: "anything"})
		if err == nil {
			t.Fatal("expected error for unknown email")
		}
	})

	t.Run("inactive user returns error", func(t *testing.T) {
		user := activeUser(t, email, password)
		user.Active = false
		repo := &mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
		}
		svc := newTestAuthService(repo)

		_, err := svc.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})
		if err == nil {
			t.Fatal("expected error for inactive user")
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	const email = "tok@example.com"
	const password = "pass123"

	user := activeUser(t, email, password)
	repo := &mockUserRepo{
		findByEmailFn:     func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
		updateLastLoginFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	svc := newTestAuthService(repo)

	pair, err := svc.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("valid access token parses successfully", func(t *testing.T) {
		claims, err := svc.ValidateToken(pair.AccessToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.UserID != user.ID.String() {
			t.Errorf("expected user id %s, got %s", user.ID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("expected email %s, got %s", email, claims.Email)
		}
	})

	t.Run("refresh token rejected by ValidateToken", func(t *testing.T) {

		_ = pair.RefreshToken
	})

	t.Run("tampered token rejected", func(t *testing.T) {
		_, err := svc.ValidateToken(pair.AccessToken + "tampered")
		if err == nil {
			t.Fatal("expected error for tampered token")
		}
	})

	t.Run("empty string rejected", func(t *testing.T) {
		_, err := svc.ValidateToken("")
		if err == nil {
			t.Fatal("expected error for empty token")
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	const email = "refresh@example.com"
	const password = "refreshPass!"

	user := activeUser(t, email, password)
	repo := &mockUserRepo{
		findByEmailFn:     func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
		findByIDFn:        func(_ context.Context, _ uuid.UUID) (*domain.User, error) { return user, nil },
		updateLastLoginFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	svc := newTestAuthService(repo)

	pair, err := svc.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("valid refresh token issues new pair", func(t *testing.T) {
		newPair, err := svc.RefreshToken(context.Background(), pair.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newPair.AccessToken == "" {
			t.Error("new access token is empty")
		}
	})

	t.Run("access token rejected as refresh token", func(t *testing.T) {
		_, err := svc.RefreshToken(context.Background(), pair.AccessToken)
		if err == nil {
			t.Fatal("expected error when using access token as refresh token")
		}
	})

	t.Run("garbage token rejected", func(t *testing.T) {
		_, err := svc.RefreshToken(context.Background(), "not.a.token")
		if err == nil {
			t.Fatal("expected error for garbage token")
		}
	})

	t.Run("user not found returns error", func(t *testing.T) {
		brokenRepo := &mockUserRepo{
			findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
				return nil, errors.New("user gone")
			},
			updateLastLoginFn: func(_ context.Context, _ uuid.UUID) error { return nil },
		}
		brokenSvc := newTestAuthService(brokenRepo)

		brokenSvc2 := service.NewAuthService(repo, testSecret, 15*time.Minute, 7*24*time.Hour)
		p2, _ := brokenSvc2.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})

		_, err := brokenSvc.RefreshToken(context.Background(), p2.RefreshToken)
		if err == nil {
			t.Fatal("expected error when user cannot be found during refresh")
		}
	})
}

func TestAuthService_ValidateToken_RefreshTokenRejected(t *testing.T) {
const email = "vt@example.com"
const password = "pass123"
user := activeUser(t, email, password)
repo := &mockUserRepo{
findByEmailFn:     func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
updateLastLoginFn: func(_ context.Context, _ uuid.UUID) error { return nil },
}
svc := newTestAuthService(repo)
pair, err := svc.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})
if err != nil {
t.Fatalf("login failed: %v", err)
}

_, err = svc.ValidateToken(pair.RefreshToken)
if err == nil {
t.Fatal("expected error: refresh token used as access token")
}
}

func TestAuthService_RefreshToken_InactiveUser(t *testing.T) {
const email = "inactive@example.com"
const password = "pass123"
user := activeUser(t, email, password)

goodRepo := &mockUserRepo{
findByEmailFn:     func(_ context.Context, _ string) (*domain.User, error) { return user, nil },
findByIDFn:        func(_ context.Context, _ uuid.UUID) (*domain.User, error) { return user, nil },
updateLastLoginFn: func(_ context.Context, _ uuid.UUID) error { return nil },
}
goodSvc := newTestAuthService(goodRepo)
pair, _ := goodSvc.Login(context.Background(), domain.LoginRequest{Email: email, Password: password})

inactive := *user
inactive.Active = false
badRepo := &mockUserRepo{
findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) { return &inactive, nil },
}
badSvc := newTestAuthService(badRepo)
_, err := badSvc.RefreshToken(context.Background(), pair.RefreshToken)
if err == nil {
t.Fatal("expected error for inactive user during refresh")
}
}
