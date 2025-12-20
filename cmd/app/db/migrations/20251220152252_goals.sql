-- +goose Up
-- +goose StatementBegin
CREATE TABLE goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    goal TEXT,
    due INTEGER,
    visible_to_public INTEGER DEFAULT 0,
    achieved INTEGER DEFAULT 0,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) STRICT;

CREATE INDEX idx_goals_users_id ON goals(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_goals_user_id
DROP TABLE IF EXISTS goals;
-- +goose StatementEnd