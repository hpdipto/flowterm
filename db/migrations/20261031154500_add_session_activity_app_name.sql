-- +goose Up

ALTER TABLE session_activities ADD COLUMN app_name TEXT NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE session_activities DROP COLUMN app_name;
