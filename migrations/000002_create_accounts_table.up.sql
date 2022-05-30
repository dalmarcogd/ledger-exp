CREATE TABLE IF NOT EXISTS accounts
(
    id         VARCHAR(36) PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NULL
);
