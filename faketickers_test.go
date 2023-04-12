package faketickers_test

import (
	"sync"
	"testing"
	"time"

	. "github.com/allenluce/faketickers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func tickingRoutine(done chan<- bool) bool {
	ticker := time.NewTicker(time.Hour)
	_, ok := <-ticker.C
	done <- ok
	return ok
}

var _ = Describe("Faketicker", func() {
	It("ticks", func() {
		ft := FakeTickers{}
		ft.Start()
		done := make(chan bool, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			Ω(tickingRoutine(done)).Should(BeTrue())
		}()
		Ω(ft.Wait(1)).Should(Succeed())
		ft.Tick()
		Eventually(done).Should(Receive(BeTrue()))
		wg.Wait()
	})
	It("closes tickers", func() {
		ft := FakeTickers{}
		ft.Start()
		done := make(chan bool, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			Ω(tickingRoutine(done)).Should(BeFalse())
		}()
		Ω(ft.Wait(1)).Should(Succeed())
		ft.Stop()
		Eventually(done).Should(Receive(BeFalse()))
		wg.Wait()
	})
	It("ticks tags", func() {
		ft := FakeTickers{}
		ft.Start()

		done1 := make(chan bool, 1)
		ft.Tag("first")
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			Ω(tickingRoutine(done1)).Should(BeTrue())
		}()
		// Make sure the first created its ticker
		Ω(ft.Wait(1)).Should(Succeed())

		done2 := make(chan bool, 1)
		ft.Tag("second")
		wg.Add(1)
		go func() {
			defer wg.Done()
			Ω(tickingRoutine(done2)).Should(BeTrue())
		}()

		// Wait for second to make its ticker
		Ω(ft.Wait(2)).Should(Succeed())
		// Make sure they are still running
		Consistently(done1).ShouldNot(Receive())
		Consistently(done2).ShouldNot(Receive())

		// Tick the first one.
		ft.Tick("first")
		Eventually(done1).Should(Receive(BeTrue())) // Should tick.
		Consistently(done2).ShouldNot(Receive())    // Should not.

		// Tick the second one.
		ft.Tick("second")
		Eventually(done2).Should(Receive(BeTrue())) // Should tick.
		// Make sure they both exit properly
		wg.Wait()
	})
	It("waits until timeout", func() {
		ft := FakeTickers{}
		ft.Start()
		done := make(chan bool, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			Ω(tickingRoutine(done)).Should(BeTrue())
		}()
		Ω(ft.Wait(2, time.Millisecond*100)).ShouldNot(Succeed())
		Ω(ft.Wait(1)).Should(Succeed())
		ft.Tick()
		Eventually(done).Should(Receive(BeTrue()))
		wg.Wait()
	})
	It("ticks immediately with an immediate argument", func() {
		ft := NewFakeTicker(true)
		defer ft.Stop()
		done := make(chan bool, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			Ω(tickingRoutine(done)).Should(BeTrue())
		}()
		Eventually(done).Should(Receive(BeTrue()))
		wg.Wait()
	})
})

var _ = Describe("InstantSleeps", func() {
	It("forces a long time.Sleep to take no time at all.", func() {
		start := time.Now()
		p := InstantSleeps()
		time.Sleep(time.Hour)
		p.Stop()
		Ω(time.Now()).Should(BeTemporally("~", start, time.Minute*10))
	})
})

// Ginkgo boilerplate, this runs all tests in this package
func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FakeTicker Tests")
}
