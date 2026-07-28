# Enhanced Keyboard Fidget — Prototype Plan

Four switches, four LEDs, an ATmega328 Metro, and Go.

| | |
|---|---|
| **Board** | Metro, ATmega328P, 16 MHz, 5V logic, 20 digital I/O (×2 on hand) |
| **TinyGo target** | `arduino-uno` |
| **Serial port** | `/dev/cu.usbserial-ADAOKdPsS` |
| **Toolchain** | Go 1.26.5 · TinyGo 0.41.1 · avrdude 8.2 · GoLand + TinyGo plugin |
| **LEDs** | 4 × red 3mm, Vf ≈ 2.0 V |
| **Resistors** | **330 Ω** (orange-orange-brown), 15 on hand |
| **Breadboard** | 300-point, 30 columns, 4 power rails |

## Status

| Phase | State |
|---|---|
| 1 — Toolchain | ✅ **Complete** — blink verified on hardware |
| 2 — Inventory | ✅ **Complete** |
| 3 — Behavior spec | ✅ **Complete** — two minor gaps remain, defaults chosen |
| 4 — Breadboard | ✅ **Complete** — built, all four switches continuity-checked |
| 5 — Firmware to spec | ✅ **Complete** — three modes, POST, reset gesture, flashed |
| 5a — Reset confirmation sweep | ✅ **Built** — reverse L4→L1 sweep on reset |
| 5b — Binary Counter bit order | ✅ **Built** — L4 = LSB, L1 = MSB |
| 5c — Rave continuous cycle | ✅ **Built** — 0→255→0 repeating; reset gate now a latched flag |
| 6 — Tuning | 🔵 **Current phase.** Dial in the timing constants by feel |
| 7 — Simon | ✅ **Built** — 2026-07-27, fourth mode, triggered by SW4 from IDLE — see behavior.md §3 Mode 4 |

**v1 is built and running.** Everything from here is refinement.

## Files

| File | What it is | State |
|---|---|---|
| `parts.md` | Inventory | Filled in |
| `behavior.md` | Behavior spec — source of truth for firmware | Filled in, synced to 5000 ms |
| `SCHEMATIC.md` | Pin map, resistor math, breadboard layout, switch orientation test | Redrawn for 30 columns |
| `TOOLCHAIN-MACOS.md` | Install + GoLand setup | Done, verified on hardware |
| `fidget_blink/` | Blink sanity check | Working |
| `firmware/config.go` | Pin map + every tunable timing constant | Flashed 2026-07-27 (Simon + `simonSuccessGapMS` raised to 500ms) |
| `firmware/hardware.go` | Timebase, debounce, software PWM, RNG | Flashed, unchanged by Simon |
| `firmware/modes.go` | The four modes + state transitions | Flashed 2026-07-27 (Simon) |
| `firmware/main.go` | POST, reset gesture, main loop | Flashed 2026-07-27 (Simon) |

---

## Decisions locked in

**Resistors: 330 Ω, all four channels.** Red 3mm at Vf ≈ 2.0 V on 5 V gives (5 − 2.0) / 330 = **9.1 mA per LED**, 36.4 mA total. Well inside the 20 mA per-pin and 200 mA total limits. Eleven spares in the bin.

**Reset gesture: all four switches held 5000 ms.** Up from 2000 ms. Returns the device to IDLE, where the next single press selects a mode.

**The reset is confirmed by a reverse LED sweep — L4, L3, L2, L1.** The moment the hold duration is satisfied, the device plays the power-on sweep backwards, then lands in IDLE. Same shape as POST, opposite direction, so the two are never confused: forwards means "I just powered up," backwards means "I just reset."

This closes the one real usability hole in the gesture. Holding four switches for five seconds with nothing happening gives you no way to tell whether the device registered the gesture, whether you started counting at the right moment, or whether it's simply broken — and in Rave the wait is closer to 7.6 s. The sweep answers all three: it means *done, let go*.

Timing is `resetSweepStepMS`, defaulting to `postStepMS` (100 ms per LED, 400 ms total). Separate constant so the two sweeps can be tuned apart if the reset wants to feel snappier than the boot animation.

**Rave mode exception: the 5-second timer starts only after every LED has reached full brightness at least once.** Holding all four *is* normal Rave play, so the hold clock can't start while the lights are still on their first climb. Practical consequence: the full gesture in Rave takes **2.55 s to the first peak + 5 s of hold, then a 0.4 s sweep ≈ 8 s**. Long, but unambiguous, it can't fire by accident, and it now ends with an unmistakable signal. Releasing any switch at any point aborts and clears the peaked flags.

**"At least once" is load-bearing.** Now that Rave cycles continuously rather than ramping once and holding, the four channels are almost never at 255 at the same instant — each one's phase depends on exactly when its switch went down, and nobody presses four switches on the same millisecond. Testing for simultaneous full brightness would make the reset effectively impossible to trigger in Rave. A latched per-channel flag preserves the original intent — don't fire during the first climb — without requiring an alignment that won't occur.

**Whack-A-Mole runs indefinitely.** No score, no timer, no win condition. The 5-second reset hold is the only exit. Confirmed as intentional.

**Breadboard rails confirmed.** Reading top to bottom:

```
  +  ← top positive rail    (unused)
  −  ← top negative rail    (GND, adjacent to row A)
 A–E   top component half
 ═══   center channel
 F–J   bottom component half
  +  ← bottom positive rail (unused, adjacent to row J)
  −  ← bottom negative rail (GND)
```

Columns run **30 on the left down to 1 on the right**.

Both negative rails are usable. Plan: Metro `GND` → top − rail, plus one jumper linking top − to bottom − so ground is reachable from either half. Neither positive rail gets connected to anything — the LEDs source their current from the GPIO pins, so there's no reason to energize them, and leaving them dead removes a whole category of short.

---

## Phase 4 — Breadboard

### Layout — see `SCHEMATIC.md` §4 for the hole-by-hole diagram

Left to right: **SW1, LED1, SW2, LED2, SW3, LED3, SW4, LED4** — each LED immediately right of the switch it belongs to. Whack-A-Mole is unplayable if you can't tell at a glance which switch owns the lit LED.

| Element | Columns | Placement |
|---|---|---|
| SW1 · SW2 · SW3 · SW4 | 30/28 · 24/22 · 18/16 · 12/10 | body straddles the channel; **only the row E pins are wired** |
| LED1 · LED2 · LED3 · LED4 | 26 · 20 · 14 · 8 | resistor across the channel `E`↔`F`, anode at `J`, cathode into the rail |

**14 wires**, down from 21 on the old 63-column layout. Components occupy columns 30–8, leaving 7 spare. Both negative rails are used and linked with a single jumper; both positive rails stay dead.

Two space-savers worth noting:

**The resistor stands across the center channel** — one leg in `E{n}`, one in `F{n}`. That 0.3" span is a natural bend for a ¼W resistor, and it bridges the column's two nodes using a *single* column instead of two.

**The LED cathode goes straight into a rail hole** — no jumper. 3mm legs reach easily from row J to the rail below.

Pin assignment unchanged: switches on **D2, D3, D4, D7**; LEDs on **D5, D6, D9, D11**.

### The switches close along the row, not across the channel

Measured on the actual parts, and it inverts the usual assumption. A classic 6mm tactile has its two pins on each side permanently bonded, with the press bridging *across* the 0.3" dimension. Yours bond nothing across the channel — the press closes the 0.2" dimension instead:

| Probe | Released | Pressed |
|---|---|---|
| `A30` ↔ `A28` (both row E) | silent | **beeps** |
| `A30` ↔ `J30` (across the channel) | silent | silent |

Both the 4mm and 6mm switches behave this way, and the body only seats one way, so the layout is built around it rather than fighting it.

Two consequences, both good:

**The short-to-ground failure mode is gone.** A classic tactile rotated 90° hard-wires the signal pin to ground and reads LOW forever — a hardware fault that looks exactly like a firmware bug. Yours can't do that; nothing is permanently connected between any two pins.

**The bottom half is inert.** Only the row E pins carry signal. `F30`/`F28` connect to nothing, so the circuit behaves identically regardless of what happens down there.

Switch grounds therefore go to the **top** `−` rail (adjacent to row A), which is why the rail-to-rail link exists. Forgetting it is the likeliest build error: the board looks finished, the LEDs work, and every switch reads as permanently pressed.

---

## Phase 5 — Firmware to spec

Four files. `config.go` is the tuning surface; nothing else hard-codes a duration.

### State machine

```
POWER ON
  │
  ▼
POST — forward sweep, L1 → L2 → L3 → L4, 100ms each
  │
  ▼
IDLE — all LEDs off, awaiting mode selection
  │
  ├── SW1 → MODE 1  Whack-A-Mole
  ├── SW2 → MODE 2  Rave
  └── SW3 → MODE 3  Binary Counter
        │
        │   all four switches held 5000ms
        │   (Rave: clock starts only once all four LEDs reach 255)
        ▼
      RESET SWEEP — reverse, L4 → L3 → L2 → L1, 100ms each
        │
        ▼
      IDLE   ← next single press selects a mode,
               but only after all four switches come up
```

Direction is the whole point of the reverse sweep: **forwards means booted, backwards means reset.** Two events that leave the device in the same state, told apart by animation direction alone.

### Reset confirmation sweep

Fires the instant the hold duration is satisfied — while the switches are still down — then the device settles into IDLE.

| Property | Value |
|---|---|
| Order | L4, L3, L2, L1 |
| Per-LED duration | `resetSweepStepMS`, default `postStepMS` = 100 ms |
| Total | 400 ms |
| One LED at a time? | Yes — on, hold, off, next. Identical shape to POST. |
| Blocking? | Yes, same as POST |

Blocking is fine here and keeps it simple. Nothing else needs to run during those 400 ms: the modes are already torn down, and IDLE is about to ignore input anyway until all four switches come up. Input sampling pauses for the duration, but the debouncer catches up within a few milliseconds once the loop resumes, so a release during the sweep is registered normally.

Implementation is one shared helper rather than two near-identical loops — POST calls it ascending, reset calls it descending.

### Why the reset gesture doesn't fight the modes

Worth spelling out, because it's non-obvious and it's the kind of thing that produces a confusing bug later:

- **Whack-A-Mole** clears a mole on the press *edge*. Hold all four down and the first edge clears the current mole; after that there are no further edges, so the next mole spawns and simply stays lit while the 5-second clock runs. No rapid-fire, no interference.
- **Binary Counter** increments on a full click — press *and* release. Holding four switches produces no releases, so the count freezes during the gesture.
- **Rave** is the only real conflict, handled by the ramp-completion precondition above.

### Mode notes

**1 — Whack-A-Mole.** Needs randomness. `math/rand` is far too heavy for AVR; a 16-bit xorshift or LFSR is ~10 lines and a couple bytes of state. Seeding is the interesting part — there's no entropy at power-on, so a fixed seed replays the identical mole sequence every boot. Fix: seed from elapsed milliseconds at the *first* user press. Genuinely unpredictable, costs nothing.

**2 — Rave.** Each held switch drives its LED on a continuous triangle: 0 → 255 → 0, repeat, ±1 per 10 ms. **5.1 s per full cycle**, 2.55 s each leg. Release turns the LED off and resets that channel to 0, so every press starts from dark. Four independent channels, each with its own level *and* direction.

Gamma correction makes the sweep *look* roughly linear rather than rushing to bright early — what you want perceptually, but a change if you were picturing raw linear PWM.

Per-channel state grows by two bytes: a direction flag and a "has peaked" flag (see below). Still trivial against the RAM budget.

**3 — Binary Counter.** Any switch click increments; wraps 15 → 0.

**Bit order: L4 is the LSB, L1 the MSB.** The row reads left to right exactly as you'd write the number down — `L1 L2 L3 L4` = `bit3 bit2 bit1 bit0`. Counting up, the rightmost LED toggles fastest, which is what anyone who has read a binary number expects. The reverse (L1 = LSB) is the more natural thing to *write* in code, since array index 0 maps to bit 0, and that's exactly why it's worth being deliberate here: the convenient indexing produces a display that reads backwards.

| Value | L1 | L2 | L3 | L4 |
|---|---|---|---|---|
| 0 | · | · | · | · |
| 1 | · | · | · | ● |
| 2 | · | · | ● | · |
| 3 | · | · | ● | ● |
| 4 | · | ● | · | · |
| 8 | ● | · | · | · |
| 15 | ● | ● | ● | ● |

In code this is one index flip — `counter & (1 << (numKeys-1-i))` rather than `1 << i`. No other mode is affected; nothing else in the firmware assigns meaning to LED order.

### Persistence — not implementing, here's what it would cost

You asked for the note. The '328 has 1KB of EEPROM, which is where a saved mode belongs (flash has limited write cycles and TinyGo doesn't expose self-programming). One byte: write on mode change, read during POST, and validate on read — EEPROM reads `0xFF` when never written, so you need a magic byte or sentinel to distinguish "unset" from a real value. Endurance is ~100,000 writes per cell: effectively unlimited for mode changes, but never write it in a loop. TinyGo's AVR EEPROM support is thin, so expect direct register access. Roughly 30 lines. Not hard, just not free.

### All timing in one block

The spec says feel matters and these numbers are estimates. Every one becomes a named constant at the top of the file — tuning is a one-line edit and a reflash.

| Constant | Value | Meaning |
|---|---|---|
| `postStepMS` | 100 | POST sweep, per LED — forwards, L1→L4 |
| `resetSweepStepMS` | `postStepMS` | Reset sweep, per LED — backwards, L4→L1 |
| `debounceMS` | 5 | contact debounce window |
| `moleRespawnMS` | 50 | Whack-A-Mole, delay before next mole |
| `raveStepMS` | 10 | Rave, per brightness increment |
| `resetHoldMS` | **5000** | all-four-held reset |

**Serial debug** is on per the spec — switch transitions and current mode, at 115200 baud. Split across two constants: `debug` for mode changes (rare, nearly free) and `debugKeys` for switch edges. Each `println` busy-waits the UART for about a millisecond, which stalls both the PWM refresh and input sampling — so turn `debugKeys` off before judging anything by feel, or you'll be tuning against the instrumentation.

---

## Remaining spec gaps

Both minor. Defaults chosen; overrule at any point.

**Wrong switch pressed in Whack-A-Mole.** Unspecified. **Default: ignore it.** No penalty, no flash. Friendlier, and a penalty mechanic implies scoring, which you've explicitly declined. ✅ **Confirmed 2026-07-27** by play-testing — feels right, keeping as final.

**SW4 pressed in IDLE.** SW1–SW3 select modes; SW4 has no mode. **Default: silent no-op.** The alternative is a brief flash meaning "not a mode," which is friendlier but adds a state. Easy to add later if the silence feels broken. 🔵 **Still open — 2026-07-27:** intended to be resolved soon, not finalized. Likely candidate: the brief-flash alternative above.

---

## Design decisions (carried forward)

**Pins.** The '328 has hardware PWM on exactly six pins — 3, 5, 6, 9, 10, 11 — and all four LEDs sit on that list. SW1/SW2 land on INT0/INT1 if interrupt-driven input is ever wanted. D0/D1 stay clear of the UART so serial debug keeps working; D13 is avoided because the onboard LED hangs off it. Twelve pins spare.

**Software PWM, not hardware.** Timer0 drives TinyGo's millisecond clock on AVR — reconfiguring it for PWM breaks `time.Since()`, which debouncing and animation depend on. One code path for all four LEDs. Timer1 and Timer2 stay free if Rave ever needs CPU back.

**One loop pass = one PWM carrier cycle = one input sample.** The pass rate sets everything: brightness carrier frequency, input sampling, and how many samples the debounce window gets. Estimated at 1–2.5 ms but **not yet measured** — toggle spare pin `D8` once per pass and scope it if the debounce ever feels unreliable.

**Gamma correction.** A linear fade appears to hold near full brightness for most of its travel, then snap off. Squaring approximates the eye's response for one multiply.

**Active-low switches, internal pull-ups.** No external resistors or capacitors. Also exactly how an MX switch wires, so the prototype transfers one-for-one.

**LEDs straight from GPIO.** 9.1 mA each, 36.4 mA total against a 200 mA budget. No transistors, no positive rail.

**No heap, no goroutines, no floats, no `fmt`.** Written for 2KB from the start.

---

## Risks

Build risks are retired. These are what's left, and they're all tuning-phase concerns.

| Risk | Impact | Mitigation |
|---|---|---|
| Loop rate never measured | Debounce may get only 2–3 samples per window, not the dozens the 5 ms implies | Scope `D8` toggled once per pass; raise `debounceMS` or lower `pwmSteps` from what you see |
| ~~Reset gesture gives no feedback~~ | ~~Feels unresponsive or broken mid-hold~~ | **Resolved** by the reverse sweep — see Phase 5a |
| Rave reset is still ~8 s end to end | Even with the end-of-gesture sweep, there's no signal that the *clock has started* | Optional: a brief dip or pulse the instant all four hit full. Judge by feel first — the completion sweep may be enough |
| Mixed 4mm/6mm switch feel | Whack-A-Mole timing feels uneven between positions | Matched pairs at symmetric positions (4mm at SW1/SW4, 6mm at SW2/SW3), or buy four matching |
| `debugKeys` left on while tuning | Every edge stalls the loop ~1 ms — you'd be tuning against the instrumentation | Turn it off in `config.go` before judging feel — **2026-07-27: tested with debug on, tuning felt fine, risk didn't materialize in practice** |
| 2KB RAM ceiling | Bites when features get added, not now | `make size` after every change; `make size-full` when something needs trimming |

---

## Completed

1. ~~Install the toolchain and verify with a blink.~~ ✅
2. ~~Fill in `parts.md`.~~ ✅
3. ~~Fill in `behavior.md`.~~ ✅
4. ~~Redraw `SCHEMATIC.md` §4 for the 30-column board.~~ ✅
5. ~~Rebuild §5 around the switches' fixed orientation.~~ ✅
6. ~~Sync `behavior.md` hold duration to 5000 ms.~~ ✅
7. ~~Continuity-test all four switches.~~ ✅
8. ~~Build the breadboard.~~ ✅
9. ~~Replace the placeholder firmware with the real state machine.~~ ✅

---

## Next actions — Phase 6, tuning

No deadline on any of these. The device works; this is making it feel right.

1. ~~**Turn off `debugKeys`** in `config.go` before judging anything by feel.~~ ✅ **2026-07-27:** tested with debug on — tuning felt fine either way, not a real issue in practice.
2. **Play each mode and note what's wrong.** The likely suspects, in order: `moleRespawnMS` (50 ms — the biggest lever on how Whack-A-Mole feels), `raveStepMS` (10 ms → 2.55 s to full), then `debounceMS` if any press ever registers twice. Simon is new and unflashed — flash it and feel out `simonPlaybackOnMS`/`simonPlaybackGapMS` (both estimates, no play-testing yet) before trusting them.
3. **Judge whether the reset needs a *start*-of-hold cue too.** The reverse sweep confirms completion; it doesn't tell you the clock has begun. In Rave especially, that's several seconds of holding on faith. Try it before adding anything.
4. **Confirm the two spec-gap defaults** now that you can feel them: wrong-switch presses ignored in Whack-A-Mole ✅ confirmed 2026-07-27, SW4 silent in IDLE 🔵 still open, to be resolved soon.
5. ~~**Record `make size`** so there's a baseline to compare against when features get added.~~ ✅ **2026-07-27:** flash 12,687 / 32,768 (38.7%), ram 766 / 2,048 (37.4%).

Optional, whenever: measure the loop rate on `D8`; persist the last mode to EEPROM; swap in four matching switches; build the Simon mode from `behavior.md` §6.
