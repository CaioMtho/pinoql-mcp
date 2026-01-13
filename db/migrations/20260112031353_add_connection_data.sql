-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS connection_data (
   id TEXT PRIMARY KEY,
   dsn TEXT NOT NULL,
   dialect TEXT NOT NULL,
   readonly INTEGER NOT NULL DEFAULT 0,
   created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
   updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_connection_data_dsn ON connection_data(dsn);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_connection_data_dsn;
DROP TABLE IF EXISTS connection_data;
-- +goose StatementEnd
