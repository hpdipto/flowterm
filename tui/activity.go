package tui

import (
	"github.com/hpdipto/flowterm/activity"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/utils"
	"time"
)

const spanSaveInterval = 5 * time.Second

type span struct {
	App   string
	Name  string
	Phase db.Phase
	Start time.Time
	End   time.Time
	ID    int64
	Saved time.Time
}

type tracker struct {
	database  *db.DB
	sessionID int64

	cur  *span
	runs []span
}

func newTracker(database *db.DB, sessionID int64) *tracker {
	return &tracker{database: database, sessionID: sessionID}
}

func (t *tracker) sample(now time.Time, s activity.Sample, phase db.Phase) {
	switch {
	case s.Locked || s.Idle > activity.IdleLimit:
		t.close(now, s)
	case t.cur != nil && t.cur.App == s.App && t.cur.Phase == phase:
		t.cur.End = now
		t.cur.Name = s.Name
		t.persist(now)
	default:
		t.finish()
		t.cur = &span{App: s.App, Name: s.Name, Phase: phase, Start: now, End: now}
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

	s := *t.cur
	t.cur = nil

	if !s.End.After(s.Start) {
		return
	}

	if s.ID != 0 {
		_ = t.database.UpdateSessionActivityEnd(s.ID, s.End)
		return
	}

	t.runs = append(t.runs, s)
	utils.LogError(t.flush())
}

func (t *tracker) flush() error {
	if len(t.runs) == 0 {
		return nil
	}

	var firstErr error
	kept := t.runs[:0]

	for _, r := range t.runs {
		a := db.SessionActivity{
			SessionID: t.sessionID,
			App:       r.App,
			AppName:   r.Name,
			Phase:     r.Phase,
			StartedAt: r.Start,
			EndedAt:   r.End,
		}

		if err := t.database.CreateSessionActivity(&a); err != nil {
			if firstErr == nil {
				firstErr = err
			}

			kept = append(kept, r)
		}
	}

	t.runs = kept

	return firstErr
}

func (t *tracker) stop() error {
	t.finish()
	return t.flush()
}

func (t *tracker) persist(now time.Time) {
	if t.cur == nil || !t.cur.End.After(t.cur.Start) {
		return
	}

	if t.cur.ID != 0 {
		if now.Sub(t.cur.Saved) < spanSaveInterval {
			return
		}

		_ = t.database.UpdateSessionActivityEnd(t.cur.ID, t.cur.End)
		t.cur.Saved = now

		return
	}

	a := db.SessionActivity{
		SessionID: t.sessionID,
		App:       t.cur.App,
		AppName:   t.cur.Name,
		Phase:     t.cur.Phase,
		StartedAt: t.cur.Start,
		EndedAt:   t.cur.End,
	}

	if err := t.database.CreateSessionActivity(&a); err != nil {
		return
	}

	t.cur.ID = a.ID
	t.cur.Saved = now
}
