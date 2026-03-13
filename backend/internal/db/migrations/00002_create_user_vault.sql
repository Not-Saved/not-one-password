-- +goose Up
-- +goose StatementBegin
CREATE TABLE vaults (
    user_id INTEGER PRIMARY KEY,
    vault BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_vault_user
          FOREIGN KEY (user_id)
          REFERENCES users(id)
          ON DELETE CASCADE
);

CREATE INDEX idx_vaults_user_updated
ON vaults (user_id, updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE vaults;
-- +goose StatementEnd
