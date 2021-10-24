package gpioutil

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpiotest"
)

func TestPulseIn_Success(t *testing.T) {
	var pin gpiotest.Pin

	done := make(chan struct{})

	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin.EdgesChan = edgesChan
	pin.L = gpio.Low

	go func() {
		edgesChan <- gpio.High
		edgesChan <- gpio.High

		clock.Advance(time.Second)

		edgesChan <- gpio.Low

		close(done)
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, -1, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	if duration != time.Second {
		t.Fatal("it should takes 1 second")
	}

	<-done
}

func TestPulseIn_Timeout_1(t *testing.T) {
	var pin gpiotest.Pin

	done := make(chan struct{})

	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin.EdgesChan = edgesChan
	pin.Clock = clock
	pin.L = gpio.Low

	go func() {
		clock.Advance(time.Second)
		close(done)
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, time.Second, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	if duration != 0 {
		t.Fatal("it should returns 0 for timeout")
	}

	<-done
}

func TestPulseIn_Timeout_2(t *testing.T) {
	var pin gpiotest.Pin

	done := make(chan struct{})

	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin.EdgesChan = edgesChan
	pin.L = gpio.Low

	go func() {
		edgesChan <- gpio.High
		edgesChan <- gpio.High

		clock.Advance(time.Second)

		close(done)
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, time.Second, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	if duration != 0 {
		t.Fatal("it should returns 0 for timeout")
	}

	<-done
}
