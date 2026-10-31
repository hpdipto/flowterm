package tui

import (
	"time"

	"hpdipto/flowterm/activity"
	"hpdipto/flowterm/db"
)

type tracker struct {
	cur  *activity.Run
	runs []activity.Run
}

func (t *tracker) sample(now time.Time, s activity.Sample) {
	switch {
	case s.Locked || s.Idle > activity.IdleLimit:
		t.close(now, s)
	case t.cur != nil && t.cur.App == s.App:
		t.cur.End = now
	default:
		t.finish()
		t.cur = &activity.Run{App: s.App, Start: now, End: now}
	}
}

func (t *tracker) close(now time.Time, s activity.Sample) {
	if t.cur == nil {
		return
	}

	end := now
	if !s.Locked {
		end = now.Add(-s.Idle)
	}
	if end.Before(t.cur.Start) {
		end = t.cur.Start
	}

	t.cur.End = end
	t.finish()
}

func (t *tracker) finish() {
	if t.cur == nil {
		return
	}

	if t.cur.End.After(t.cur.Start) {
		t.runs = append(t.runs, *t.cur)
	}

	t.cur = nil
}

func (t *tracker) save(database *db.DB, sessionID int64) error {
	t.finish()

	for _, r := range t.runs {
		a := db.SessionActivity{
			SessionID: sessionID,
			App:       r.App,
			StartedAt: r.Start,
			EndedAt:   r.End,
		}

		if err := database.CreateSessionActivity(&a); err != nil {
			return err
		}
	}

	t.runs = nil

	return nil
}
