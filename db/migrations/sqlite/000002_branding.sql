-- +goose Up
-- +goose StatementBegin
CREATE TABLE branding (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    title TEXT,
    description TEXT,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_branding_user_id ON branding(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP IF EXISTS idx_branding_user_id;
DROP TABLE IF EXISTS branding;
-- +goose StatementEnd