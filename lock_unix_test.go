//go:build unix

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAcquireLockSecondCallFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowterm.db")

	first, err := acquireLock(path)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	if _, err := acquireLock(path); err == nil {
		t.Fatal("second acquire succeeded, want refusal")
	} else if !strings.Contains(err.Error(), "already running") {
		t.Fatalf("second acquire error = %v, want already running", err)
	}

	first.Close()

	again, err := acquireLock(path)
	if err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	again.Close()
}

func TestAcquireLockDifferentDatabases(t *testing.T) {
	dir := t.TempDir()

	first, err := acquireLock(filepath.Join(dir, "one.db"))
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer first.Close()

	second, err := acquireLock(filepath.Join(dir, "two.db"))
	if err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	second.Close()
}

// TestAcquireLockReleasedOnKill holds the lock in a child process, checks a
// second process is refused, then kills the child without any cleanup. The
// lock must be free afterwards: that is the crash-safety promise.
func TestAcquireLockReleasedOnKill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flowterm.db")

	cmd := exec.Command(os.Args[0], "-test.run=^TestLockHolderChild$")
	cmd.Env = append(os.Environ(), "FLOWTERM_LOCK_CHILD="+path)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}

	locked := make(chan string, 1)
	go func() {
		line, _ := bufio.NewReader(stdout).ReadString('\n')
		locked <- line
	}()

	select {
	case line := <-locked:
		if strings.TrimSpace(line) != "locked" {
			cmd.Process.Kill()
			cmd.Wait()
			t.Fatalf("child output = %q, want locked", line)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal("child never reported the lock")
	}

	if _, err := acquireLock(path); err == nil {
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal("acquired lock while child held it")
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("kill child: %v", err)
	}
	cmd.Wait()

	after, err := acquireLock(path)
	if err != nil {
		t.Fatalf("acquire after kill: %v", err)
	}
	after.Close()
}

// TestLockHolderChild is the helper process for TestAcquireLockReleasedOnKill.
// It reports readiness and waits to be killed.
func TestLockHolderChild(t *testing.T) {
	path := os.Getenv("FLOWTERM_LOCK_CHILD")
	if path == "" {
		t.Skip("helper process only")
	}

	lock, err := acquireLock(path)
	if err != nil {
		t.Fatalf("child acquire: %v", err)
	}

	fmt.Println("locked")
	os.Stdout.Sync()

	time.Sleep(time.Hour)

	lock.Close()
}
