# Breadboard Schematic — 4 Switches + 4 LEDs

**Board: Metro** — ATmega328P, 16 MHz, 5V logic, 20 digital I/O (D0–D13 plus A0–A5), USB programming.
**Breadboard: 300-point**, 30 columns, 4 power rails.
**LEDs: 4 × red 3mm** behind **330 Ω** (orange-orange-brown).
**Switches: 4-pin tactile**, contact closes **along the top row**, not across the channel — see §5.

---

## 1. Pin assignment

| Function | Pin | Why this pin |
|---|---|---|
| SW1 | `D2` | INT0 — hardware interrupt available if we ever want it |
| SW2 | `D3` | INT1 — same |
| SW3 | `D4` | Plain digital in |
| SW4 | `D7` | Plain digital in |
| LED1 | `D5` | OC0B — hardware PWM |
| LED2 | `D6` | OC0A — hardware PWM |
| LED3 | `D9` | OC1A — hardware PWM |
| LED4 | `D11` | OC2A — hardware PWM |
| Ground | `GND` | Any GND pin; one is enough |

The '328 has hardware PWM on exactly six pins — **3, 5, 6, 9, 10, 11**. All four LEDs are on that list, so brightness control stays available whichever way we implement it.

**Deliberately avoided**

- `D0` / `D1` — hardware UART, and the USB-serial bridge is wired to them. Using these breaks both programming and `tinygo monitor`.
- `D13` — tied to the onboard LED through a resistor. Unreliable as an input.

**Left free:** `D8`, `D10`, `D12`, and all six analog pins (`A0`–`A5`, usable as plain digital I/O). Twelve spare pins.

Note on hardware PWM: **Timer0 also drives TinyGo's millisecond clock** on AVR. Reconfiguring it for PWM would break `time.Since()`, which the firmware depends on for debouncing and animation. That's why the firmware uses software PWM. Timer1 (`D9`/`D10`) and Timer2 (`D11`/`D3`) are free if you later want hardware PWM on a subset.

---

## 2. Logical schematic

### Switch (×4) — active low, internal pull-up

```
             ┌─────────── MCU internal pull-up (~20-50kΩ, enabled in firmware)
             │
      5V ────┴──[ Rpu ]──┬─────► GPIO Dn   HIGH when open, LOW when pressed
                         │
                        ─┴─
                         ○  SW  (SPST momentary)
                        ─┬─
                         │
                        GND
```

No external resistor. No external capacitor. Debouncing is done in firmware at 5 ms.

This is exactly how an MX-style keyboard switch wires up — an MX switch is a plain SPST momentary contact. Whatever you prototype with transfers over one-for-one.

### LED (×4) — current-sourced from the GPIO

```
   GPIO Dn ──[ 330Ω ]──►|──── GND
                       LED
              long leg (anode) ─┘  └─ short leg (cathode), flat spot on rim
```

LED lights when the pin is driven HIGH.

---

## 3. Resistor value

Your parts: red 3mm LEDs, Vf ≈ 2.0 V, on a 5 V rail.

```
R = (5.0 − 2.0) / 0.009 = 333 Ω   →   use 330 Ω
Actual current: (5.0 − 2.0) / 330 = 9.1 mA per LED
```

**Use the orange-orange-brown resistors — 330 Ω.** Four needed, eleven spare.

If 9.1 mA turns out brighter than you want, the brown-black-red parts are 1 kΩ and give 3.0 mA — noticeably dimmer but perfectly readable. Anything from 220 Ω to 1 kΩ is electrically safe here; you're only trading brightness.

### Current limits — ATmega328P

| Limit | Value | Our draw |
|---|---|---|
| Per pin, recommended | 20 mA | 9.1 mA |
| Per pin, absolute max | 40 mA | 9.1 mA |
| Total through VCC or GND | 200 mA | 36.4 mA |

Comfortable margin everywhere.

---

## 4. Breadboard layout

### The board

Rails, reading from the top edge downward:

```
  +   top positive rail      ← UNUSED
  −   top negative rail      ← GND. Switches ground here.
 A B C D E                     top component half
 ═══════════ center channel ═══════════
 F G H I J                     bottom component half
  +   bottom positive rail   ← UNUSED
  −   bottom negative rail   ← GND. LED cathodes ground here.
```

Columns run **30 at the left down to 1 at the right**.

**The one rule that matters:** in any column, holes **A–E form one connected node**, and holes **F–J form a second, separate node**. The center channel separates them. A component with both legs in the same node is shorted out and does nothing.

Both negative rails are used and get linked with a single jumper. Both positive rails stay dead — the LEDs source their current from the GPIO pins, so there's nothing to power.

### How the switch is wired, given its fixed orientation

Your switches close **between two columns within the top half** (§5). The body still straddles the channel — that's the only way it physically seats — but the two pins that land in row F are **not connected to anything**. Nothing is wired to them, so whatever they do internally is irrelevant to this circuit.

That means each switch is simply a contact between **column *n*'s top node** and **column *n+2*'s top node**: signal in on one, ground out on the other.

### Layout

Left to right: SW1, LED1, SW2, LED2, SW3, LED3, SW4, LED4 — each LED immediately to the right of the switch it belongs to.

```
 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10  9  8  7  6  5  4  3  2  1
┌───────────────────────────────────────────────────────────────────────────────────────────┐
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │  + rail  UNUSED
│ ·  ·  g  ·  ·  ·  ·  ·  g  ·  ·  ·  ·  ·  g  ·  ·  ·  ·  ·  g  ·  ·  ·  ·  ·  ·  ·  L  G  │  − rail  GND
├───────────────────────────────────────────────────────────────────────────────────────────┤
│ •  ·  g  ·  •  ·  •  ·  g  ·  •  ·  •  ·  g  ·  •  ·  •  ·  g  ·  •  ·  ·  ·  ·  ·  ·  ·  │ A
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │ B
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │ C
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │ D
│ S  ·  S  ·  r  ·  S  ·  S  ·  r  ·  S  ·  S  ·  r  ·  S  ·  S  ·  r  ·  ·  ·  ·  ·  ·  ·  │ E
├═══════════════════════ CENTER CHANNEL ════════════════════════════════════════════════════┤
│ x  ·  x  ·  r  ·  x  ·  x  ·  r  ·  x  ·  x  ·  r  ·  x  ·  x  ·  r  ·  ·  ·  ·  ·  ·  ·  │ F
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │ G
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │ H
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │ I
│ ·  ·  ·  ·  a  ·  ·  ·  ·  ·  a  ·  ·  ·  ·  ·  a  ·  ·  ·  ·  ·  a  ·  ·  ·  ·  ·  ·  ·  │ J
├───────────────────────────────────────────────────────────────────────────────────────────┤
│ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  │  + rail  UNUSED
│ ·  ·  ·  ·  k  ·  ·  ·  ·  ·  k  ·  ·  ·  ·  ·  k  ·  ·  ·  ·  ·  k  ·  ·  ·  ·  ·  L  ·  │  − rail  GND
└───────────────────────────────────────────────────────────────────────────────────────────┘
  └─SW1─┘    LED1   └─SW2─┘    LED2   └─SW3─┘    LED3   └─SW4─┘    LED4   ← spare cols

  S = switch pin, WIRED        x = switch pin, unused (nothing connects here)
  r = resistor leg             a = LED anode  (LONG leg)
  • = signal jumper            k = LED cathode (SHORT leg)
  g = ground jumper            L = rail link      G = Metro GND
```

Components occupy columns 30 down to 8. Columns 7–1 are free.

### Wire list — 14 wires total

**Ground first.** Two wires, and everything else depends on them:

| # | From | To |
|---|---|---|
| 1 | Metro `GND` | top `−` rail (diagram shows column 1) |
| 2 | top `−` rail (column 2) | bottom `−` rail (column 2) |

**Switches** — body straddles the channel; only the row **E** pins are wired:

| Switch | Wired pins | Unused pins | Signal wire | Ground wire |
|---|---|---|---|---|
| SW1 | `E30` `E28` | `F30` `F28` | `A30` → Metro `D2` | `A28` → top `−` rail |
| SW2 | `E24` `E22` | `F24` `F22` | `A24` → Metro `D3` | `A22` → top `−` rail |
| SW3 | `E18` `E16` | `F18` `F16` | `A18` → Metro `D4` | `A16` → top `−` rail |
| SW4 | `E12` `E10` | `F12` `F10` | `A12` → Metro `D7` | `A10` → top `−` rail |

**LEDs** — the resistor stands across the center channel, one leg each side. 0.3" is a natural span for a ¼W resistor, and it bridges the column's two nodes using only one column:

| LED | Resistor legs | Signal wire | **LED long leg (anode)** | **LED short leg (cathode)** |
|---|---|---|---|---|
| LED1 | `E26` and `F26` | `A26` → Metro `D5` | `J26` | bottom `−` rail, hole nearest col 26 |
| LED2 | `E20` and `F20` | `A20` → Metro `D6` | `J20` | bottom `−` rail, nearest col 20 |
| LED3 | `E14` and `F14` | `A14` → Metro `D9` | `J14` | bottom `−` rail, nearest col 14 |
| LED4 | `E8` and `F8` | `A8` → Metro `D11` | `J8` | bottom `−` rail, nearest col 8 |

The cathode goes **straight into a rail hole** — no jumper. A 3mm LED's legs are ~25 mm, easily long enough to reach from row J to the rail below it.

### Trace it to confirm nothing is shorted

**SW1:** `D2` → jumper → node A–E col 30 → **pin E30** → *(switch closes)* → **pin E28** → node A–E col 28 → jumper from `A28` → top `−` rail → GND.

Pins `F30` and `F28` sit in the bottom-half nodes of columns 30 and 28. Nothing else is wired to those nodes, so it does not matter whether the switch connects them, switches them, or leaves them floating. **This layout is immune to whatever the bottom pins do.**

**LED1:** `D5` → jumper → node A–E col 26 → **resistor** → node F–J col 26 → **LED anode at J26** → through the LED → **cathode in the − rail** → GND.

Each component bridges two *different* nodes. The resistor's legs are on opposite sides of the channel; the LED's legs are in a column node and a rail. Nothing is parallel to itself.

---

## 5. Switch behavior — fixed orientation

### What your switches actually do

Measured, not assumed. With the switch seated straddling the channel at columns 30 and 28 — the only orientation it physically fits:

| Probe | Released | Pressed |
|---|---|---|
| `A30` ↔ `A28` — both in row E, two columns apart | **silent** | **beeps** |
| `A30` ↔ `J30` — across the channel | **silent** | **silent** |

So the working contact runs **along the top row**, and there is no connection across the gap at all. Both the 6mm and the 4mm switches behave this way.

That's the opposite of the classic 6mm tactile, where the two pins on each side are permanently bonded and the press bridges *across* the 0.3" dimension. Yours bond nothing across the channel; the press closes the 0.2" dimension instead. Since the body only seats one way, there's no rotating out of it — so the layout is built around it rather than fighting it.

### Why this is actually the better case

Two useful consequences:

**No short-to-ground failure mode.** In the classic arrangement, a switch rotated 90° hard-wires the signal pin to ground and the input reads LOW forever — a failure that looks exactly like a firmware bug. Your switches can't do that, because nothing is permanently connected between any two pins.

**The bottom half is inert.** Only the row E pins carry signal. `F30` and `F28` connect to nothing, so the circuit behaves identically regardless of what's going on down there.

### Confirm before wiring

Still worth thirty seconds per switch, mostly to catch a dead part:

1. Seat the switch straddling the channel at its two columns.
2. Multimeter on continuity, probe `A{n}` ↔ `A{n+2}` — the two columns the switch spans.
3. **Silent when released, beeps when pressed.** That's the whole test.
4. If it beeps while released, that switch is faulty or its pins are bridged — set it aside.

Repeat at each position: `A30`↔`A28`, `A24`↔`A22`, `A18`↔`A16`, `A12`↔`A10`.

### Mixed switch sizes

You have 2 × 4mm and 2 × 6mm. They wire identically now that both behave the same way, but they have different actuation force and travel. Whack-A-Mole is a reaction game where you're timing presses — inconsistent feel between positions is exactly the kind of thing that registers as "this feels wrong" without being obvious why, and `behavior.md` calls feel out explicitly.

Put matched pairs at symmetric positions: **4mm at SW1 and SW4, 6mm at SW2 and SW3**. Four matching switches would be better still, whenever you get around to it.

---

## 6. Assembly notes

**LED polarity, restated because it's the most common silent failure.** The **long leg is the anode** and goes into row J. The **short leg is the cathode** — it also has a flat spot on the plastic rim — and goes into the bottom negative rail. Backwards LEDs don't blow up, they just stay dark. Check before you spend an hour debugging firmware.

**Resistors have no polarity.** Either leg in either hole.

**Both positive rails stay empty.** Nothing in this circuit connects to 5V, 3V3, or VIN.

**Build the two ground wires first** — the Metro `GND` wire and the rail-to-rail link. Everything references them, and a missing ground produces symptoms that look like a dozen other problems. The link is easy to forget because the board looks finished without it: the LEDs will work and every switch will read as permanently pressed.

---

## 7. Bring-up checklist

In order. Each step isolates a different failure, so when something breaks you know what changed.

1. **Power only.** USB in, nothing else connected. Onboard power LED on? Does `ls /dev/cu.*` show `/dev/cu.usbserial-ADAOKdPsS`?
2. **Blink.** Flash `fidget_blink/` to the onboard `D13` LED. Confirms toolchain and bootloader before any wiring exists.
3. **Ground wires.** Metro `GND` → top `−` rail, and top `−` → bottom `−`. Confirm with the multimeter: continuity from a *bottom* rail hole back to a Metro GND pin. That checks both wires at once.
4. **One LED.** Wire LED1 only — resistor `E26`/`F26`, anode `J26`, cathode into the bottom rail, jumper `A26` → `D5`. Drive `D5` high. Confirms resistor value, LED polarity, and ground in one shot.
5. **All four LEDs.** Sweep `D5, D6, D9, D11`. Any dark LED is a wiring problem, not a code problem — you already know the pattern works from step 4.
6. **One switch.** Continuity-test SW1 per §5, seat it, wire signal and ground, print its state with `tinygo monitor`. Reads `1` idle, `0` pressed. A constant `1` means the ground path is broken; a constant `0` means signal and ground are bridged.
7. **All four switches**, reading independently.
8. **Full firmware.**

---

## 8. Power notes

**Use USB for all prototyping.** Simpler, and you get serial output on the same cable.

The barrel jack accepts **7–9V** center-positive, and 9V is recommended — your 9V 1A supply is correct. It feeds a linear regulator, so at 9V in and 5V out roughly 44% of the input power leaves as heat. That's normal, but it means a 9V alkaline battery would run this for a couple of hours at best. Fine for a demo, not for something you carry. An untethered version wants a LiPo and a charger board.
