// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c_test

import (
	"flag"
	"fmt"
	"log"

	"periph.io/x/conn/driver/driverreg"
	"periph.io/x/conn/i2c"
	"periph.io/x/conn/i2c/i2creg"
)

func Example() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
	if _, err := driverreg.Init(); err != nil {
		log.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Dev is a valid conn.Conn.
	d := &i2c.Dev{Addr: 23, Bus: b}

	// Send a command 0x10 and expect a 5 bytes reply.
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

	// Use i2creg I²C port registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Prints out the gpio pin used.
	if p, ok := b.(i2c.Pins); ok {
		fmt.Printf("SDA: %s", p.SDA())
		fmt.Printf("SCL: %s", p.SCL())
	}
}

func ExampleAddr_flag() {
	var addr i2c.Addr
	flag.Var(&addr, "addr", "i2c device address")
	flag.Parse()
}
