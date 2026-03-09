package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/luisforni/nexus-logistics/backend/internal/domain"
	"gorm.io/gorm"
)

type ShipmentRepository struct {
	db *gorm.DB
}

func NewShipmentRepository(db *gorm.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

func (r *ShipmentRepository) Create(ctx context.Context, s *domain.Shipment) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *ShipmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	var s domain.Shipment
	err := r.db.WithContext(ctx).
		Preload("Events").
		First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("shipment %s not found", id)
	}
	return &s, err
}

func (r *ShipmentRepository) FindByTrackingNumber(ctx context.Context, tn string) (*domain.Shipment, error) {
	var s domain.Shipment
	err := r.db.WithContext(ctx).
		Preload("Events").
		First(&s, "tracking_number = ?", tn).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("shipment %s not found", tn)
	}
	return &s, err
}

func (r *ShipmentRepository) ListBySender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.Shipment, int64, error) {
	var shipments []domain.Shipment
	var total int64

	tx := r.db.WithContext(ctx).Model(&domain.Shipment{}).Where("sender_id = ?", senderID)
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Limit(limit).Offset(offset).Order("created_at DESC").Find(&shipments).Error; err != nil {
		return nil, 0, err
	}

	return shipments, total, nil
}

func (r *ShipmentRepository) Update(ctx context.Context, s *domain.Shipment) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *ShipmentRepository) AddEvent(ctx context.Context, event *domain.ShipmentEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}
