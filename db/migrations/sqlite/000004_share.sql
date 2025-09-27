-- +goose Up
-- +goose StatementBegin
CREATE TABLE share (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    public_id TEXT,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_share_user_id ON share(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP IF EXISTS idx_share_user_id;
DROP TABLE IF EXISTS share;
-- +goose StatementEnd