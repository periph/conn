// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire_test

import (
	"fmt"
	"log"

	"periph.io/x/conn/driver/driverreg"
	"periph.io/x/conn/onewire"
	"periph.io/x/conn/onewire/onewirereg"
)

func Example() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
	if _, err := driverreg.Init(); err != nil {
		log.Fatal(err)
	}

	// Use onewirereg 1-wire bus registry to find the first available 1-wire bus.
	b, err := onewirereg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Dev is a valid conn.Conn.
	d := &onewire.Dev{Addr: 23, Bus: b}

	// Send a command and expect a 5 bytes reply.
	write := []byte{0x10}
	read := make([]byte, 5)
	if err := d.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", read)
}

func ExamplePins() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
	if _, err := driverreg.Init(); err != nil {
		log.Fatal(err)
	}

	// Use onewirereg 1-wire bus registry to find the first available 1-wire bus.
	b, err := onewirereg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Prints out the gpio pin used.
	if p, ok := b.(onewire.Pins); ok {
		fmt.Printf("Q: %s", p.Q())
	}
}
