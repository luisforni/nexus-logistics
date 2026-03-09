package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/api"
	"github.com/luisforni/nexus-logistics/backend/internal/config"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/luisforni/nexus-logistics/backend/internal/service"
	"github.com/stretchr/testify/assert"
)

type stubUserRepo struct{}

func (*stubUserRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, nil
}
func (*stubUserRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (*stubUserRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }

type stubShipRepo struct{}

func (*stubShipRepo) Create(_ context.Context, _ *domain.Shipment) error { return nil }
func (*stubShipRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
	return nil, nil
}
func (*stubShipRepo) FindByTrackingNumber(_ context.Context, _ string) (*domain.Shipment, error) {
	return nil, nil
}
func (*stubShipRepo) ListBySender(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.Shipment, int64, error) {
	return nil, 0, nil
}
func (*stubShipRepo) Update(_ context.Context, _ *domain.Shipment) error { return nil }
func (*stubShipRepo) AddEvent(_ context.Context, _ *domain.ShipmentEvent) error { return nil }

const routerTestSecret = "test-secret-that-is-at-least-32-chars!!"

func TestNewRouter_Development(t *testing.T) {
	cfg := &config.Config{AppEnv: "development", OptimizerHost: "localhost", OptimizerPort: "9999"}
	authSvc := service.NewAuthService(&stubUserRepo{}, routerTestSecret, 15*time.Minute, 7*24*time.Hour)
	shipSvc := service.NewShipmentService(&stubShipRepo{}, nil, nil)

	r := api.NewRouter(cfg, authSvc, shipSvc)
	assert.NotNil(t, r)
}

func TestNewRouter_Production(t *testing.T) {
	cfg := &config.Config{AppEnv: "production", OptimizerHost: "optimizer", OptimizerPort: "8080"}
	authSvc := service.NewAuthService(&stubUserRepo{}, routerTestSecret, 15*time.Minute, 7*24*time.Hour)
	shipSvc := service.NewShipmentService(&stubShipRepo{}, nil, nil)

	r := api.NewRouter(cfg, authSvc, shipSvc)
	assert.NotNil(t, r)
}
