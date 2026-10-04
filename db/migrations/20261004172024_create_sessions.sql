-- +goose Up

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    focus_session_name TEXT NOT NULL,
    break_session_name TEXT NOT NULL,
    focus_session_duration INTEGER NOT NULL,
    break_session_duration INTEGER NOT NULL,
    interrupted BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE session_tags (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE (session_id, tag_id),

    FOREIGN KEY (session_id) REFERENCES sessions(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

-- +goose Down

DROP TABLE session_tags;
DROP TABLE tags;
DROP TABLE sessions;
