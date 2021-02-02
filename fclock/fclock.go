// Package fclock manages fcntl file locks.
//
// Note that there are generally 2 file locking facilities on UNIX. fcntl() and flock(). On some systems these 2 locking
// facilities are completely independent, and locks obtained by one have no impact on the other, where as on other
// systems they do interact. man 2 flock (https://linux.die.net/man/2/flock, Notes section) has a good explanation.
package fclock

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

type File interface {
	Fd() uintptr
}

type LockError struct {
	Pid int32
}

func (le LockError) Error() string {
	return fmt.Sprintf("conflicting lock (pid=%d)", le.Pid)
}
func IsLockError(err error) bool {
	_, ok := err.(LockError)
	return ok
}

// Lock attempts to place a lock on the given file.
func Lock(f File, blocking bool) error {
	lk := &syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Whence: 0,
		Start:  0,
		Len:    0,
	}

	var cmd int
	if blocking {
		cmd = syscall.F_SETLKW
	} else {
		cmd = syscall.F_SETLK
	}

	for {
		err := syscall.FcntlFlock(f.Fd(), cmd, lk)
		if err == nil {
			return nil
		}
		if err == syscall.EACCES || err == syscall.EAGAIN {
			return LockError{Pid: lk.Pid}
		}
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}

// Check determines whether a call to Lock() would succeed without actually obtaining a lock.
// Returns error if lock would fail.
func Check(f File) error {
	lk := &syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Whence: 0,
		Start:  0,
		Len:    0,
	}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_GETLK, lk); err != nil {
		return err
	}

	if lk.Type == syscall.F_UNLCK { return nil }
	if lk.Pid != 0 { return LockError{lk.Pid} }
	return nil
}

// CreateFile creates a new file and locks it atomically. There will be no period of time in which the file exists on
// the filesystem without a lock.
//
// One of either the os.O_RDWR or os.O_WRONLY flags must be present.
func CreateFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	flag |= unix.O_TMPFILE
	flag &^= os.O_CREATE
	f, err := os.OpenFile(filepath.Dir(path), flag, perm)
	if err != nil { return nil, err }

	if err := Lock(f, false); err != nil {
		f.Close()
		return nil, err
	}

	if err := unix.Linkat(int(f.Fd()), "", unix.AT_FDCWD, path, unix.AT_EMPTY_PATH); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}
