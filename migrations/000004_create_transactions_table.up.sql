CREATE TABLE IF NOT EXISTS transactions
(
    id              VARCHAR(36) PRIMARY KEY,
    from_account_id VARCHAR(36),
    to_account_id   VARCHAR(36),
    type            VARCHAR(36)  NOT NULL,
    amount          DECIMAL      NOT NULL,
    description     VARCHAR(200) NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    FOREIGN KEY (from_account_id) REFERENCES accounts (id),
    FOREIGN KEY (to_account_id) REFERENCES accounts (id)
);

CREATE INDEX transactions_from_account_id_index ON transactions (from_account_id);
CREATE INDEX transactions_to_account_id_index ON transactions (to_account_id);