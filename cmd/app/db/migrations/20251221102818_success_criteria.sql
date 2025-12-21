-- +goose Up
-- +goose StatementBegin
CREATE TABLE success_criteria (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    goal_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    completed INTEGER DEFAULT 0,
    position INTEGER,
    created_at INTEGER NOT NULL,

    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) STRICT;

CREATE INDEX idx_success_criteria_goal_id ON success_criteria(goal_id);
CREATE INDEX idx_success_criteria_user_id ON success_criteria(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_success_criteria_user_id;
DROP INDEX IF EXISTS idx_success_criteria_goal_id;
DROP TABLE IF EXISTS success_criteria;
-- +goose StatementEnd
