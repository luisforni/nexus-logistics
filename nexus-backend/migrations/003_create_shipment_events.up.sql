

CREATE TABLE IF NOT EXISTS shipment_events (
    id           UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_id  UUID             NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status       shipment_status  NOT NULL,
    location     TEXT,
    notes        TEXT,
    recorded_by  UUID             REFERENCES users(id),
    tx_hash      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shipment_events_shipment_id ON shipment_events (shipment_id);
CREATE INDEX IF NOT EXISTS idx_shipment_events_tx_hash ON shipment_events (tx_hash) WHERE tx_hash IS NOT NULL;
