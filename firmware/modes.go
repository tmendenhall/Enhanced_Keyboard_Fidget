// modes.go — the three behaviors from behavior.md §3.
//
// Every update function runs once per main-loop pass and must not block.
// Edges arrive pre-drained in edgeDown / edgeUp; holds are read from
// keys[i].state.

package main

type state uint8

const (
	stateIdle state = iota
	stateMole
	stateRave
	stateBinary
)

var (
	current state

	// After a reset gesture the four switches are still down. behavior.md:
	// "When all 4 switches are off then the next switch press indicates the
	// mode." So IDLE ignores input until everything has come up.
	awaitRelease bool

	// A press that selected a mode will later produce a release edge inside
	// that mode. Binary Counter counts releases, so without this the entry
	// press would immediately increment the counter to 1.
	swallowUp [numKeys]bool
)

// ---------------------------------------------------------------------------
// Transitions
// ---------------------------------------------------------------------------

func enterIdle() {
	current = stateIdle
	allLEDsOff()
	resetArmed = false
	awaitRelease = anyHeld()
	if debug {
		println("mode 0 idle")
	}
}

func enterMode(s state, now uint32) {
	current = s
	allLEDsOff()
	resetArmed = false

	for i := 0; i < numKeys; i++ {
		swallowUp[i] = keys[i].state
	}

	switch s {
	case stateMole:
		moleActive = false
		// Backdate so the first mole appears on the next pass.
		moleClearedAt = now - moleRespawnMS
	case stateRave:
		for i := 0; i < numKeys; i++ {
			raveLastStep[i] = now
			raveFalling[i] = false
			ravePeakedFlag[i] = false
		}
	case stateBinary:
		counter = 0
	}

	if debug {
		println("mode", int(s))
	}
}

// ---------------------------------------------------------------------------
// IDLE — all dark, waiting for a mode selection
// ---------------------------------------------------------------------------

func updateIdle(now uint32) {
	if awaitRelease {
		if !anyHeld() {
			awaitRelease = false
		}
		return
	}

	switch {
	case edgeDown[0]:
		enterMode(stateMole, now)
	case edgeDown[1]:
		enterMode(stateRave, now)
	case edgeDown[2]:
		enterMode(stateBinary, now)
	}
	// SW4 has no mode. Silent no-op — its edge was already drained.
}

// ---------------------------------------------------------------------------
// Mode 1 — Whack-A-Mole
// ---------------------------------------------------------------------------
//
// One random LED lights. Pressing its switch clears it; after moleRespawnMS a
// new random one appears. Wrong-switch presses are ignored (their edges are
// drained by collectEdges and simply not read here). Runs forever — the only
// exit is the reset hold.

var (
	mole          uint8
	moleActive    bool
	moleClearedAt uint32
)

func updateMole(now uint32) {
	if moleActive {
		if edgeDown[mole] {
			level[mole] = 0
			moleActive = false
			moleClearedAt = now
		}
		return
	}

	if elapsed(now, moleClearedAt, moleRespawnMS) {
		// Take the top two bits, not the bottom two. The final xorshift stage
		// (<<8) leaves the low byte untouched, so the low bits have had one
		// less round of mixing than the high bits.
		//
		// Repeats are allowed — a mole can appear twice in a row. If that
		// feels wrong, reroll while the new value equals the old one.
		mole = uint8(rngNext() >> 14)
		level[mole] = 255
		moleActive = true
	}
}

// ---------------------------------------------------------------------------
// Mode 2 — Rave
// ---------------------------------------------------------------------------
//
// While a switch is held its LED cycles continuously: 0 -> 255 -> 0, repeat,
// one step every raveStepMS in whichever direction it is currently travelling.
// 2.55s per leg, 5.1s per full cycle. Release drops the LED to dark and resets
// the channel, so every press starts from 0 travelling up.
//
// Four independent channels — each with its own level, direction, and phase.

var (
	raveLastStep [numKeys]uint32
	raveFalling  [numKeys]bool // false = climbing toward 255

	// Latched the first time a channel reaches 255 since its switch went down,
	// cleared on release. The reset gesture gates on this — see ravePeaked().
	ravePeakedFlag [numKeys]bool
)

func updateRave(now uint32) {
	for i := 0; i < numKeys; i++ {
		if !keys[i].state {
			level[i] = 0
			raveFalling[i] = false
			ravePeakedFlag[i] = false
			raveLastStep[i] = now
			continue
		}

		if !elapsed(now, raveLastStep[i], raveStepMS) {
			continue
		}
		// Advance by the period rather than snapping to now. Assigning now
		// would round every step up to the next loop pass, stretching a
		// nominal 2.55s leg by however long a pass takes.
		raveLastStep[i] += raveStepMS

		if raveFalling[i] {
			if level[i] == 0 {
				raveFalling[i] = false
				level[i]++
			} else {
				level[i]--
			}
			continue
		}

		if level[i] == 255 {
			ravePeakedFlag[i] = true
			raveFalling[i] = true
			level[i]--
		} else {
			level[i]++
			if level[i] == 255 {
				ravePeakedFlag[i] = true
			}
		}
	}
}

// ravePeaked reports whether every channel has hit full brightness at least
// once since its switch went down. The reset gesture waits on this, because
// holding all four switches IS normal Rave play and the hold clock must not
// start during the first climb.
//
// Deliberately a latch rather than a test for simultaneous full brightness.
// The channels cycle independently, phased by when each switch was pressed, so
// all four are essentially never at 255 on the same pass — testing for that
// would make the reset impossible to trigger in this mode.
func ravePeaked() bool {
	for i := 0; i < numKeys; i++ {
		if !ravePeakedFlag[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Mode 3 — Binary Counter
// ---------------------------------------------------------------------------
//
// Any completed click — press then release — increments a 4-bit counter.
// Wraps 15 -> 0.
//
// Bit order: L4 is the LSB, L1 the MSB, so the row reads left to right exactly
// as you'd write the number down (L1 L2 L3 L4 = bit3 bit2 bit1 bit0) and the
// rightmost LED toggles fastest.
//
//	value 1 -> · · · ●        value 4 -> · ● · ·
//	value 2 -> · · ● ·        value 8 -> ● · · ·
//	value 3 -> · · ● ●        value 15 -> ● ● ● ●
//
// Note this is the opposite of the convenient indexing. `1 << i` would map
// array slot 0 to bit 0 and produce a display that reads backwards; the index
// is flipped at the point of display only, so the counter itself is untouched.

var counter uint8

func updateBinary(now uint32) {
	for i := 0; i < numKeys; i++ {
		if !edgeUp[i] {
			continue
		}
		if swallowUp[i] {
			swallowUp[i] = false // this was the release of the entry press
			continue
		}
		counter = (counter + 1) & 0x0F
	}

	for i := 0; i < numKeys; i++ {
		bit := uint(numKeys - 1 - i) // L1 -> bit3 ... L4 -> bit0
		if counter&(1<<bit) != 0 {
			level[i] = 255
		} else {
			level[i] = 0
		}
	}
}
