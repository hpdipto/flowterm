//go:build !darwin

package activity

import (
	"errors"
	"time"
)

const IdleLimit = 60 * time.Second

type Sample struct {
	App    string
	Name   string
	Idle   time.Duration
	Locked bool
}

func Current() (Sample, error) {
	return Sample{}, errors.New("activity: only supported on macOS")
}
