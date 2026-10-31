package db

import "time"

type BaseModel struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Session struct {
	BaseModel

	FocusSessionName     string
	BreakSessionName     string
	FocusSessionDuration int
	BreakSessionDuration int
	Interrupted          bool
}

type Tag struct {
	BaseModel

	Name string
}

type SessionTag struct {
	BaseModel

	SessionID int64
	TagID     int64
}

type SessionActivity struct {
	BaseModel

	SessionID int64
	App       string
	StartedAt time.Time
	EndedAt   time.Time
}
