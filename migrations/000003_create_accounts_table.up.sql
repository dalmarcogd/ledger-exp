CREATE TABLE IF NOT EXISTS accounts
(
    id         VARCHAR(36) PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    agency     VARCHAR(4)   NOT NULL,
    number     VARCHAR(50)  NOT NULL,
    holder_id  VARCHAR(36)  NOT NULL,
    status     VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NULL,
    FOREIGN KEY (holder_id) REFERENCES holders (id)
);

CREATE UNIQUE INDEX accounts_holder_number ON accounts (number, holder_id);
