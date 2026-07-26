# behavior.md — Behavior Modes

---

## Naming

Switches and LEDs are numbered **1–4**, left to right as you face the fidget.
LED *n* sits directly with switch *n*.

- `SW1..SW4` — the switches
- `L1..L4` — the LEDs

---

## 0. Power-on sequence

What happens in the first second or two after power is applied, before the device is "ready"?

| Step | Duration | LED behavior | Notes |
|---|---|---|---|
| 1 | 100ms | L1 on | then L1 off |
| 2 | 100ms | L2 on| then L2 off |
| 3 | 100ms | L3 on| then L3 off |
| 4 | 100ms | L4 on | then L4 off |

- No mode will be active once startup is complete and all L1-L4 are off. Then depending on SW1 .. SW3 is pressed that will select the run mode.
- Should the last-used mode be remembered across power cycles? → no
  
  Add notes on what will change in the program if data needs to be stored.

---

## 1. Mode list

Modes and the switch that triggers them.

| # | Mode name | One-line description |
|---|---|---|
| 1 | Whack-A-Mole| A random Light L1-4 is On. When the corresponding switch is pressed down the light is Off then wait 50ms and repeat|
| 2 | Rave Mode | For each SW the corresponding light cycles continuously 0 → 255 → 0 using PWM for as long as the switch is held. Release turns it off. |
| 3 | Binary Counter | Start with all L-4 off. If any swicth SW 1-4 is on then off then count in binary from 0 through 15 where L4 is the LSB and L1 is the MSB, so the row reads left-to-right like a written number. Cycle back to 0 once you reach 15. |

---

## 2. How you change modes

Once POST is complete SW1 activates mode 1, SW2 is 2 and SW 3 is 3.

Holding all 4 SW1-SW4 down for the full hold duration goes back to the starting state. When all 4 switches are off then the next swtich press indicates the mode.

Details:
- Hold duration: 5000 ms
- **Rave mode exception:** the 5000 ms hold clock does not start until every LED has reached full brightness (255) **at least once** since its switch went down. Holding all four switches *is* normal Rave play, so the reset can't be allowed to fire while the lights are still on their first climb. End to end in Rave this is ~2.55 s to the first peak + 5 s of hold ≈ 7.6 s. Releasing any switch aborts the gesture and clears the peaked flags.

  *Note the "at least once" wording.* Now that Rave cycles continuously, the four lights are almost never at 255 simultaneously — their phases depend on exactly when each switch was pressed. Requiring all four at full at the same instant would make the reset nearly impossible to trigger. A latched "has peaked" flag per channel keeps the original intent — don't fire during the first climb — without demanding an alignment that won't happen.

### Reset confirmation sweep

The moment the hold duration is satisfied, before returning to the starting state, cycle the LEDs **in reverse** — the opposite direction of the power-on test:

| Step | Duration | LED behavior |
|---|---|---|
| 1 | 100ms | L4 on, then L4 off |
| 2 | 100ms | L3 on, then L3 off |
| 3 | 100ms | L2 on, then L2 off |
| 4 | 100ms | L1 on, then L1 off |

Then all LEDs off, no mode active, waiting for all four switches to be released before the next press selects a mode.

- Timing is configurable and defaults to the same value as the power-on sweep (100 ms per LED, 400 ms total).
- The sweep plays while the switches are still held. It means *done, let go*.
- **Direction carries the meaning.** Forwards (L1→L4) = just powered on. Backwards (L4→L1) = just reset. Two events that leave the device in the same state, distinguished by animation direction alone.


---

## 3. Mode detail

Copy this block once per mode.

### Mode 1 — Whack-a-Mole

**Idle behavior** (no switch touched, for a while): None

A random Light L1-4 is On. When the corresponding switch is pressed down the light is Off then wait 50ms and repeat

---

### Mode 2 — Rave Mode

**Idle behavior: None

For each SW(N) the corresponding light L(N) is lit using PWM from 0 to 255 brightness.

**While the switch is held, the light cycles continuously: 0 → 255 → 0, then repeats.** The brightness level changes by 1 every 10ms in whichever direction it is currently travelling, reversing at each end. When the switch is released the light turns off immediately and the cycle resets to 0, so the next press always starts from dark.

The lights L1-L4 are independent — each has its own level and its own direction, and each starts its cycle from 0 when its switch goes down.

Timing:

| Leg | Duration |
|---|---|
| 0 → 255 (rising) | 2.55 s |
| 255 → 0 (falling) | 2.55 s |
| Full cycle | **5.1 s** |

The triangle is symmetric — rise and fall share the same 10ms step. No pause at the top or bottom.

---

### Mode 3 — Binary Counter

Start with all Lights L1-4 off. If any swicth SW 1-4 is clicked on then off then increment and display in binary from 0 through 15 where **L4 is the LSB and L1 is the MSB**. Cycle back to 0 once you reach 15.

The LEDs read left to right exactly as you'd write the number down: L1 L2 L3 L4 = bit3 bit2 bit1 bit0. Counting up, the rightmost light toggles fastest.

Examples

|L1|L2|L3|L4|value|
|---|---|---|---|---|
|0|0|0|0|0|
|0|0|0|1|1|
|0|0|1|0|2|
|0|0|1|1|3|
|0|1|0|0|4|
|1|0|0|0|8|
|1|1|1|1|15|

---

### Mode 4 — <name>

**Idle behavior:**

**On press:**

**On hold:**

**On release:**

**Multi-key behavior:**

**Timing / speeds:**

**Anything else:**

---

## 4. Cross-cutting behavior

Things that apply regardless of mode.

| Question | Answer |
|---|---|
| Debug over USB serial? | yes show switch on and off and current mode |
| Should it act as a USB HID keyboard eventually? | no |

---

## 5. Feel and intent

Free-form — this is the part that's hardest to infer from a spec and most useful to me.

- Anything that would make it feel *wrong*?
Yes it will feel wrong if the responses are too fast or too slow. The timing values are estimates and may be changed. Structure the code so that the values are easy to change to dial in the feeling.

At any point if all 4 switches are held down for 5 seconds (5000ms) then reset the state to right after power on when the next switch selects the mode. In Rave mode the 5000ms only begins counting once all four lights have reached full brightness.

When the hold completes, the LEDs sweep in reverse — L4, L3, L2, L1 — so it is obvious the reset registered and you can let go. See §2.

---

## 6. Out of scope for the prototype

Things you want eventually but explicitly not in v1.
Simon mode.

Produce a random set of light patterns. Display the pattern in full and wait. The user needs to trigger the corresponding switch for each light in the same order as displayed. Once a complete pattern has been reproduced then wait 100ms and display a new pattern adding one additional light until you reach a maximum of 15 lights shown in order. If the user does not reproduce the complete pattern then blink all of the lights on and off. 50ms on and 50 ms off for 3 blink cycles and start the game over.
