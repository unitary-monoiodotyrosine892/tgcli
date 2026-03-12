package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Lock represents a file lock on the store directory.
type Lock struct {
	file *os.File
	path string
}

// Acquire attempts to acquire an exclusive lock on the store directory.
func Acquire(storeDir string) (*Lock, error) {
	// Create store directory with restricted permissions (owner only)
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	lockPath := filepath.Join(storeDir, "LOCK")
	
	// Create lock file with restricted permissions (owner read/write only)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	// Ensure correct permissions even if file existed
	if err := os.Chmod(lockPath, 0600); err != nil {
		f.Close()
		return nil, fmt.Errorf("set lock file permissions: %w", err)
	}

	// Try to acquire exclusive lock (non-blocking)
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another tgcli instance is running (lock held on %s)", lockPath)
	}

	return &Lock{file: f, path: lockPath}, nil
}

// Release releases the lock.
func (l *Lock) Release() error {
	if l.file == nil {
		return nil
	}
	defer func() {
		l.file = nil
	}()

	// Unlock
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		l.file.Close()
		return err
	}

	return l.file.Close()
}
