package ticker

import (
	"sync"
	"time"
)

type Ticker struct {
	duration time.Duration
	nextTick time.Time
	timer    *time.Timer
	mutex    sync.Mutex

	f func()

	C <-chan struct{}
	c chan struct{}
}

func NewTicker(d time.Duration) *Ticker {
	if d == 0 {
		panic("Ticker duration cannot be 0")
	}
	chn := make(chan struct{}, 1)
	return &Ticker{
		duration: d,
		C:        chn,
		c:        chn,
	}
}

func NewTickerFunc(d time.Duration, f func()) *Ticker {
	return &Ticker{
		duration: d,
		f:        f,
	}
}

func (t *Ticker) Stop() {
	t.mutex.Lock()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	t.mutex.Unlock()
}

// Starts the ticker.
// The first tick will be after the interval has passed.
func (t *Ticker) Start() {
	t.mutex.Lock()
	if t.timer == nil {
		t.reset()
	}
	t.mutex.Unlock()
}

// StartNow starts the ticker and triggers an immediate tick.
func (t *Ticker) StartNow() {
	t.Start()
	t.send()
}

// ResetNow triggers an immediate tick and resets the timer so the next tick occurs after the configured interval.
func (t *Ticker) ResetNow() {
	t.Reset()
	t.send()
}

// Reset resets the timer so the next tick occurs after the configured interval.
func (t *Ticker) Reset() {
	t.mutex.Lock()
	t.reset()
	t.mutex.Unlock()
}

func (t *Ticker) reset() {
	if t.timer != nil {
		t.timer.Stop()
	}
	t.nextTick = time.Now().Add(t.duration)
	t.timer = time.AfterFunc(t.duration, t.tick)
}

func (t *Ticker) send() {
	if t.f != nil {
		go t.f()
	} else {
		select {
		case t.c <- struct{}{}:
		default:
		}
	}
}

func (t *Ticker) tick() {
	now := time.Now()
	t.send()

	t.mutex.Lock()
	if t.timer == nil {
		// we were stopped in between the tick firing and getting a lock.
		t.mutex.Unlock()
		return
	}

	if now.Before(t.nextTick) {
		// timer was reset and we're an orphaned tick
		t.mutex.Unlock()
		return
	}

	// This should do nothing since we're in the fire function, and are already stopped. However if the timer got reset, there may be a race condition where the first timer fired (which would be the current execution of tick()), but while waiting for a lock a new timer was created.
	t.timer.Stop()

	t.nextTick = t.nextTick.Add(t.duration)
	tDiff := t.nextTick.Sub(time.Now())
	if tDiff < 0 {
		// We missed the next scheduled tick. Reschedule for the future.
		tDiff = t.duration + (tDiff % t.duration)
	}
	t.timer = time.AfterFunc(tDiff, t.tick)
	t.mutex.Unlock()
}
