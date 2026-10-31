//go:build darwin

package activity

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>

extern CFDictionaryRef CGSessionCopyCurrentDictionary(void);

static double idle_seconds(void) {
	return CGEventSourceSecondsSinceLastEventType(kCGEventSourceStateCombinedSessionState, kCGAnyInputEventType);
}

static int screen_locked(void) {
	CFDictionaryRef d = CGSessionCopyCurrentDictionary();
	if (!d) return 0;
	CFBooleanRef v = CFDictionaryGetValue(d, CFSTR("CGSSessionScreenIsLocked"));
	int locked = v && CFBooleanGetValue(v);
	CFRelease(d);
	return locked;
}
*/
import "C"

import (
	"errors"
	"os/exec"
	"strings"
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
	app, err := frontApp()
	if err != nil {
		return Sample{}, err
	}

	return Sample{App: app, Idle: idleTime(), Locked: screenLocked()}, nil
}

func frontApp() (string, error) {
	asn, err := exec.Command("lsappinfo", "front").Output()
	if err != nil {
		return "", err
	}

	info, err := exec.Command("lsappinfo", "info", "-only", "bundleid", strings.TrimSpace(string(asn))).Output()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(info), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "bundleID="); ok {
			return strings.Trim(v, `"`), nil
		}
	}

	return "", errors.New("activity: no frontmost application")
}

func idleTime() time.Duration {
	return time.Duration(float64(C.idle_seconds()) * float64(time.Second))
}

func screenLocked() bool {
	return C.screen_locked() != 0
}
