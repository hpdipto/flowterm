//go:build unix

package main

import (
	"errors"
	"fmt"
	"hpdipto/flowterm/utils"
	"os"
	"path/filepath"
	"syscall"
)

// acquireLock locks a file next to the database so only one flowterm runs;
// the OS frees the lock when the process dies, so a crash never leaves it stuck.
func acquireLock(dbPath string) (*os.File, error) {
	path := filepath.Join(filepath.Dir(dbPath), "."+filepath.Base(dbPath)+".lock")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		utils.LogDeferredError(f.Close)
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, fmt.Errorf("flowterm is already running")
		}

		return nil, err
	}

	return f, nil
}
