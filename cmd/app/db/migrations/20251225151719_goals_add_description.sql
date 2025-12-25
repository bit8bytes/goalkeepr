-- +goose Up
-- +goose StatementBegin
ALTER TABLE goals ADD description TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE goals DROP description; 
-- +goose StatementEnd
