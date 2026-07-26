// Enhanced Keyboard Fidget — firmware
//
// Hardware: Metro (ATmega328P, 16 MHz, 5V), 4 tactile switches, 4 red LEDs.
// Wiring:   SCHEMATIC.md
// Behavior: behavior.md  — that document is the spec; this file implements it.
//
// Build:
//	tinygo flash -target=arduino-uno -port=/dev/cu.usbserial-ADAOKdPsS .
//	tinygo build -o /dev/null -target=arduino-uno -size short .   # check RAM
//
// State machine:
//
//	POST  — 4 x 100ms LED sweep
//	  |
//	IDLE  — dark, waiting
//	  |-- SW1 --> Whack-A-Mole
//	  |-- SW2 --> Rave
//	  |-- SW3 --> Binary Counter
//	         |
//	         +-- all four held 5000ms --> IDLE
//	             (Rave: clock starts only once all four LEDs are at full)
//
// AVR notes — 32KB flash, 2KB RAM. This code allocates nothing on the heap,
// uses no goroutines or channels, no floating point, and no fmt.

package main

import (
	"machine"
	"time"
)

// ---------------------------------------------------------------------------
// Reset gesture
// ---------------------------------------------------------------------------

var (
	resetArmed   bool
	resetArmedAt uint32
)

// handleReset returns true if the gesture fired this pass, in which case the
// caller should skip the mode update — we are already back in IDLE.
//
// Worth noting why this does not collide with the modes:
//
//   - Whack-A-Mole clears on the press edge. Holding all four produces one
//     edge, then nothing; the next mole spawns and simply stays lit.
//   - Binary Counter counts releases. Holding produces none, so the count
//     freezes for the duration of the gesture.
//   - Rave is the real conflict, handled by the raveAtFull precondition.
func handleReset(now uint32) bool {
	if current == stateIdle || !allHeld() {
		resetArmed = false
		return false
	}

	if current == stateRave && !raveAtFull() {
		resetArmed = false // still ramping; this is normal play, not a gesture
		return false
	}

	if !resetArmed {
		resetArmed = true
		resetArmedAt = now
		return false
	}

	if elapsed(now, resetArmedAt, resetHoldMS) {
		if debug {
			println("reset")
		}
		enterIdle()
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Power-on self test
// ---------------------------------------------------------------------------

// post sweeps each LED in turn. Doubles as a wiring check: any LED dark here
// is a hardware problem, not a code one.
func post() {
	for i := 0; i < numKeys; i++ {
		ledPins[i].High()
		time.Sleep(postStepMS * time.Millisecond)
		ledPins[i].Low()
	}
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	bootTime = time.Now()

	for i := 0; i < numKeys; i++ {
		ledPins[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
		ledPins[i].Low()

		keys[i].pin = switchPins[i]
		switchPins[i].Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	// TinyGo's AVR runtime brings the UART up before main() at its default
	// 9600 baud. `tinygo monitor` expects 115200, so without this line every
	// debug print arrives as garbage.
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})

	// Not for the UART's benefit — that's already running. This is slack for
	// a host terminal that's still opening the port after a flash.
	time.Sleep(200 * time.Millisecond)
	if debug {
		println("fidget ready")
	}

	post()
	enterIdle()

	for {
		now := millis()

		sampleInputs(now)
		collectEdges()

		// Deliberately behind its own flag. Every println busy-waits on the
		// UART — roughly 1ms per line at 115200 — and during that stall the
		// PWM loop stops refreshing (LEDs latch at whatever phase they were
		// in) and inputs stop being sampled. Fine for wiring bring-up,
		// not something to leave on while tuning how Rave feels.
		if debugKeys {
			for i := 0; i < numKeys; i++ {
				if edgeDown[i] {
					println("SW", i+1, "down")
				}
				if edgeUp[i] {
					println("SW", i+1, "up")
				}
			}
		}

		// The first press the device ever sees is the only unpredictable
		// event available, so it seeds the RNG.
		for i := 0; i < numKeys; i++ {
			if edgeDown[i] {
				rngSeed(now)
				break
			}
		}

		if !handleReset(now) {
			switch current {
			case stateIdle:
				updateIdle(now)
			case stateMole:
				updateMole(now)
			case stateRave:
				updateRave(now)
			case stateBinary:
				updateBinary(now)
			}
		}

		// One software-PWM carrier period. This is also what paces the loop,
		// and therefore the input sample rate.
		renderPWM()
	}
}
