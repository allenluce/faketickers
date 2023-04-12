# Controllable tickers for Go

    go get "github.com/allenluce/faketickers"

FakeTickers control time within your Go production code. The
FakeTicker control system gives you direct control over objects
produced by `time.NewTicker()` so you can test code without modifying
the code under test.

This is useful for when you want to make a ticker work faster than it
normally would in production code. And when you want to make sure your
test has things set up properly before a production goroutine ticks.

Consider this example of production code:

    func tickloop() {
    	ticker := time.NewTicker(time.Minute * 15)
    	for {
    		<-ticker
    		RunReportCode()
    	}
    }

Since you don't want to wait fifteen minutes for a test to run you
might rewrite `tickloop()` to accept a duration as an argument and
then pass in a shorter duration for `time.NewTicker()` (although now
you'll have to hope that testing finishes before the next tick
occurs). Another option is to pass in a test-controlled ticker object.
Or pass in an optional channel and select on both the ticker and
channel for testing.

Or you could avoid messing with your production code at all and and
instead use FakeTickers:

    import "github.com/allenluce/faketickers"

    ft := faketickers.NewFakeTicker()
    go tickloop()
    ft.Wait(1) // Make sure time.NewTicker() was called
    ft.Tick()

Now you can be assured that your RunReportCode() will be called
quickly and only once.

FakeTickers lets you make all future tickers tick as fast as possible:

    ft := faketickers.NewFakeTicker(true)

When done, stop the tickers to avoid hangs or panics:

    ft.Stop()

This turns off any fake tickers and restores the system ticker
facilities.

FakeTickers operates by replacing `time.NewTicker()` with its own
routine. This new routine hands out specially constructed ticker
objects that give you control over time.

Note that FakeTickers must be initialized before the call to
`time.NewTicker()`.  It will only hand out fake tickers after
initialization. Any previously created tickers will continue to work
as normal.

The `InstantSleeps()` facility of FakeTickers lets you force
`time.Sleep()` to take no time at all:

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
Start initializes the fake tickers and replaces `time.NewTicker()` with its own
routine.

#### func (*FakeTickers) Stop

```go
func (t *FakeTickers) Stop()
```
Stop closes all the existing ticker channels and restores `time.NewTicker()` to the
system default.

#### func (*FakeTickers) Tag

```go
func (t *FakeTickers) Tag(tag string)
```
Tag sets the given string tag on all subsequent `time.NewTicker()` calls. Set a tag
before a `time.NewTicker()` call when you want to control that ticker separately.

#### func (*FakeTickers) Tick

```go
func (t *FakeTickers) Tick(tag ...string)
```
Tick will send one tick to all tickers created since `Start()` was called. If
given the optional tag argument, it will only send ticks to those tickers
that have that tag.

#### func (*FakeTickers) Wait

```go
func (t *FakeTickers) Wait(minTickers int, timeoutInterval ...time.Duration) error
```
Wait blocks until the total number of calls to NewTicker is equal or
greater than `minTickers` or until `timeout` expires. Use this when
you don't want to proceed until the intended code has its ticker(s)
set up.
