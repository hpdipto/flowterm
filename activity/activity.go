//go:build darwin

package activity

/*
#cgo LDFLAGS: -framework AppKit -framework CoreGraphics
#include <stdlib.h>
#include <string.h>
#include <objc/message.h>
#include <objc/runtime.h>
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>

extern CFDictionaryRef CGSessionCopyCurrentDictionary(void);

static char* front_bundle_id(void) {
	id ws = ((id (*)(id, SEL))objc_msgSend)((id)objc_getClass("NSWorkspace"), sel_registerName("sharedWorkspace"));
	id app = ((id (*)(id, SEL))objc_msgSend)(ws, sel_registerName("frontmostApplication"));
	if (!app) return NULL;
	id bid = ((id (*)(id, SEL))objc_msgSend)(app, sel_registerName("bundleIdentifier"));
	if (!bid) return NULL;
	const char *s = ((const char* (*)(id, SEL))objc_msgSend)(bid, sel_registerName("UTF8String"));
	return s ? strdup(s) : NULL;
}

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
	"time"
	"unsafe"
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
	p := C.front_bundle_id()
	if p == nil {
		return "", errors.New("activity: no frontmost application")
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p), nil
}

func idleTime() time.Duration {
	return time.Duration(float64(C.idle_seconds()) * float64(time.Second))
}

func screenLocked() bool {
	return C.screen_locked() != 0
}
