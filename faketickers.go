/*
Package faketickers provides a way to control time.Ticker objects
from test code without requiring modifications of the code under test.

This is super-handy when you want to make a ticker work faster than it
normally would in production code or when you want to make sure your
test has things set up properly before a production goroutine ticks.

Say you have this bit of production code:

	func tickloop() {
		ticker := time.NewTicker(time.Minute * 15)
		for {
			<-ticker
			RunReportCode()
		}
	}

You could test this by rewriting tickloop() to take a a shorter
duration as an argument. You'd then have to hope that testing finishes
before the next tick. Or you could have it take a ticker object as an
argument and control that from your tests. Or you could have it take
an optional channel and select on both the ticker and channel for
testing.

Or you could skip rewriting your production code and use FakeTickers:

	ft := faketickers.FakeTickers{}
	ft.Start()
	go tickloop()
	ft.Wait(1) // Make sure time.NewTicker() was called
	ft.Tick()

Now you can be assured that your RunReportCode() will be called
quickly and only once.

FakeTickers operates by replacing time.NewTicker() with its own
routine. This new routine hands out specially constructed ticker
objects that give you control over time.

Note that FakeTickers must be initialized before the call to
time.NewTicker() -- it'll only hand out fake tickers then, not replace
existing tickers.

*/
package faketickers

import (
	"fmt"
	"sync"
	"time"

	"bou.ke/monkey"
)

type taggedTicker struct {
	Tag    string
	Ticker *chan time.Time
}

// FakeTickers provides a flexible ticker system.
type FakeTickers struct {
	tickers   []taggedTicker
	guard     *monkey.PatchGuard
	tag       string
	tickerMut sync.Mutex
	immediate bool
	done      chan interface{}
	wg        *sync.WaitGroup
}

// NewFakeTicker creates and starts the fake tickers
func NewFakeTicker(immediate ...bool) *FakeTickers {
	ft := &FakeTickers{}
	if len(immediate) > 0 {
		ft.immediate = immediate[0]
	}
	ft.Start()
	return ft
}

func (t *FakeTickers) newTicker(d time.Duration) *time.Ticker {
	defer t.tickerMut.Unlock()
	t.tickerMut.Lock()
	var ticker time.Ticker
	c := make(chan time.Time)
	ticker.C = c
	t.tickers = append(t.tickers, taggedTicker{Tag: t.tag, Ticker: &c})
	if t.immediate { // Tick quickly and forever.
		t.wg.Add(1)
		go func(c chan time.Time) {
			for {
				now := time.Now()
				select {
				case <-t.done: // Exit when closed.
					t.wg.Done()
					return
				case c <- now: // Endless ticks
				}
			}
		}(c)
	}
	return &ticker
}

// Start initializes the fake tickers and replaces time.NewTicker()
// with its own routine.
func (t *FakeTickers) Start() {
	t.done = make(chan interface{})
	t.wg = &sync.WaitGroup{}
	t.guard = monkey.Patch(time.NewTicker, t.newTicker)
	t.tickers = []taggedTicker{}
}

// Tick will send one tick down all tickers created since Start() was
// called. If given the optional tag argument, it will only send ticks
// to those NewTickers that have that tag.
func (t *FakeTickers) Tick(tag ...string) {
	if len(tag) == 0 {
		tag = []string{""}
	}
	now := time.Now()
	for _, ticker := range t.tickers {
		if ticker.Tag == tag[0] {
			*ticker.Ticker <- now
		}
	}
}

// Tag sets the given string tag on all subsequent NewTicker()
// calls. Set a tag before a NewTicker() call when you want to control
// that ticker separately.
func (t *FakeTickers) Tag(tag string) {
	t.tag = tag
}

// Stop closes all the existing ticker channels and restores
// time.NewTicker to the system default.
func (t *FakeTickers) Stop() {
	close(t.done)
	t.wg.Wait()
	t.guard.Unpatch()
	for _, ticker := range t.tickers {
		close(*ticker.Ticker)
	}
}

const pollingInterval = time.Millisecond * 10

// Wait blocks until the total number of calls to NewTicker is equal
// or greater than minTickers or until timeout.  Use when you don't want to
// proceed until the intended code has its ticker(s) set up.
func (t *FakeTickers) Wait(minTickers int, timeoutInterval ...time.Duration) error {
	var timeout <-chan time.Time
	if len(timeoutInterval) > 0 {
		timeout = time.After(timeoutInterval[0])
	}
	for {
		select {
		case <-time.After(pollingInterval):
			if len(t.tickers) >= minTickers {
				return nil
			}
		case <-timeout:
			return fmt.Errorf("Timeout, only %d NewTicker calls (not %d)", len(t.tickers), minTickers)
		}
	}
}

// Sleeper stores the state of the old time.Sleep call for restoration
// with Stop()
type Sleeper struct {
	guard *monkey.PatchGuard
}

// InstantSleeps makes time.Sleep() return instantly, no matter what
// the duration.
func InstantSleeps() *Sleeper {
	s := &Sleeper{}
	s.guard = monkey.Patch(time.Sleep, func(time.Duration) {})
	return s
}

// Stop restores the normal functioning of time.Sleep
func (s *Sleeper) Stop() {
	s.guard.Unpatch()
}
