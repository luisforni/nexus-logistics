

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'VIEWER'
                              CHECK (role IN ('ADMIN', 'OPERATOR', 'VIEWER')),
    active        BOOLEAN     NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
