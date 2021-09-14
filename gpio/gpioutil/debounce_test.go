// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpiotest"
)

func TestDebounce_Err(t *testing.T) {
	f := gpiotest.Pin{}
	if _, err := Debounce(&f, time.Second, 0, gpio.BothEdges); err == nil {
		t.Fatal("expected error")
	}
}

func TestDebounce_Zero(t *testing.T) {
	f := gpiotest.Pin{}
	p, err := Debounce(&f, 0, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal("expected error")
	}
	if p1, ok := p.(*gpiotest.Pin); !ok || p1 != &f {
		t.Fatal("expected the pin to be returned as-is")
	}
}

func TestDebounce_In(t *testing.T) {
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		t.Fatal(err)
	}
	if p.Halt() != nil {
		t.Fatal(err)
	}
}

func TestDebounce_Read_Low(t *testing.T) {
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, time.Second, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if p.Read() != gpio.Low {
		t.Fatal("expected level")
	}
	if p.Read() != gpio.Low {
		t.Fatal("expected level")
	}
}

func TestDebounce_Read_High(t *testing.T) {
	f := gpiotest.Pin{L: gpio.High, EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, time.Second, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if p.Read() != gpio.High {
		t.Fatal("expected level")
	}
	if p.Read() != gpio.High {
		t.Fatal("expected level")
	}
}

func TestDebounce_WaitForEdge_Got(t *testing.T) {
	fakeClock := clockwork.NewFakeClock()
	f := gpiotest.Pin{
		Clock:     fakeClock,
		EdgesChan: make(chan gpio.Level, 1),
	}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	p.(*debounced).clock = fakeClock
	f.Out(gpio.High)
	f.EdgesChan <- gpio.Low
	go func() {
		// Sleepers:
		// * debounce.WaitForEdge's d.Clock.Sleep
		//
		// gpiotest.WaitForEdge doesn't call sleep due to infinite timeout.
		const numSleepers = 1

		fakeClock.BlockUntil(numSleepers)
		fakeClock.Advance(2 * time.Second)
	}()
	if !p.WaitForEdge(-1) {
		t.Fatal("expected edge")
	}
}

func TestDebounce_WaitForEdge_Noise_NoEdge(t *testing.T) {
	fakeClock := clockwork.NewFakeClock()
	f := gpiotest.Pin{
		Clock:     fakeClock,
		EdgesChan: make(chan gpio.Level, 1),
	}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	p.(*debounced).clock = fakeClock
	f.Out(gpio.Low)

	go func() {
		// Sleepers:
		// * gpiotest.WaitForEdge's After
		// * debounce.WaitForEdge's d.Clock.Sleep
		const numSleepers = 2

		// Short high, comes back down too soon
		f.EdgesChan <- gpio.High
		fakeClock.BlockUntil(numSleepers)
		fakeClock.Advance(100 * time.Millisecond)
		f.Out(gpio.Low)
		fakeClock.Advance(2 * time.Second)
	}()

	if p.WaitForEdge(1 * time.Second) {
		t.Fatal("expected no edge")
	}
}

func TestDebounce_WaitForEdge_Noise_Edge(t *testing.T) {
	fakeClock := clockwork.NewFakeClock()

	f := gpiotest.Pin{
		Clock:     fakeClock,
		EdgesChan: make(chan gpio.Level, 2),
	}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	p.(*debounced).clock = fakeClock
	f.Out(gpio.Low)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		// Sleepers:
		// * gpiotest.WaitForEdge's After
		// * debounce.WaitForEdge's d.Clock.Sleep
		const numSleepers = 2

		// 100ms high (too short)
		f.EdgesChan <- gpio.High
		fakeClock.BlockUntil(numSleepers)
		f.Out(gpio.Low)
		fakeClock.Advance(100 * time.Millisecond)

		// stays high indefinitely (long enough)
		fakeClock.BlockUntil(numSleepers)
		f.Out(gpio.High)
		fakeClock.Advance(2 * time.Second)
	}()

	if !p.WaitForEdge(4 * time.Second) {
		t.Fatal("expected edge")
	}
}

func TestDebounce_WaitForEdge_Timeout(t *testing.T) {
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if p.WaitForEdge(0) {
		t.Fatal("expected no edge")
	}
}

func TestDebounce_RealPin(t *testing.T) {
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := p.(gpio.RealPin)
	if !ok {
		t.Fatal("expected gpio.RealPin")
	}
	a, ok := r.Real().(*gpiotest.Pin)
	if !ok {
		t.Fatal("expected gpiotest.Pin")
	}
	if a != &f {
		t.Fatal("expected actual pin")
	}
}

func TestDebounce_RealPin_Deep(t *testing.T) {
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Debounce(p, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := p.(gpio.RealPin)
	if !ok {
		t.Fatal("expected gpio.RealPin")
	}
	a, ok := r.Real().(*gpiotest.Pin)
	if !ok {
		t.Fatal("expected gpiotest.Pin")
	}
	if a != &f {
		t.Fatal("expected actual pin")
	}
}
