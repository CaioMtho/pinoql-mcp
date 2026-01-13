-- +goose Up
-- +goose StatementBegin
ALTER TABLE connection_data ADD COLUMN dek TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE connection_data DROP COLUMN dek;
-- +goose StatementEnd
