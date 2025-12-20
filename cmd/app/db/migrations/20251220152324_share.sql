-- +goose Up
-- +goose StatementBegin
CREATE TABLE share (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    public_id TEXT NOT NULL UNIQUE,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) STRICT;

CREATE INDEX idx_share_user_id ON share(user_id);
CREATE INDEX idx_share_public_id ON share(public_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_share_public_id;
DROP INDEX IF EXISTS idx_share_user_id;
DROP TABLE IF EXISTS share;
-- +goose StatementEnd
