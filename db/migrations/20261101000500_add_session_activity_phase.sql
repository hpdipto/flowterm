-- +goose Up

ALTER TABLE session_activities ADD COLUMN phase TEXT NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE session_activities DROP COLUMN phase;
