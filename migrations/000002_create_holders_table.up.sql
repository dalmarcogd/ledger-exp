CREATE TABLE IF NOT EXISTS holders
(
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    document_number VARCHAR(100) NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX holders_document_number ON holders (document_number);
