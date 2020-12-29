// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package driverreg

import (
	"errors"
	"testing"

	"periph.io/x/conn/driver"
)

func TestInitSimple(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	})
	if len(byName) != 1 {
		t.Fatal(byName)
	}
	s, err := Init()
	if err != nil || len(s.Loaded) != 1 {
		t.Fatal(s, err)
	}

	// Call a second time, should return the same data.
	s2, err2 := Init()
	if err2 != nil || len(s2.Loaded) != len(s.Loaded) || s2.Loaded[0] != s.Loaded[0] {
		t.Fatal(s2, err2)
	}
}

func TestInitSkip(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: nil,
			ok:      false,
			err:     nil,
		},
	})
	s, err := Init()
	if err != nil || len(s.Skipped) != 1 {
		t.Fatal(s, err)
	}
	if s := s.Skipped[0].String(); s != "CPU: <nil>" {
		t.Fatal(s)
	}
}

func TestInitErr(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     errors.New("oops"),
		},
	})
	s, err := Init()
	if err != nil || len(s.Loaded) != 0 || len(s.Failed) != 1 {
		t.Fatal(s, err)
	}
	if s := s.Failed[0].String(); s != "CPU: oops" {
		t.Fatal(s)
	}
}

func TestInitPrerequisitesCircular(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: []string{"Board"},
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "Board",
			prereqs: []string{"CPU"},
			ok:      true,
			err:     nil,
		},
	})
	s, err := Init()
	if err == nil || len(s.Loaded) != 0 {
		t.Fatal(s, err)
	}
	if err.Error() != "periph: found cycle(s) in drivers dependencies:\nBoard: CPU\nCPU: Board" {
		t.Fatal(err)
	}
}

func TestInitPrerequisitesMissing(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: []string{"Board"},
			ok:      true,
			err:     nil,
		},
	})
	s, err := Init()
	if err == nil || len(s.Loaded) != 0 {
		t.Fatal(s, err)
	}
}

func TestInitAfterMissing(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:  "CPU",
			after: []string{"Board"},
			ok:    true,
			err:   nil,
		},
	})
	s, err := Init()
	if err != nil || len(s.Loaded) != 1 {
		t.Fatal(s, err)
	}
}

func TestDependencySkipped(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: nil,
			ok:      false,
			err:     errors.New("skipped"),
		},
		&driverFake{
			name:    "Board",
			prereqs: []string{"CPU"},
			ok:      true,
			err:     nil,
		},
	})
	s, err := Init()
	if err != nil || len(s.Skipped) != 2 {
		t.Fatal(s, err)
	}
}

func TestRegisterLate(t *testing.T) {
	defer reset()
	reset()
	if _, err := Init(); err != nil {
		t.Fatal(err)
	}
	d := &driverFake{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if Register(d) == nil {
		t.Fatal("can't register after Init()")
	}
}

func TestRegisterTwice(t *testing.T) {
	defer reset()
	reset()
	d := &driverFake{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err != nil {
		t.Fatal(err)
	}
	if Register(d) == nil {
		t.Fatal("can't register twice")
	}
}

func TestMustRegisterPanic(t *testing.T) {
	defer reset()
	reset()
	d := &driverFake{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err != nil {
		t.Fatal(err)
	}
	panicked := false
	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()
	MustRegister(d)
	if !panicked {
		t.Fatal("MustRegister() should have panicked on driver registration failure")
	}
}

func TestPrerequisitesExplodeStagesSimple(t *testing.T) {
	defer reset()
	reset()
	d := []driver.Impl{
		&driverFake{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 1 || len(actual[0].drvs) != 1 {
		t.Fatal(actual)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrerequisitesExplodeStages1Dep(t *testing.T) {
	defer reset()
	reset()
	// This explodes the stage into two.
	d := []driver.Impl{
		&driverFake{
			name:    "CPU-specialized",
			prereqs: []string{"CPU-generic"},
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "CPU-generic",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 2 || len(actual[0].drvs) != 1 || actual[0].drvs["CPU-generic"] != d[1] || len(actual[1].drvs) != 1 || actual[1].drvs["CPU-specialized"] != d[0] || err != nil {
		t.Fatal(actual, err)
	}
}

func TestPrerequisitesExplodeStagesCycle(t *testing.T) {
	defer reset()
	reset()
	d := []driver.Impl{
		&driverFake{
			name:    "A",
			prereqs: []string{"B"},
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "B",
			prereqs: []string{"C"},
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "C",
			prereqs: []string{"A"},
			ok:      true,
			err:     nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 0 {
		t.Fatal(actual)
	}
	if err == nil {
		t.Fatal("cycle should have been detected")
	}
}

func TestPrerequisitesExplodeStages3Dep(t *testing.T) {
	defer reset()
	reset()
	// This explodes the stage into 3 due to diamond shaped DAG.
	d := []driver.Impl{
		&driverFake{
			name:    "base2",
			prereqs: []string{"root"},
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "base1",
			prereqs: []string{"root"},
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "root",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:    "super",
			prereqs: []string{"base1", "base2"},
			ok:      true,
			err:     nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 3 || len(actual[0].drvs) != 1 || len(actual[1].drvs) != 2 || len(actual[2].drvs) != 1 {
		t.Fatal(actual)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestAfterExplodeStages3Dep(t *testing.T) {
	defer reset()
	reset()
	// This explodes the stage into 3 due to diamond shaped DAG.
	d := []driver.Impl{
		&driverFake{
			name:  "base2",
			after: []string{"root"},
			ok:    true,
			err:   nil,
		},
		&driverFake{
			name:  "base1",
			after: []string{"root"},
			ok:    true,
			err:   nil,
		},
		&driverFake{
			name:    "root",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
		&driverFake{
			name:  "super",
			after: []string{"base1", "base2"},
			ok:    true,
			err:   nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 3 || len(actual[0].drvs) != 1 || len(actual[1].drvs) != 2 || len(actual[2].drvs) != 1 {
		t.Fatal(actual)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestDrivers(t *testing.T) {
	var d []driver.Impl
	d = insertDriver(d, &driverFake{name: "b"})
	d = insertDriver(d, &driverFake{name: "d"})
	d = insertDriver(d, &driverFake{name: "c"})
	d = insertDriver(d, &driverFake{name: "a"})
	for i, l := range []string{"a", "b", "c", "d"} {
		if d[i].String() != l {
			t.Fatal(d)
		}
	}
}

func TestFailures(t *testing.T) {
	var d []DriverFailure
	d = insertDriverFailure(d, DriverFailure{D: &driverFake{name: "b"}})
	d = insertDriverFailure(d, DriverFailure{D: &driverFake{name: "d"}})
	d = insertDriverFailure(d, DriverFailure{D: &driverFake{name: "c"}})
	d = insertDriverFailure(d, DriverFailure{D: &driverFake{name: "a"}})
	for i, l := range []string{"a", "b", "c", "d"} {
		if d[i].D.String() != l {
			t.Fatal(d)
		}
	}
}

//

func reset() {
	byName = map[string]driver.Impl{}
	state = nil
}

func registerDrivers(drivers []driver.Impl) {
	for _, d := range drivers {
		MustRegister(d)
	}
}

type driverFake struct {
	name    string
	prereqs []string
	after   []string
	ok      bool
	err     error
}

func (d *driverFake) String() string {
	return d.name
}

func (d *driverFake) Prerequisites() []string {
	return d.prereqs
}

func (d *driverFake) After() []string {
	return d.after
}

func (d *driverFake) Init() (bool, error) {
	return d.ok, d.err
}
