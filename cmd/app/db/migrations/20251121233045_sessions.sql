-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BLOB NOT NULL,
	expiry REAL NOT NULL
) STRICT;

CREATE INDEX sessions_expiry_idx ON sessions(expiry);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS sessions_expiry_idx;
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
