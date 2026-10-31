//go:build !darwin

package activity

import (
	"errors"
	"time"
)

const IdleLimit = 60 * time.Second

type Run struct {
	App   string
	Start time.Time
	End   time.Time
}

func (r Run) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

type Sample struct {
	App    string
	Idle   time.Duration
	Locked bool
}

func Current() (Sample, error) {
	return Sample{}, errors.New("activity: only supported on macOS")
}
