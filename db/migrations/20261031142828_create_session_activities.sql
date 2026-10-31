-- +goose Up

CREATE TABLE session_activities (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    app TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    ended_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX idx_session_activities_session_id
    ON session_activities (session_id, started_at);

-- +goose Down

DROP INDEX idx_session_activities_session_id;
DROP TABLE session_activities;
