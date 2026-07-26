package main

import (
	"machine"
	"time"
)

func main() {
	led := machine.LED // D13
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		led.High()
		time.Sleep(300 * time.Millisecond)
		led.Low()
		time.Sleep(300 * time.Millisecond)
	}
}
