package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"github.com/rs/zerolog/log"
)

const shipmentCacheTTL = 30 * time.Second

type cacheClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type shipmentRepository interface {
	Create(ctx context.Context, s *domain.Shipment) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)
	FindByTrackingNumber(ctx context.Context, tn string) (*domain.Shipment, error)
	ListBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error)
	Update(ctx context.Context, s *domain.Shipment) error
	AddEvent(ctx context.Context, event *domain.ShipmentEvent) error
}

type chainLogger interface {
	RecordEvent(ctx context.Context, shipmentID, status, notes string) (string, error)
}

type ShipmentService struct {
	repo        shipmentRepository
	chainClient chainLogger
	cache       cacheClient
}

func NewShipmentService(repo shipmentRepository, chainClient chainLogger, rdb cacheClient) *ShipmentService {
	return &ShipmentService{repo: repo, chainClient: chainClient, cache: rdb}
}

func (s *ShipmentService) Create(ctx context.Context, req domain.CreateShipmentRequest, senderID uuid.UUID) (*domain.Shipment, error) {
	shipment := &domain.Shipment{
		ID:             uuid.New(),
		TrackingNumber: generateTrackingNumber(),
		Status:         domain.StatusPending,
		SenderID:       senderID,
		RecipientName:  req.RecipientName,
		RecipientEmail: req.RecipientEmail,
		OriginAddress:  req.Origin,
		DestAddress:    req.Destination,
		WeightKg:       req.WeightKg,
		DimensionsCm:   req.Dimensions,
		EstimatedAt:    req.EstimatedAt,
	}

	if err := s.repo.Create(ctx, shipment); err != nil {
		return nil, fmt.Errorf("persist shipment: %w", err)
	}

	log.Info().
		Str("shipment_id", shipment.ID.String()).
		Str("tracking_number", shipment.TrackingNumber).
		Str("sender_id", senderID.String()).
		Msg("shipment created")

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if s.chainClient != nil {
			txHash, err := s.chainClient.RecordEvent(bgCtx, shipment.ID.String(), string(domain.StatusPending), "")
			if err != nil {
				log.Error().Err(err).Str("shipment_id", shipment.ID.String()).Msg("blockchain anchor failed")
				return
			}
			shipment.BlockchainTxHash = txHash
			if err := s.repo.Update(bgCtx, shipment); err != nil {
				log.Error().Err(err).Msg("failed to persist tx hash")
			}
		}
	}()

	return shipment, nil
}

func (s *ShipmentService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	cacheKey := "shipment:" + id.String()
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			var shipment domain.Shipment
			if err := json.Unmarshal([]byte(cached), &shipment); err == nil {
				return &shipment, nil
			}
		}
	}

	shipment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if data, err := json.Marshal(shipment); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(data), shipmentCacheTTL)
		}
	}

	return shipment, nil
}

func (s *ShipmentService) UpdateStatus(ctx context.Context, id uuid.UUID, req domain.UpdateStatusRequest, actorID uuid.UUID) (*domain.Shipment, error) {
	shipment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isValidTransition(shipment.Status, req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", shipment.Status, req.Status)
	}

	event := &domain.ShipmentEvent{
		ID:         uuid.New(),
		ShipmentID: shipment.ID,
		Status:     req.Status,
		Location:   req.Location,
		Notes:      req.Notes,
		RecordedBy: actorID,
	}

	if err := s.repo.AddEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("persist event: %w", err)
	}

	shipment.Status = req.Status
	if req.Status == domain.StatusDelivered {
		now := time.Now()
		shipment.DeliveredAt = &now
	}

	if err := s.repo.Update(ctx, shipment); err != nil {
		return nil, fmt.Errorf("update shipment: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.Delete(ctx, "shipment:"+id.String())
	}

	log.Info().
		Str("shipment_id", id.String()).
		Str("status", string(req.Status)).
		Str("actor_id", actorID.String()).
		Msg("shipment status updated")

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if s.chainClient != nil {
			txHash, err := s.chainClient.RecordEvent(bgCtx, id.String(), string(req.Status), req.Notes)
			if err != nil {
				log.Error().Err(err).Msg("blockchain event anchor failed")
				return
			}
			event.TxHash = txHash
			if err := s.repo.AddEvent(bgCtx, event); err != nil {
				log.Error().Err(err).Msg("failed to update event tx hash")
			}
		}
	}()

	return shipment, nil
}

func (s *ShipmentService) List(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListBySender(ctx, senderID, limit, offset)
}

func generateTrackingNumber() string {
	return fmt.Sprintf("NX-%s", uuid.New().String()[:8])
}

var validTransitions = map[domain.ShipmentStatus]map[domain.ShipmentStatus]struct{}{
	domain.StatusPending:   {domain.StatusPickedUp: {}, domain.StatusFailed: {}},
	domain.StatusPickedUp:  {domain.StatusInTransit: {}, domain.StatusFailed: {}},
	domain.StatusInTransit: {domain.StatusAtHub: {}, domain.StatusOutForDel: {}, domain.StatusFailed: {}},
	domain.StatusAtHub:     {domain.StatusInTransit: {}, domain.StatusOutForDel: {}, domain.StatusFailed: {}},
	domain.StatusOutForDel: {domain.StatusDelivered: {}, domain.StatusFailed: {}},
	domain.StatusFailed:    {domain.StatusReturned: {}},
}

func isValidTransition(from, to domain.ShipmentStatus) bool {
	ifs, ok := validTransitions[from]
	if !ok {
		return false
	}
	_, ok = ifs[to]
	return ok
}
