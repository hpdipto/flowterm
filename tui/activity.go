package tui

import (
	"time"

	"hpdipto/flowterm/activity"
	"hpdipto/flowterm/db"
)

type span struct {
	App   string
	Name  string
	Phase db.Phase
	Start time.Time
	End   time.Time
}

type tracker struct {
	cur  *span
	runs []span
}

func (t *tracker) sample(now time.Time, s activity.Sample, phase db.Phase) {
	switch {
	case s.Locked || s.Idle > activity.IdleLimit:
		t.close(now, s)
	case t.cur != nil && t.cur.App == s.App && t.cur.Phase == phase:
		t.cur.End = now
		t.cur.Name = s.Name
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
			AppName:   r.Name,
			Phase:     r.Phase,
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
