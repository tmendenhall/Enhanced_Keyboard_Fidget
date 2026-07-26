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
| 5 — Firmware to spec | ⬜ **Written, not yet compiled.** First `make flash` is the real test |
| 6 — Tuning | ⬜ |
| 7 — Stretch (Simon) | ⬜ Out of scope for v1 |

## Files

| File | What it is | State |
|---|---|---|
| `parts.md` | Inventory | Filled in |
| `behavior.md` | Behavior spec — source of truth for firmware | Filled in, synced to 5000 ms |
| `SCHEMATIC.md` | Pin map, resistor math, breadboard layout, switch orientation test | Redrawn for 30 columns |
| `TOOLCHAIN-MACOS.md` | Install + GoLand setup | Done, verified on hardware |
| `fidget_blink/` | Blink sanity check | Working |
| `firmware/config.go` | Pin map + every tunable timing constant | Written |
| `firmware/hardware.go` | Timebase, debounce, software PWM, RNG | Written |
| `firmware/modes.go` | The three modes + state transitions | Written |
| `firmware/main.go` | POST, reset gesture, main loop | Written |

---

## Decisions locked in

**Resistors: 330 Ω, all four channels.** Red 3mm at Vf ≈ 2.0 V on 5 V gives (5 − 2.0) / 330 = **9.1 mA per LED**, 36.4 mA total. Well inside the 20 mA per-pin and 200 mA total limits. Eleven spares in the bin.

**Reset gesture: all four switches held 5000 ms.** Up from 2000 ms. Returns the device to IDLE, where the next single press selects a mode.

**Rave mode exception: the 5-second timer starts only after all four LEDs reach full brightness.** This resolves the collision — holding all four *is* normal Rave play, so the hold clock can't start until the ramp is finished. Practical consequence: the full gesture in Rave takes **2.55 s of ramp + 5 s of hold ≈ 7.6 s**. Long, but unambiguous, and it can't fire by accident. Releasing any switch at any point aborts and resets the ramp.

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

`firmware/main.go` currently holds placeholder modes that predate the spec. The scaffolding survives — debouncing, sticky edge latching, software PWM, gamma correction, the millisecond timebase — but the mode layer is replaced.

### State machine

```
POST — 400ms LED sweep, 100ms each
  ↓
IDLE — all LEDs off, awaiting mode selection
  │
  ├── SW1 → MODE 1  Whack-A-Mole
  ├── SW2 → MODE 2  Rave
  └── SW3 → MODE 3  Binary Counter
        │
        └── all four held 5000ms → IDLE
            (Rave: clock starts only once all four LEDs are at 255)
```

### Why the reset gesture doesn't fight the modes

Worth spelling out, because it's non-obvious and it's the kind of thing that produces a confusing bug later:

- **Whack-A-Mole** clears a mole on the press *edge*. Hold all four down and the first edge clears the current mole; after that there are no further edges, so the next mole spawns and simply stays lit while the 5-second clock runs. No rapid-fire, no interference.
- **Binary Counter** increments on a full click — press *and* release. Holding four switches produces no releases, so the count freezes during the gesture.
- **Rave** is the only real conflict, handled by the ramp-completion precondition above.

### Mode notes

**1 — Whack-A-Mole.** Needs randomness. `math/rand` is far too heavy for AVR; a 16-bit xorshift or LFSR is ~10 lines and a couple bytes of state. Seeding is the interesting part — there's no entropy at power-on, so a fixed seed replays the identical mole sequence every boot. Fix: seed from elapsed milliseconds at the *first* user press. Genuinely unpredictable, costs nothing.

**2 — Rave.** +1 brightness per 10 ms → 2.55 s dark to full. Existing software PWM handles it. Note that gamma correction makes the ramp *look* roughly linear rather than rushing to bright early — which is what you want perceptually, but it's a change if you were picturing raw linear PWM.

**3 — Binary Counter.** Any switch click increments. L1 = LSB, L4 = MSB, wraps 15 → 0.

### Persistence — not implementing, here's what it would cost

You asked for the note. The '328 has 1KB of EEPROM, which is where a saved mode belongs (flash has limited write cycles and TinyGo doesn't expose self-programming). One byte: write on mode change, read during POST, and validate on read — EEPROM reads `0xFF` when never written, so you need a magic byte or sentinel to distinguish "unset" from a real value. Endurance is ~100,000 writes per cell: effectively unlimited for mode changes, but never write it in a loop. TinyGo's AVR EEPROM support is thin, so expect direct register access. Roughly 30 lines. Not hard, just not free.

### All timing in one block

The spec says feel matters and these numbers are estimates. Every one becomes a named constant at the top of the file — tuning is a one-line edit and a reflash.

| Constant | Value | Meaning |
|---|---|---|
| `postStepMS` | 100 | POST sweep, per LED |
| `debounceMS` | 5 | contact debounce window |
| `moleRespawnMS` | 50 | Whack-A-Mole, delay before next mole |
| `raveStepMS` | 10 | Rave, per brightness increment |
| `resetHoldMS` | **5000** | all-four-held reset |

**Serial debug** is on per the spec — switch transitions and current mode. `println` isn't free on a 2KB chip, so it stays behind the `debug` constant. Turn it off once wiring is proven and reclaim the RAM.

---

## Remaining spec gaps

Both minor. Defaults chosen; overrule at any point.

**Wrong switch pressed in Whack-A-Mole.** Unspecified. **Default: ignore it.** No penalty, no flash. Friendlier, and a penalty mechanic implies scoring, which you've explicitly declined.

**SW4 pressed in IDLE.** SW1–SW3 select modes; SW4 has no mode. **Default: silent no-op.** The alternative is a brief flash meaning "not a mode," which is friendlier but adds a state. Easy to add later if the silence feels broken.

---

## Design decisions (carried forward)

**Pins.** The '328 has hardware PWM on exactly six pins — 3, 5, 6, 9, 10, 11 — and all four LEDs sit on that list. SW1/SW2 land on INT0/INT1 if interrupt-driven input is ever wanted. D0/D1 stay clear of the UART so serial debug keeps working; D13 is avoided because the onboard LED hangs off it. Twelve pins spare.

**Software PWM, not hardware.** Timer0 drives TinyGo's millisecond clock on AVR — reconfiguring it for PWM breaks `time.Since()`, which debouncing and animation depend on. One code path for all four LEDs. Timer1 and Timer2 stay free if Rave ever needs CPU back.

**Inputs sampled ~1 ms, animation at 125 Hz.** Deliberately decoupled — sampling at the animation rate would make a 5 ms debounce window meaningless. Press edges latch as sticky flags so nothing is dropped between the fast input loop and the slower render pass.

**Gamma correction.** A linear fade appears to hold near full brightness for most of its travel, then snap off. Squaring approximates the eye's response for one multiply.

**Active-low switches, internal pull-ups.** No external resistors or capacitors. Also exactly how an MX switch wires, so the prototype transfers one-for-one.

**LEDs straight from GPIO.** 9.1 mA each, 36.4 mA total against a 200 mA budget. No transistors, no positive rail.

**No heap, no goroutines, no floats, no `fmt`.** Written for 2KB from the start.

---

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| 2KB RAM ceiling | Three modes + RNG + serial debug exceeds the placeholder's footprint | `make size` after every change; `debug` off when not needed |
| Mixed 4mm/6mm switch feel | Whack-A-Mole timing feels uneven between positions | Place matched pairs symmetrically (4mm at SW1/SW4, 6mm at SW2/SW3), or buy four matching |
| Missing rail-to-rail ground link | Board looks finished, LEDs work, all four switches read permanently pressed | Step 3 of the bring-up checklist tests continuity from a *bottom* rail hole back to Metro GND |
| Rave reset gesture is ~7.6 s end to end | May feel unresponsive or broken mid-hold | Consider a visual cue when the hold clock actually starts — flag for tuning |
| TinyGo AVR backend maturity | Hard-to-attribute misbehavior | Blink already verified, so the toolchain is never the suspect |

---

## Next actions

1. ~~Redraw `SCHEMATIC.md` §4 for the 30-column board.~~ ✅
2. ~~Sync `behavior.md` hold duration to 5000 ms.~~ ✅
3. **Continuity-test all four switches** per `SCHEMATIC.md` §5 — before anything goes in the breadboard.
4. **Build it**, working the bring-up checklist in `SCHEMATIC.md` §7 in order.
5. Replace the placeholder firmware with the real state machine.
