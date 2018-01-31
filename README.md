# Controllable tickers for Go

    go get "github.com/allenluce/faketickers"

Package faketickers provides a way to control time.Ticker objects from test code
without requiring modifications of the code under test.

This is super-handy when you want to make a ticker work faster than it normally
would in production code or when you want to make sure your test has things set
up properly before a production goroutine ticks.

Say you have this bit of production code:

    func tickloop() {
    	ticker := time.NewTicker(time.Minute * 15)
    	for {
    		<-ticker
    		RunReportCode()
    	}
    }

You could test this by rewriting tickloop() to take a a shorter duration as an
argument. You'd then have to hope that testing finishes before the next tick. Or
you could have it take a ticker object as an argument and control that from your
tests. Or you could have it take an optional channel and select on both the
ticker and channel for testing.

Or you could skip rewriting your production code and use FakeTickers:

    import "github.com/allenluce/faketickers"

    ft := faketickers.NewFakeTicker()
    go tickloop()
    ft.Wait(1) // Make sure time.NewTicker() was called
    ft.Tick()

Now you can be assured that your RunReportCode() will be called quickly and only
once.

You can also set all tickers to tick as fast as possible:

    ft := faketickers.NewFakeTicker(true)

When done, it's best to stop the tickers (to avoid hangs/panics):

    ft.Stop()

This turns off any fake tickers and restores the system ticker
facilities.

FakeTickers operates by replacing time.NewTicker() with its own routine. This
new routine hands out specially constructed ticker objects that give you control
over time.

Note that FakeTickers must be initialized before the call to time.NewTicker() --
it'll only hand out fake tickers then, not replace existing tickers.

You can also make `time.Sleep()` calls of any duration take no time at all:

    p := faketickers.InstantSleeps()
    time.Sleep(time.Hour * 876000) // Instant.
    p.Stop()
    time.Sleep(time.Hour * 876000) // Waits 100 years.

## Usage

#### type FakeTickers

```go
type FakeTickers struct {
}
```


#### func (*FakeTickers) Start

```go
func (t *FakeTickers) Start()
```
Start initializes the fake tickers and replaces time.NewTicker() with its own
routine.

#### func (*FakeTickers) Stop

```go
func (t *FakeTickers) Stop()
```
Stop closes all the existing ticker channels and restores time.NewTicker to the
system default.

#### func (*FakeTickers) Tag

```go
func (t *FakeTickers) Tag(tag string)
```
Tag sets the given string tag on all subsequent NewTicker() calls. Set a tag
before a NewTicker() call when you want to control that ticker separately.

#### func (*FakeTickers) Tick

```go
func (t *FakeTickers) Tick(tag ...string)
```
Tick will send one tick down all tickers created since Start() was called. If
given the optional tag argument, it will only send ticks to those NewTickers
that have that tag.

#### func (*FakeTickers) Wait

```go
func (t *FakeTickers) Wait(minTickers int, timeoutInterval ...time.Duration) error
```
Wait blocks until the total number of calls to NewTicker is equal or greater
than minTickers or until timeout. Use when you don't want to proceed until the
intended code has its ticker(s) set up.
