package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentStatus string

const (
	StatusPending    ShipmentStatus = "PENDING"
	StatusPickedUp   ShipmentStatus = "PICKED_UP"
	StatusInTransit  ShipmentStatus = "IN_TRANSIT"
	StatusAtHub      ShipmentStatus = "AT_HUB"
	StatusOutForDel  ShipmentStatus = "OUT_FOR_DELIVERY"
	StatusDelivered  ShipmentStatus = "DELIVERED"
	StatusFailed     ShipmentStatus = "FAILED"
	StatusReturned   ShipmentStatus = "RETURNED"
)

type Shipment struct {
	ID               uuid.UUID      `json:"id"                gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TrackingNumber   string         `json:"tracking_number"   gorm:"uniqueIndex;not null"`
	Status           ShipmentStatus `json:"status"            gorm:"not null;default:'PENDING'"`
	SenderID         uuid.UUID      `json:"sender_id"         gorm:"type:uuid;not null;index"`
	RecipientName    string         `json:"recipient_name"    gorm:"not null"`
	RecipientEmail   string         `json:"recipient_email"`
	OriginAddress    Address        `json:"origin"            gorm:"embedded;embeddedPrefix:origin_"`
	DestAddress      Address        `json:"destination"       gorm:"embedded;embeddedPrefix:dest_"`
	WeightKg         float64        `json:"weight_kg"`
	DimensionsCm     Dimensions     `json:"dimensions"        gorm:"embedded;embeddedPrefix:dim_"`
	EstimatedAt      time.Time      `json:"estimated_at"`
	DeliveredAt      *time.Time     `json:"delivered_at,omitempty"`
	BlockchainTxHash string         `json:"blockchain_tx_hash,omitempty" gorm:"index"`
	Events           []ShipmentEvent `json:"events,omitempty" gorm:"foreignKey:ShipmentID"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type Address struct {
	Street     string  `json:"street"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

type Dimensions struct {
	LengthCm float64 `json:"length_cm"`
	WidthCm  float64 `json:"width_cm"`
	HeightCm float64 `json:"height_cm"`
}

type ShipmentEvent struct {
	ID         uuid.UUID      `json:"id"          gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ShipmentID uuid.UUID      `json:"shipment_id" gorm:"type:uuid;not null;index"`
	Status     ShipmentStatus `json:"status"      gorm:"not null"`
	Location   string         `json:"location"`
	Notes      string         `json:"notes"`
	RecordedBy uuid.UUID      `json:"recorded_by" gorm:"type:uuid"`
	TxHash     string         `json:"tx_hash"     gorm:"index"`
	CreatedAt  time.Time      `json:"created_at"`
}

type CreateShipmentRequest struct {
	RecipientName  string     `json:"recipient_name"  binding:"required,max=120"`
	RecipientEmail string     `json:"recipient_email" binding:"omitempty,email"`
	Origin         Address    `json:"origin"          binding:"required"`
	Destination    Address    `json:"destination"     binding:"required"`
	WeightKg       float64    `json:"weight_kg"       binding:"required,gt=0"`
	Dimensions     Dimensions `json:"dimensions"`
	EstimatedAt    time.Time  `json:"estimated_at"    binding:"required"`
}

type UpdateStatusRequest struct {
	Status   ShipmentStatus `json:"status"   binding:"required,oneof=PICKED_UP IN_TRANSIT AT_HUB OUT_FOR_DELIVERY DELIVERED FAILED RETURNED"`
	Location string         `json:"location" binding:"max=200"`
	Notes    string         `json:"notes"    binding:"max=500"`
}
