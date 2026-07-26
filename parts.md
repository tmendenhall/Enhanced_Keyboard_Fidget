# parts.md — Inventory

> Fill this in. Anything you don't have, write `NEED`. Anything you're unsure of, write `?`.
> I'll adjust the schematic and firmware to match whatever's actually here.

---

## 1. Controller — mostly known already

Confirmed: **ATmega328P, 16 MHz, 5V logic, 20 digital I/O, USB programming.** TinyGo target `arduino-uno`.

Two things still worth filling in:

| Field | Value |
|---|---|
| Exact board name (silkscreen text) | `Metro` |
| How many of these boards do you have? | 2 |

**Serial port check:** run `ls /dev/cu.*` before and after plugging the board in — note what appears.

/dev/cu.usbserial-ADAOKdPsS
---

## 2. Switches (prototype stand-ins)

| # | Type | Poles | Momentary or latching | Qty on hand | Notes |
|---|---|---|---|---|---|
| 2 | 4 mm tactile pushbutton | SPST(?) | momentary | | |
| 2 | 6mm tactile pushbutton| SPST(?) | | | |

- Breadboard-friendly leads (0.1" pitch)? yes 
- If tactile: 4-leg (4 pins) plugged into breadboard with both legs on the same line. like ::
---

## 3. LEDs

| # | Color | Package (3mm/5mm/SMD) | Forward voltage (Vf) if known | Qty |
|---|---|---|---|---|
| 1 | Red |3mm | | 4 |
| 2 | | | | |
| 3 | | | | |
| 4 | | | | |

- Common-anode/common-cathode RGB units on hand? no
- Any NeoPixel / WS2812 strips or rings?: no 

---

## 4. Resistors

List every value and rough quantity. Values in the 100Ω–1kΩ range are what matter most.

| Value | Qty | Tolerance |
|---|---|---|
| brown black red | 10 | 5% |
| yellow black red | 12 | 5% |
| brown black orange | 15 | 5% |
| orange orange brown | 15 | 5% |

- Do you have a resistor kit (assorted book)?  no

---

## 5. Prototyping hardware

| Item | Have? | Details |
|---|---|---|
| Breadboard | yes | 1 (300 pt) / mini |
| Male–male jumper wires |yes | 30 |
| Male–female jumper wires | | 10 |
| Wire strippers |yes | 1|

---

## 6. Power

| Item | Have? | Details |
|---|---|---|
| USB-A → micro USB cable (**data**, not charge-only) | yes | 6 in |
| Barrel jack supply, 9V center-positive | yes | 9v 1 amp |

---

## 7. Test & debug gear

| Item | Have? | Notes |
|---|---|---|
| Multimeter | yes | continuity beeper |

---

## 8. Optional / stretch

| Item | Have? | Notes |
|---|---|---|
| Real MX-style keyboard switches | no | the eventual target |
| Keycaps |  no| |

---

## 9. Anything else in the bin

* Many more LEDs if necessary Red and Green
