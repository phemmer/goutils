package watchdog

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sync/atomic"
	"time"
)

type WatchDog struct {
	expiry   time.Duration
	lastPing atomic.Value
}

func NewWatchDog(expiry time.Duration) *WatchDog {
	return &WatchDog{
		expiry: expiry,
	}
}

func (wd *WatchDog) Start() {
	wd.Ping()
	go func() {
		ticker := time.NewTicker(time.Second)
		for t := range ticker.C {
			if t.Sub(wd.lastPing.Load().(time.Time)) > wd.expiry {
				break
			}
		}

		fmt.Fprintf(os.Stderr, "WatchDog triggered!\n")
		//debug.PrintStack()
		pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
		os.Exit(1)
	}()
}

func (wd *WatchDog) Ping() {
	wd.lastPing.Store(time.Now())
}
