# Enhanced Keyboard Fidget

A four-switch, four-LED desk fidget running Go on an ATmega328. Press a switch, get a light. Four interaction modes, selected at power-on, plus a deliberate reset gesture to get back between them.

Built on a breadboard with tactile buttons standing in for MX keyboard switches — an MX switch is a plain SPST momentary contact, so the prototype transfers to real switches one-for-one.

**Status:** working. Currently in the tuning phase — the behavior is complete, the timing constants are estimates being dialled in by feel.

---

## What it does

```
POWER ON
  │
  ▼
POST — forward sweep, L1 → L2 → L3 → L4, 100ms each
  │
  ▼
IDLE — all LEDs dark, waiting
  │
  ├── SW1 → Whack-A-Mole
  ├── SW2 → Rave
  ├── SW3 → Binary Counter
  └── SW4 → Simon
        │
        │  all four switches held 5000ms
        ▼
      RESET SWEEP — reverse, L4 → L3 → L2 → L1
        │
        ▼
      IDLE
```

(Simon has a second way back to IDLE — see below.)

Direction carries the meaning. Power-on and reset leave the device in the same state, so the sweep runs **forwards for boot** and **backwards for reset** — that's the only thing distinguishing them, and it needs no explanation to read.

### Modes

**1 — Whack-A-Mole.** One random LED lights. Press its switch and it goes out; 50 ms later another appears somewhere. Wrong-switch presses are ignored. No score, no timer, no end — the reset gesture is the only exit.

**2 — Rave.** Each held switch drives its LED on a continuous triangle: `0 → 255 → 0`, repeating, 5.1 s per cycle. Release turns it off and resets that channel to dark. Four independent channels, each phased by when its switch went down.

**3 — Binary Counter.** Any completed click — press *and* release — increments a 4-bit counter. **L4 is the LSB, L1 the MSB**, so the row reads left to right exactly as you'd write the number down and the rightmost LED toggles fastest. Wraps 15 → 0.

| Value | L1 | L2 | L3 | L4 |
|---|---|---|---|---|
| 1 | · | · | · | ● |
| 2 | · | · | ● | · |
| 3 | · | · | ● | ● |
| 8 | ● | · | · | · |
| 15 | ● | ● | ● | ● |

**4 — Simon.** A random pattern of LEDs plays back one at a time, then you reproduce it on the matching switches, in order. A fully correct repeat pauses briefly and grows the pattern by one, up to 15 steps. Press the wrong switch at any point and all four LEDs blink together, then the game restarts at length 1. Reach and repeat a 15-step pattern and the device plays a win sweep — two round trips, L1→L4→L1→L4→L1 — and returns straight to IDLE, no reset hold required.

State machine and per-transition variable writes for all four modes are diagrammed in [`docs/STATE-DIAGRAMS.md`](docs/STATE-DIAGRAMS.md).

### The reset gesture

Hold all four switches for five seconds, anywhere, and the device returns to IDLE. The reverse sweep confirms it: *done, let go.*

Rave gets an exception. Holding all four switches **is** normal Rave play, so the hold clock doesn't start until every channel has reached full brightness at least once. Note "at least once" rather than "simultaneously" — the channels cycle independently, so all four are essentially never at 255 on the same pass, and testing for that alignment would make the reset impossible to trigger. A latched per-channel flag preserves the intent (don't fire during the first climb) without requiring an alignment that won't happen.

The gesture doesn't collide with the other modes, for reasons worth knowing:

- **Whack-A-Mole** clears on the press *edge*. Holding produces one edge, then nothing — the next mole spawns and simply stays lit.
- **Binary Counter** counts *releases*. Holding produces none, so the count freezes.
- **Simon** needs no exception either — holding all four isn't something normal play ever does, so if it happens the gesture just runs its course like it would in any other mode.

---

## Hardware

| | |
|---|---|
| Board | Metro — ATmega328P, 16 MHz, 5V logic, 20 digital I/O |
| LEDs | 4 × red 3mm, Vf ≈ 2.0 V |
| Resistors | 4 × 330 Ω → 9.1 mA per LED, 36.4 mA total |
| Switches | 4 × tactile momentary (2 × 4mm, 2 × 6mm) |
| Breadboard | 300-point, 30 columns, 4 power rails |
| Wires | 14 |

### Pin map

| Function | Pin | Note |
|---|---|---|
| SW1 – SW4 | `D2` `D3` `D4` `D7` | Active low, internal pull-ups. D2/D3 are INT0/INT1 if interrupts are ever wanted |
| LED1 – LED4 | `D5` `D6` `D9` `D11` | All four are hardware-PWM pins on the '328 |
| Ground | `GND` | → top `−` rail, plus a rail-to-rail link |

`D0`/`D1` stay clear of the UART so serial debugging keeps working, and `D13` is avoided because the onboard LED hangs off it. That leaves `D8`, `D10`, `D12` and all six analog pins spare.

### Circuit

```
   GPIO ──[ 330Ω ]──►|──── GND          5V ──[ internal pull-up ]──┬── GPIO
                    LED                                            │
       long leg (anode) ─┘  └─ short leg (cathode)                ─┴─ SW
                                                                   │
                                                                  GND
```

LEDs are driven straight from the GPIO pins — 9.1 mA each against a 20 mA per-pin limit, so no transistors and no power rail. Switches need no external resistors or capacitors; debouncing is done in firmware.

**One quirk worth knowing.** These particular tactile switches close *along the top row* rather than across the breadboard's center channel — the opposite of a classic 6mm tactile, where the pins on each side are permanently bonded. The body only seats one way, so the layout is built around it: each switch bridges two top-half column nodes, and the two pins landing in row F connect to nothing at all. A pleasant side effect is that the usual short-to-ground failure mode is impossible here.

Full hole-by-hole wiring, resistor math, and the switch continuity test are in [`SCHEMATIC.md`](SCHEMATIC.md).

---

## Build system

Go, compiled with [TinyGo](https://tinygo.org) for the `arduino-uno` target — same MCU, bootloader, and pin mapping as the Metro 328.

### Prerequisites

```bash
brew install go
brew tap tinygo-org/tools
brew trust --formula tinygo-org/tools/tinygo
brew install tinygo
brew install avrdude
```

Verified against Go 1.26.5, TinyGo 0.41.1, avrdude 8.2 on macOS arm64. Full setup including the GoLand plugin is in [`TOOLCHAIN-MACOS.md`](TOOLCHAIN-MACOS.md).

### Commands

All from `firmware/`:

```bash
make flash PORT=/dev/cu.usbserial-XXXX   # build + upload
make monitor PORT=/dev/cu.usbserial-XXXX # serial, 115200
make size                                # flash + RAM usage
make size-full                           # per-package breakdown
make fmt                                 # gofmt
make clean
```

`PORT` is optional — TinyGo auto-detects when only one board is attached, but auto-detection gets unreliable with other USB serial devices present. `TARGET` overrides the same way if you ever move to a different chip.

### Layout

```
├── firmware/            the fidget
│   ├── config.go        pin map + every timing constant — edit here to tune
│   ├── hardware.go      timebase, debounce, software PWM, RNG
│   ├── modes.go         the four modes + state transitions
│   ├── main.go          POST, sweeps, reset gesture, main loop
│   └── Makefile
├── fidget_blink/        minimal blink, for proving the toolchain
├── docs/
│   └── STATE-DIAGRAMS.md  mode state machines + per-loop variable writes
├── PLAN.md              phases, design decisions, risks
├── behavior.md          behavior spec — the source of truth for firmware
├── SCHEMATIC.md         wiring, resistor math, bring-up checklist
├── TOOLCHAIN-MACOS.md   install + GoLand setup
└── parts.md             inventory
```

`firmware/` and `fidget_blink/` are separate Go modules. Open one of them as your project root, not the repository root — the tooling wants `go.mod` at the top.

### Tuning

Every duration lives in `config.go`. Nothing else hard-codes one.

| Constant | Default | Effect |
|---|---|---|
| `postStepMS` | 100 | Boot sweep, per LED |
| `resetSweepStepMS` | `postStepMS` | Reset sweep, per LED |
| `debounceMS` | 5 | Contact debounce window |
| `moleRespawnMS` | 50 | Whack-A-Mole pause between moles — the biggest lever on how the game feels |
| `raveStepMS` | 10 | Rave, per brightness step (10 ms → 2.55 s per leg) |
| `resetHoldMS` | 5000 | Reset gesture hold |
| `pwmSteps` | 32 | Brightness resolution, 1..256 |
| `simonMaxLength` | 15 | Simon, longest pattern before the win sweep (not a duration — a count) |
| `simonPlaybackOnMS` | 400 | Simon, how long each LED stays lit during playback |
| `simonPlaybackGapMS` | 200 | Simon, dark gap between playback steps |
| `simonSuccessGapMS` | 500 | Simon, pause after a correct full repeat, before the pattern grows — raised from the original 100ms spec after play-testing |
| `simonFailBlinkMS` | 50 | Simon, fail-blink on/off duration |
| `simonFailBlinkCount` | 3 | Simon, fail-blink cycle count |
| `simonWinSweepStepMS` | `postStepMS` | Simon, per-LED duration of the win sweep |

Two debug flags, also in `config.go`. `debug` logs mode changes and is nearly free. `debugKeys` logs every switch edge — **turn it off before judging anything by feel**, because each `println` busy-waits the UART for about a millisecond, and during that stall the PWM refresh and input sampling both stop. Tuning with it on means tuning against the instrumentation.

---

## Design notes

**Software PWM, not hardware.** Timer0 drives TinyGo's millisecond clock on AVR, and reconfiguring it for PWM would break `time.Since()` — which debouncing and every mode timer depend on. One code path for all four LEDs. Timer1 and Timer2 stay free if the CPU is ever needed back.

**One loop pass = one PWM carrier cycle = one input sample.** The pass rate sets everything: carrier frequency, sampling rate, and how many samples the debounce window gets. Estimated at 1–2.5 ms but not yet measured; toggle spare pin `D8` once per pass and scope it if debouncing ever feels unreliable.

**Gamma correction.** A linear fade appears to hold near full brightness for most of its travel and then snap off. Squaring approximates the eye's response for one multiply. The correction rounds *up*, because a plain `(v*v)>>8` returns zero for every value ≤ 15 — which would have made the first 160 ms of every Rave ramp completely dark.

**Written for 2KB.** No heap allocation, no goroutines, no channels, no floating point, no `fmt`. Timestamps are `uint32` milliseconds rather than `time.Time` (24 bytes each, and 64-bit arithmetic is slow at 16 MHz), and all elapsed-time comparisons use unsigned subtraction so they stay correct across the ~49-day wrap. Application state is roughly 120 bytes.

---

## Not built

- **Persistence.** Mode is not remembered across power cycles, by choice. The '328 has 1KB of EEPROM if that changes — see `PLAN.md` for what it would cost.
- **USB HID.** Not a keyboard, not planned.
