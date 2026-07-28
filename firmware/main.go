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
//	  |-- SW4 --> Simon
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
//   - Simon needs no exception either: holding all four isn't something
//     normal play ever does, so if it happens the round simply runs out
//     the gesture the same as any other mode would.
//   - Rave is the real conflict, handled by the raveAtFull precondition.
func handleReset(now uint32) bool {
	if current == stateIdle || !allHeld() {
		resetArmed = false
		return false
	}

	if current == stateRave && !ravePeaked() {
		resetArmed = false // still on the first climb; normal play, not a gesture
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
		// Clear first so the sweep plays against a dark board rather than
		// whatever the mode left lit, then confirm, then settle into IDLE.
		allLEDsOff()
		resetSweep()
		enterIdle()
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// LED sweeps
// ---------------------------------------------------------------------------

// sweep lights each LED in turn, one at a time, holding each for stepMS.
//
// Direction is the whole point. Power-on and reset both leave the device in
// IDLE, so the animation direction is the only thing that tells them apart:
// forwards (L1 -> L4) means the device just booted, backwards (L4 -> L1) means
// it just reset.
//
// Blocking, which is fine at both call sites — nothing else needs to run.
// Inputs aren't sampled during the sweep, but the debouncer catches up within
// a few milliseconds once the loop resumes, so a switch released mid-sweep is
// still registered normally.
func sweep(descending bool, stepMS uint32) {
	for n := 0; n < numKeys; n++ {
		i := n
		if descending {
			i = numKeys - 1 - n
		}
		ledPins[i].High()
		time.Sleep(time.Duration(stepMS) * time.Millisecond)
		ledPins[i].Low()
	}
}

// post is the power-on self test: forwards, L1 -> L4. Doubles as a wiring
// check — any LED dark here is a hardware problem, not a code one.
func post() {
	sweep(false, postStepMS)
}

// resetSweep confirms the reset gesture: backwards, L4 -> L1. Plays while the
// switches are still held. It means "done, let go."
func resetSweep() {
	sweep(true, resetSweepStepMS)
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
			case stateSimon:
				updateSimon(now)
			}
		}

		// One software-PWM carrier period. This is also what paces the loop,
		// and therefore the input sample rate.
		renderPWM()
	}
}
