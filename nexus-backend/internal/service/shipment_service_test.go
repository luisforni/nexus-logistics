package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockShipmentRepo struct {
	createFn       func(ctx context.Context, s *domain.Shipment) error
	findByIDFn     func(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)
	listBySenderFn func(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error)
	updateFn       func(ctx context.Context, s *domain.Shipment) error
	addEventFn     func(ctx context.Context, event *domain.ShipmentEvent) error
}

func (m *mockShipmentRepo) Create(ctx context.Context, s *domain.Shipment) error {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}

func (m *mockShipmentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
}

func (m *mockShipmentRepo) FindByTrackingNumber(_ context.Context, _ string) (*domain.Shipment, error) {
	return nil, nil
}

func (m *mockShipmentRepo) ListBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error) {
	if m.listBySenderFn != nil {
		return m.listBySenderFn(ctx, senderID, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockShipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, s)
	}
	return nil
}

func (m *mockShipmentRepo) AddEvent(ctx context.Context, event *domain.ShipmentEvent) error {
	if m.addEventFn != nil {
		return m.addEventFn(ctx, event)
	}
	return nil
}

type mockCache struct {
	getVal string
	getErr error
	setErr error
}

func (m *mockCache) Get(_ context.Context, _ string) (string, error) {
	return m.getVal, m.getErr
}

func (m *mockCache) Set(_ context.Context, _, _ string, _ time.Duration) error {
	return m.setErr
}

func (m *mockCache) Delete(_ context.Context, _ ...string) error {
	return nil
}

type mockChain struct {
	recordEventFn func(ctx context.Context, shipmentID, status, notes string) (string, error)
}

func (m *mockChain) RecordEvent(ctx context.Context, shipmentID, status, notes string) (string, error) {
	if m.recordEventFn != nil {
		return m.recordEventFn(ctx, shipmentID, status, notes)
	}
	return "0xdeadbeef", nil
}

func TestShipmentService_Create(t *testing.T) {
	senderID := uuid.New()

	tests := []struct {
		name    string
		req     domain.CreateShipmentRequest
		wantErr bool
	}{
		{
			name: "valid shipment",
			req: domain.CreateShipmentRequest{
				RecipientName:  "Alice",
				RecipientEmail: "alice@example.com",
				Origin:         domain.Address{City: "Madrid", Country: "ES"},
				Destination:    domain.Address{City: "Barcelona", Country: "ES"},
				WeightKg:       2.5,
				EstimatedAt:    time.Now().Add(48 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: domain.CreateShipmentRequest{
				RecipientName: "Bob",
				Origin:        domain.Address{City: "Sevilla", Country: "ES"},
				Destination:   domain.Address{City: "Malaga", Country: "ES"},
				WeightKg:      1.0,
				EstimatedAt:   time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockShipmentRepo{}
			if tc.wantErr {
				repo.createFn = func(_ context.Context, _ *domain.Shipment) error {
					return fmt.Errorf("db error")
				}
			}
			svc := NewShipmentService(repo, nil, &mockCache{getErr: fmt.Errorf("miss")})

			result, err := svc.Create(context.Background(), tc.req, senderID)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, result.TrackingNumber)
			assert.Equal(t, domain.StatusPending, result.Status)
			assert.Equal(t, senderID, result.SenderID)
		})
	}
}

func TestShipmentService_UpdateStatus(t *testing.T) {
	actorID := uuid.New()
	shipmentID := uuid.New()

	tests := []struct {
		name    string
		from    domain.ShipmentStatus
		to      domain.ShipmentStatus
		wantErr bool
	}{
		{"pending to picked_up valid", domain.StatusPending, domain.StatusPickedUp, false},
		{"picked_up to in_transit valid", domain.StatusPickedUp, domain.StatusInTransit, false},
		{"in_transit to out_for_delivery valid", domain.StatusInTransit, domain.StatusOutForDel, false},
		{"out_for_delivery to delivered valid", domain.StatusOutForDel, domain.StatusDelivered, false},
		{"pending to delivered invalid", domain.StatusPending, domain.StatusDelivered, true},
		{"delivered to pending invalid", domain.StatusDelivered, domain.StatusPending, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockShipmentRepo{
				findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
					return &domain.Shipment{ID: id, Status: tc.from}, nil
				},
			}
			svc := NewShipmentService(repo, nil, nil)

			req := domain.UpdateStatusRequest{Status: tc.to, Location: "Test Hub", Notes: "test"}
			_, err := svc.UpdateStatus(context.Background(), shipmentID, req, actorID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShipmentService_List(t *testing.T) {
	senderID := uuid.New()
	shipments := []domain.Shipment{
		{ID: uuid.New(), Status: domain.StatusPending},
		{ID: uuid.New(), Status: domain.StatusInTransit},
	}

	repo := &mockShipmentRepo{
		listBySenderFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.Shipment, int64, error) {
			return shipments, int64(len(shipments)), nil
		},
	}
	svc := NewShipmentService(repo, nil, nil)

	got, total, err := svc.List(context.Background(), senderID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, got, 2)
}

func TestShipmentService_GetByID_CacheMiss(t *testing.T) {
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
		},
	}
	cache := &mockCache{getErr: fmt.Errorf("miss")}
	svc := NewShipmentService(repo, nil, cache)

	shipment, err := svc.GetByID(context.Background(), shipmentID)
	require.NoError(t, err)
	assert.Equal(t, shipmentID, shipment.ID)
}

func TestShipmentService_GetByID_CacheHit(t *testing.T) {
	shipmentID := uuid.New()
	cached := domain.Shipment{ID: shipmentID, Status: domain.StatusPending}
	data, _ := json.Marshal(cached)
	svc := NewShipmentService(&mockShipmentRepo{}, nil, &mockCache{getVal: string(data)})

	shipment, err := svc.GetByID(context.Background(), shipmentID)
	require.NoError(t, err)
	assert.Equal(t, shipmentID, shipment.ID)
}

func TestShipmentService_GetByID_CacheHit_InvalidJSON(t *testing.T) {
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
		},
	}
	svc := NewShipmentService(repo, nil, &mockCache{getVal: "not-valid-json"})

	shipment, err := svc.GetByID(context.Background(), shipmentID)
	require.NoError(t, err)
	assert.Equal(t, shipmentID, shipment.ID)
}

func TestShipmentService_GetByID_RepoError(t *testing.T) {
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	svc := NewShipmentService(repo, nil, &mockCache{getErr: fmt.Errorf("miss")})
	_, err := svc.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestShipmentService_GetByID_NilCache(t *testing.T) {
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusInTransit}, nil
		},
	}
	svc := NewShipmentService(repo, nil, nil)
	shipment, err := svc.GetByID(context.Background(), shipmentID)
	require.NoError(t, err)
	assert.Equal(t, shipmentID, shipment.ID)
}

func TestShipmentService_Create_WithChain(t *testing.T) {
	svc := NewShipmentService(&mockShipmentRepo{}, &mockChain{}, &mockCache{getErr: fmt.Errorf("miss")})
	result, err := svc.Create(context.Background(), domain.CreateShipmentRequest{
		RecipientName: "Test",
		Origin:        domain.Address{City: "A", Country: "ES"},
		Destination:   domain.Address{City: "B", Country: "ES"},
		WeightKg:      1,
		EstimatedAt:   time.Now().Add(24 * time.Hour),
	}, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestShipmentService_Create_ChainError(t *testing.T) {
	chain := &mockChain{
		recordEventFn: func(_ context.Context, _, _, _ string) (string, error) {
			return "", fmt.Errorf("chain down")
		},
	}
	svc := NewShipmentService(&mockShipmentRepo{}, chain, nil)
	result, err := svc.Create(context.Background(), domain.CreateShipmentRequest{
		RecipientName: "Test",
		Origin:        domain.Address{City: "A", Country: "ES"},
		Destination:   domain.Address{City: "B", Country: "ES"},
		WeightKg:      1,
		EstimatedAt:   time.Now().Add(24 * time.Hour),
	}, uuid.New())
	require.NoError(t, err)
	assert.NotNil(t, result)

	time.Sleep(50 * time.Millisecond)
}

func TestShipmentService_UpdateStatus_Delivered(t *testing.T) {
	actorID := uuid.New()
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusOutForDel}, nil
		},
	}
	cache := &mockCache{getErr: fmt.Errorf("miss")}
	svc := NewShipmentService(repo, nil, cache)

	updated, err := svc.UpdateStatus(context.Background(), shipmentID,
		domain.UpdateStatusRequest{Status: domain.StatusDelivered}, actorID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDelivered, updated.Status)
	assert.NotNil(t, updated.DeliveredAt)
}

func TestShipmentService_UpdateStatus_WithChain(t *testing.T) {
	actorID := uuid.New()
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
		},
	}
	svc := NewShipmentService(repo, &mockChain{}, &mockCache{getErr: fmt.Errorf("miss")})
	_, err := svc.UpdateStatus(context.Background(), shipmentID,
		domain.UpdateStatusRequest{Status: domain.StatusPickedUp}, actorID)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
}

func TestShipmentService_UpdateStatus_FindError(t *testing.T) {
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Shipment, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := NewShipmentService(repo, nil, nil)
	_, err := svc.UpdateStatus(context.Background(), uuid.New(),
		domain.UpdateStatusRequest{Status: domain.StatusPickedUp}, uuid.New())
	assert.Error(t, err)
}

func TestShipmentService_UpdateStatus_AddEventError(t *testing.T) {
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
		},
		addEventFn: func(_ context.Context, _ *domain.ShipmentEvent) error {
			return fmt.Errorf("persist fail")
		},
	}
	svc := NewShipmentService(repo, nil, nil)
	_, err := svc.UpdateStatus(context.Background(), shipmentID,
		domain.UpdateStatusRequest{Status: domain.StatusPickedUp}, uuid.New())
	assert.Error(t, err)
}

func TestShipmentService_UpdateStatus_UpdateError(t *testing.T) {
	shipmentID := uuid.New()
	repo := &mockShipmentRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Shipment, error) {
			return &domain.Shipment{ID: id, Status: domain.StatusPending}, nil
		},
		updateFn: func(_ context.Context, _ *domain.Shipment) error {
			return fmt.Errorf("update fail")
		},
	}
	svc := NewShipmentService(repo, nil, nil)
	_, err := svc.UpdateStatus(context.Background(), shipmentID,
		domain.UpdateStatusRequest{Status: domain.StatusPickedUp}, uuid.New())
	assert.Error(t, err)
}

func TestShipmentService_List_InvalidPagination(t *testing.T) {
	senderID := uuid.New()
	var gotLimit, gotOffset int
	repo := &mockShipmentRepo{
		listBySenderFn: func(_ context.Context, _ uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error) {
			gotLimit, gotOffset = limit, offset
			return nil, 0, nil
		},
	}
	svc := NewShipmentService(repo, nil, nil)

	_, _, err := svc.List(context.Background(), senderID, 0, -5)
	require.NoError(t, err)
	assert.Equal(t, 20, gotLimit)
	assert.Equal(t, 0, gotOffset)

	gotLimit = 0
	_, _, err = svc.List(context.Background(), senderID, 200, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, gotLimit)
}

func TestShipmentService_List_Error(t *testing.T) {
	repo := &mockShipmentRepo{
		listBySenderFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]domain.Shipment, int64, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	svc := NewShipmentService(repo, nil, nil)
	_, _, err := svc.List(context.Background(), uuid.New(), 10, 0)
	assert.Error(t, err)
}

func TestGenerateTrackingNumber(t *testing.T) {
	tn := generateTrackingNumber()
	assert.True(t, len(tn) > 0)
	assert.Contains(t, tn, "NX-")
}
