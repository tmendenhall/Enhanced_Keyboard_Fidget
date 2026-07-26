// config.go — everything you'd want to tune, in one place.
//
// behavior.md says feel matters and that the timing numbers are estimates.
// So every one of them lives here. Change a value, `make flash`, try it.
// Nothing in the other files hard-codes a duration.

package main

import "machine"

// ---------------------------------------------------------------------------
// Pin map — see SCHEMATIC.md
// ---------------------------------------------------------------------------

const numKeys = 4

// Switches: GPIO -> switch -> GND, internal pull-up enabled.
// Active low: the pin reads LOW when the switch is closed.
var switchPins = [numKeys]machine.Pin{
	machine.D2, // SW1
	machine.D3, // SW2
	machine.D4, // SW3
	machine.D7, // SW4
}

// LEDs: GPIO -> 330Ω -> anode, cathode -> GND. Active high.
var ledPins = [numKeys]machine.Pin{
	machine.D5,  // LED1
	machine.D6,  // LED2
	machine.D9,  // LED3
	machine.D11, // LED4
}

// ---------------------------------------------------------------------------
// Timing — the tuning surface. All values in milliseconds.
// ---------------------------------------------------------------------------

const (
	// POST: each LED lights this long, in sequence, at power-on.
	// Four steps, so the whole sweep is 4 x this.
	postStepMS = 100

	// Contact debounce window. Inputs are sampled once per loop pass, and a
	// pass is one full software-PWM cycle plus a 64-bit divide in millis() —
	// call it 1-2.5ms until it's actually measured on hardware. So 5ms is
	// only a handful of samples, not the dozens you'd get on a faster chip.
	// Raise to 10-15 if a single press ever registers twice.
	//
	// To measure: toggle a spare pin (D8 is free) once per pass and scope it.
	debounceMS = 5

	// Whack-A-Mole: pause after a mole is hit, before the next one appears.
	// This is the single biggest lever on how the game feels. Lower is
	// frantic, higher is plodding.
	moleRespawnMS = 50

	// Rave: brightness climbs by 1 every this many ms while a switch is held.
	// 10ms x 255 steps = 2.55s from dark to full.
	raveStepMS = 10

	// Hold all four switches this long to return to IDLE.
	// In Rave the clock does not start until all four LEDs hit full.
	resetHoldMS = 5000
)

// ---------------------------------------------------------------------------
// Display
// ---------------------------------------------------------------------------

// Brightness resolution for software PWM. Must be 1..256.
//
// Higher = finer fades but a slower carrier, and a slow carrier reads as
// flicker. This also sets the loop rate, and therefore the input sample rate.
// 64 is the starting guess at 16MHz; if the LEDs flicker, halve it before
// changing anything else.
const pwmSteps = 64

// ---------------------------------------------------------------------------
// Debug
// ---------------------------------------------------------------------------

// Serial debug at 115200 baud:
//
//	tinygo monitor
//
// behavior.md asks for switch and mode reporting, so both ship on.
//
// These are split because they cost very different amounts. Mode changes are
// rare, so `debug` is nearly free. Key edges fire constantly and each println
// busy-waits the UART for ~1ms, stalling both the PWM refresh and input
// sampling — so turn `debugKeys` off once the wiring is proven, and
// definitely before tuning anything by feel.
//
// Both are compile-time constants, so switching them off removes the call
// sites and their format strings from flash entirely. It does not save much
// RAM: TinyGo brings up the UART and its receive buffer before main() either
// way. If RAM ever gets genuinely tight, the levers are `-gc=leaking` and
// reading `-size full`.
const (
	debug     = true
	debugKeys = true
)
