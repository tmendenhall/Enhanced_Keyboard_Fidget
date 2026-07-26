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
// Each LED ramps 0 -> 255 while its own switch is held, one step every
// raveStepMS. Release drops it straight back to dark. Four independent
// channels, no interaction between them.

var raveLastStep [numKeys]uint32

func updateRave(now uint32) {
	for i := 0; i < numKeys; i++ {
		if !keys[i].state {
			level[i] = 0
			raveLastStep[i] = now
			continue
		}
		if level[i] < 255 && elapsed(now, raveLastStep[i], raveStepMS) {
			level[i]++
			// Advance by the period rather than snapping to now. Assigning
			// now would round every step up to the next loop pass, stretching
			// a nominal 2.55s ramp by however long a pass takes.
			raveLastStep[i] += raveStepMS
		}
	}
}

// raveAtFull reports whether all four channels have finished ramping. The
// reset gesture waits on this — holding all four switches IS normal Rave play,
// so the hold clock cannot be allowed to start during the ramp.
func raveAtFull() bool {
	for i := 0; i < numKeys; i++ {
		if level[i] != 255 {
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
// L1 is the LSB, L4 the MSB. Wraps 15 -> 0.

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
		if counter&(1<<uint(i)) != 0 {
			level[i] = 255
		} else {
			level[i] = 0
		}
	}
}
