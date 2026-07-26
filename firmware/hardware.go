// hardware.go — timebase, debounced input, software PWM, and a tiny RNG.
//
// Nothing here knows about modes. This is the layer the mode logic sits on.

package main

import (
	"machine"
	"time"
)

// ---------------------------------------------------------------------------
// Timebase
// ---------------------------------------------------------------------------

var bootTime time.Time

// millis returns milliseconds since boot.
//
// Timestamps are stored as uint32 rather than time.Time deliberately: a
// time.Time is 24 bytes and 64-bit arithmetic is slow at 16MHz. Called once
// per loop pass, never inside the PWM inner loop.
//
// The one expensive part is .Milliseconds(), which is a 64-bit divide by 1e6 —
// software-emulated on AVR, so on the order of tens of microseconds. If the
// loop turns out too slow (visible LED flicker), the cheap swap is
// `uint32(time.Since(bootTime).Nanoseconds() >> 20)`: a shift instead of a
// divide, at the cost of ~4.8% fast timing. Every duration in config.go would
// then read about 5% short.
func millis() uint32 {
	return uint32(time.Since(bootTime).Milliseconds())
}

// elapsed reports whether at least dur ms have passed since the timestamp.
// Unsigned subtraction makes this correct across the ~49-day uint32 wrap;
// comparing timestamps directly would not be.
func elapsed(now, since, dur uint32) bool {
	return now-since >= dur
}

// ---------------------------------------------------------------------------
// Debounced input
// ---------------------------------------------------------------------------

type debouncer struct {
	pin       machine.Pin
	raw       bool   // last raw reading
	state     bool   // debounced state; true = pressed
	changedAt uint32 // when raw last differed from its previous value

	// Set by update(), drained by collectEdges() on the following pass.
	pendingDown bool
	pendingUp   bool
}

func (d *debouncer) update(now uint32) {
	raw := !d.pin.Get() // active low
	if raw != d.raw {
		d.raw = raw
		d.changedAt = now
	}
	if raw != d.state && now-d.changedAt >= debounceMS {
		d.state = raw
		if raw {
			d.pendingDown = true
		} else {
			d.pendingUp = true
		}
	}
}

var (
	keys [numKeys]debouncer

	// Edges for the current pass. collectEdges drains the debouncers every
	// single pass, so a stale edge can never fire later in a different mode.
	//
	// The tradeoff: an edge is only visible for the one pass on which it is
	// collected. Any code path that returns early without reading these
	// discards them. That is intentional in all three places it happens —
	// during a reset gesture, and while IDLE waits for switches to come up —
	// but keep it in mind when adding new early returns.
	edgeDown [numKeys]bool
	edgeUp   [numKeys]bool
)

func sampleInputs(now uint32) {
	for i := 0; i < numKeys; i++ {
		keys[i].update(now)
	}
}

func collectEdges() {
	for i := 0; i < numKeys; i++ {
		edgeDown[i] = keys[i].pendingDown
		edgeUp[i] = keys[i].pendingUp
		keys[i].pendingDown = false
		keys[i].pendingUp = false
	}
}

func anyHeld() bool {
	for i := 0; i < numKeys; i++ {
		if keys[i].state {
			return true
		}
	}
	return false
}

func allHeld() bool {
	for i := 0; i < numKeys; i++ {
		if !keys[i].state {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Software PWM
// ---------------------------------------------------------------------------
//
// The '328 has hardware PWM on all four LED pins, but Timer0 also drives
// TinyGo's AVR millisecond clock — reconfiguring it would break time.Since(),
// which debouncing and every mode timer depend on. Software PWM sidesteps
// that and keeps all four LEDs on one code path.
//
// One complete carrier cycle runs per main-loop pass. Phase sweeps 0..255 in
// steps of 256/pwmSteps; an LED is lit while its display value exceeds the
// phase. No timekeeping happens inside the loop.

var (
	level [numKeys]uint8 // requested brightness, 0..255, linear
	disp  [numKeys]uint8 // gamma-corrected, what the PWM loop compares
)

// gamma corrects for the eye's nonlinear response. Without it a linear ramp
// appears to sit near full brightness for most of its travel and then snap
// off at the end. Squaring is a decent approximation for one multiply.
//
// The +255 rounds up. Plain (v*v)>>8 returns 0 for every v <= 15, which would
// make the first 16 steps of a Rave ramp — 160ms of holding the switch —
// completely dark. Rounding up guarantees any nonzero level emits light.
// gamma(1) = 1, gamma(255) = 255.
func gamma(v uint8) uint8 {
	if v == 0 {
		return 0
	}
	return uint8((uint16(v)*uint16(v) + 255) >> 8)
}

// renderPWM refreshes the gamma table and runs exactly one carrier period.
// Keep the inner loop tight — its length sets the carrier frequency.
//
// Iterating over the step count rather than the phase range keeps this correct
// for any pwmSteps in 1..256. Stepping the phase by 256/pwmSteps would divide
// to zero above 256 (infinite loop, device hangs) and silently round for any
// value that isn't a power of two.
func renderPWM() {
	for i := 0; i < numKeys; i++ {
		disp[i] = gamma(level[i])
	}
	for i := 0; i < pwmSteps; i++ {
		p := uint8(i * 256 / pwmSteps)
		ledPins[0].Set(disp[0] > p)
		ledPins[1].Set(disp[1] > p)
		ledPins[2].Set(disp[2] > p)
		ledPins[3].Set(disp[3] > p)
	}
}

func allLEDsOff() {
	for i := 0; i < numKeys; i++ {
		level[i] = 0
		disp[i] = 0
		ledPins[i].Low()
	}
}

// ---------------------------------------------------------------------------
// Random numbers
// ---------------------------------------------------------------------------
//
// math/rand is far too heavy for AVR. This is a 16-bit xorshift: three shifts
// and three XORs, two bytes of state, period 65535.

var (
	rngState  uint16 = 0xACE1
	rngSeeded bool
)

func rngNext() uint16 {
	rngState ^= rngState << 7
	rngState ^= rngState >> 9
	rngState ^= rngState << 8
	return rngState
}

// rngSeed mixes in a timestamp. Called on the first press the device ever
// sees, because that is the only unpredictable thing available — there is no
// entropy source at power-on, so a fixed seed would replay the identical mole
// sequence on every boot. State must never reach zero or the generator locks.
func rngSeed(now uint32) {
	if rngSeeded {
		return
	}
	rngSeeded = true
	s := uint16(now) ^ uint16(now>>16) ^ 0x5A5A
	if s == 0 {
		s = 0xACE1
	}
	rngState = s
}
