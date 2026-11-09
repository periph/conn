// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Example implementing and using a connection interface.

package conn_test

import (
	"fmt"
	"log"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

func Example() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
	if _, err := driverreg.Init(); err != nil {
		log.Print(err)
		return
	}

	// Using SPI as an example. See package ./spi/spireg for more details.
	p, err := spireg.Open("")
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if closeErr := p.Close(); closeErr != nil {
			log.Print(closeErr)
		}
	}()
	c, err := p.Connect(physic.MegaHertz, spi.Mode3, 8)
	if err != nil {
		log.Print(err)
		return
	}

	// Write 0x10 to the device, and read a byte right after.
	write := []byte{0x10, 0x00}
	read := make([]byte, len(write))
	if err := c.Tx(write, read); err != nil {
		log.Print(err)
		return
	}
	// Use read.
	fmt.Printf("%v\n", read[1:])
}
