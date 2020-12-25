// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package onewirereg defines a registry for onewire buses present on the host.
package onewirereg

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"periph.io/x/conn/onewire"
)

// Opener opens an handle to a bus.
//
// It is provided by the actual bus driver.
type Opener func() (onewire.BusCloser, error)

// Ref references an 1-wire bus.
//
// It is returned by All() to enumerate all registered buses.
type Ref struct {
	// Name of the bus.
	//
	// It must not be a sole number. It must be unique across the host.
	Name string
	// Aliases are the alternative names that can be used to reference this bus.
	Aliases []string
	// Number of the bus or -1 if the bus doesn't have any "native" number.
	//
	// Buses provided by the CPU normally have a 0 based number. Buses provided
	// via an addon (like over USB) generally are not numbered.
	Number int
	// Open is the factory to open an handle to this 1-wire bus.
	Open Opener
}

// Open opens an 1-wire bus by its name, an alias or its number and returns an
// handle to it.
//
// Specify the empty string "" to get the first available bus. This is the
// recommended default value unless an application knows the exact bus to use.
//
// Each bus can register multiple aliases, each leading to the same bus handle.
//
// "Bus number" is a generic concept that is highly dependent on the platform
// and OS. On some platform, the first bus may have the number 0, 1 or higher.
// Bus numbers are not necessarily continuous and may not start at 0. It was
// observed that the bus number as reported by the OS may change across OS
// revisions.
//
// When the 1-wire bus is provided by an off board plug and play bus like USB
// via a FT232H USB device, there can be no associated number.
func Open(name string) (onewire.BusCloser, error) {
	var r *Ref
	var err error
	func() {
		mu.Lock()
		defer mu.Unlock()
		if len(byName) == 0 {
			err = errors.New("onewirereg: no bus found; did you forget to call Init()?")
			return
		}
		if len(name) == 0 {
			r = getDefault()
			return
		}
		// Try by name, by alias, by number.
		if r = byName[name]; r == nil {
			if r = byAlias[name]; r == nil {
				if i, err2 := strconv.Atoi(name); err2 == nil {
					r = byNumber[i]
				}
			}
		}
	}()
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New("onewirereg: can't open unknown bus: " + strconv.Quote(name))
	}
	return r.Open()
}

// All returns a copy of all the registered references to all know 1-wire buses
// available on this host.
//
// The list is sorted by the bus name.
func All() []*Ref {
	mu.Lock()
	defer mu.Unlock()
	out := make([]*Ref, 0, len(byName))
	for _, v := range byName {
		r := &Ref{Name: v.Name, Aliases: make([]string, len(v.Aliases)), Number: v.Number, Open: v.Open}
		copy(r.Aliases, v.Aliases)
		out = insertRef(out, r)
	}
	return out
}

// Register registers an 1-wire bus.
//
// Registering the same bus name twice is an error, e.g. o.Name(). o.Number()
// can be -1 to signify that the bus doesn't have an inherent "bus number". A
// good example is a bus provided over a FT232H device connected on an USB bus.
// In this case, the bus name should be created from the serial number of the
// device for unique identification.
func Register(name string, aliases []string, number int, o Opener) error {
	if len(name) == 0 {
		return errors.New("onewirereg: can't register a bus with no name")
	}
	if o == nil {
		return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with nil Opener")
	}
	if number < -1 {
		return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with invalid bus number " + strconv.Itoa(number))
	}
	if _, err := strconv.Atoi(name); err == nil {
		return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with name being only a number")
	}
	if strings.Contains(name, ":") {
		return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with name containing ':'")
	}
	for _, alias := range aliases {
		if len(alias) == 0 {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with an empty alias")
		}
		if name == alias {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with an alias the same as the bus name")
		}
		if _, err := strconv.Atoi(alias); err == nil {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with an alias that is a number: " + strconv.Quote(alias))
		}
		if strings.Contains(alias, ":") {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " with an alias containing ':': " + strconv.Quote(alias))
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " twice")
	}
	if _, ok := byAlias[name]; ok {
		return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " twice; it is already an alias")
	}
	if number != -1 {
		if _, ok := byNumber[number]; ok {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + "; bus number " + strconv.Itoa(number) + " is already registered")
		}
	}
	for _, alias := range aliases {
		if _, ok := byName[alias]; ok {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " twice; alias " + strconv.Quote(alias) + " is already a bus")
		}
		if _, ok := byAlias[alias]; ok {
			return errors.New("onewirereg: can't register bus " + strconv.Quote(name) + " twice; alias " + strconv.Quote(alias) + " is already an alias")
		}
	}

	r := &Ref{Name: name, Aliases: make([]string, len(aliases)), Number: number, Open: o}
	copy(r.Aliases, aliases)
	byName[name] = r
	if number != -1 {
		byNumber[number] = r
	}
	for _, alias := range aliases {
		byAlias[alias] = r
	}
	return nil
}

// Unregister removes a previously registered 1-wire bus.
//
// This can happen when an 1-wire bus is exposed via an USB device and the
// device is unplugged.
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()
	r := byName[name]
	if r == nil {
		return errors.New("onewirereg: can't unregister unknown bus name " + strconv.Quote(name))
	}
	delete(byName, name)
	delete(byNumber, r.Number)
	for _, alias := range r.Aliases {
		delete(byAlias, alias)
	}
	return nil
}

//

var (
	mu     sync.Mutex
	byName = map[string]*Ref{}
	// Caches
	byNumber = map[int]*Ref{}
	byAlias  = map[string]*Ref{}
)

// getDefault returns the Ref that should be used as the default bus.
func getDefault() *Ref {
	var o *Ref
	if len(byNumber) == 0 {
		// Fallback to use byName using a lexical sort.
		name := ""
		for n, o2 := range byName {
			if len(name) == 0 || n < name {
				o = o2
				name = n
			}
		}
		return o
	}
	number := int((^uint(0)) >> 1)
	for n, o2 := range byNumber {
		if number > n {
			number = n
			o = o2
		}
	}
	return o
}

func insertRef(l []*Ref, r *Ref) []*Ref {
	n := r.Name
	i := search(len(l), func(i int) bool { return l[i].Name > n })
	l = append(l, nil)
	copy(l[i+1:], l[i:])
	l[i] = r
	return l
}

// search implements the same algorithm as sort.Search().
//
// It was extracted to to not depend on sort, which depends on reflect.
func search(n int, f func(int) bool) int {
	lo := 0
	for hi := n; lo < hi; {
		if i := int(uint(lo+hi) >> 1); !f(i) {
			lo = i + 1
		} else {
			hi = i
		}
	}
	return lo
}
