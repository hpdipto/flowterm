-- +goose Up
INSERT INTO tags (name, created_at, updated_at)
VALUES ('work', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- +goose Down
DELETE FROM tags WHERE name = 'work';
