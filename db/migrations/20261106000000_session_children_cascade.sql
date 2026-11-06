-- +goose Up

DELETE FROM session_activities
WHERE session_id NOT IN (SELECT id FROM sessions);

DELETE FROM session_tags
WHERE session_id NOT IN (SELECT id FROM sessions);

CREATE TABLE session_activities_new (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    app TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    ended_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    app_name TEXT NOT NULL DEFAULT '',
    phase TEXT NOT NULL DEFAULT '',

    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

INSERT INTO session_activities_new (
    id,
    session_id,
    app,
    started_at,
    ended_at,
    created_at,
    updated_at,
    app_name,
    phase
)
SELECT
    id,
    session_id,
    app,
    started_at,
    ended_at,
    created_at,
    updated_at,
    app_name,
    phase
FROM session_activities;

DROP TABLE session_activities;

ALTER TABLE session_activities_new RENAME TO session_activities;

CREATE INDEX idx_session_activities_session_id
    ON session_activities (session_id, started_at);

CREATE TABLE session_tags_new (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE (session_id, tag_id),

    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

INSERT INTO session_tags_new (id, session_id, tag_id, created_at, updated_at)
SELECT id, session_id, tag_id, created_at, updated_at
FROM session_tags;

DROP TABLE session_tags;

ALTER TABLE session_tags_new RENAME TO session_tags;

-- +goose Down

CREATE TABLE session_activities_old (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    app TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    ended_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    app_name TEXT NOT NULL DEFAULT '',
    phase TEXT NOT NULL DEFAULT '',

    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

INSERT INTO session_activities_old (
    id,
    session_id,
    app,
    started_at,
    ended_at,
    created_at,
    updated_at,
    app_name,
    phase
)
SELECT
    id,
    session_id,
    app,
    started_at,
    ended_at,
    created_at,
    updated_at,
    app_name,
    phase
FROM session_activities;

DROP TABLE session_activities;

ALTER TABLE session_activities_old RENAME TO session_activities;

CREATE INDEX idx_session_activities_session_id
    ON session_activities (session_id, started_at);

CREATE TABLE session_tags_old (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE (session_id, tag_id),

    FOREIGN KEY (session_id) REFERENCES sessions(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

INSERT INTO session_tags_old (id, session_id, tag_id, created_at, updated_at)
SELECT id, session_id, tag_id, created_at, updated_at
FROM session_tags;

DROP TABLE session_tags;

ALTER TABLE session_tags_old RENAME TO session_tags;
