

CREATE TYPE shipment_status AS ENUM (
    'PENDING',
    'PICKED_UP',
    'IN_TRANSIT',
    'AT_HUB',
    'OUT_FOR_DELIVERY',
    'DELIVERED',
    'FAILED',
    'RETURNED'
);

CREATE TABLE IF NOT EXISTS shipments (
    id                  UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_number     TEXT             NOT NULL,
    status              shipment_status  NOT NULL DEFAULT 'PENDING',
    sender_id           UUID             NOT NULL REFERENCES users(id),

    recipient_name      TEXT             NOT NULL,
    recipient_email     TEXT,

    origin_street       TEXT,
    origin_city         TEXT,
    origin_state        TEXT,
    origin_country      TEXT,
    origin_postal_code  TEXT,
    origin_latitude     DOUBLE PRECISION,
    origin_longitude    DOUBLE PRECISION,

    dest_street         TEXT,
    dest_city           TEXT,
    dest_state          TEXT,
    dest_country        TEXT,
    dest_postal_code    TEXT,
    dest_latitude       DOUBLE PRECISION,
    dest_longitude      DOUBLE PRECISION,

    weight_kg           DOUBLE PRECISION,
    dim_length_cm       DOUBLE PRECISION,
    dim_width_cm        DOUBLE PRECISION,
    dim_height_cm       DOUBLE PRECISION,

    estimated_at        TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    blockchain_tx_hash  TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shipments_tracking_number ON shipments (tracking_number);
CREATE INDEX IF NOT EXISTS idx_shipments_sender_id ON shipments (sender_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments (status);
CREATE INDEX IF NOT EXISTS idx_shipments_blockchain_tx_hash ON shipments (blockchain_tx_hash) WHERE blockchain_tx_hash IS NOT NULL;
