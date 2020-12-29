// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package uart_test

import (
	"fmt"
	"log"

	"periph.io/x/conn/driver/driverreg"
	"periph.io/x/conn/uart"
	"periph.io/x/conn/uart/uartreg"
)

func ExamplePins() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
	if _, err := driverreg.Init(); err != nil {
		log.Fatal(err)
	}

	// Use uartreg UART port registry to find the first available UART port.
	p, err := uartreg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Prints out the gpio pin used.
	if p, ok := p.(uart.Pins); ok {
		fmt.Printf("  RX : %s", p.RX())
		fmt.Printf("  TX : %s", p.TX())
		fmt.Printf("  RTS: %s", p.RTS())
		fmt.Printf("  CTS: %s", p.CTS())
	}
}
