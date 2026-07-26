# Toolchain Setup — Go on an ATmega328 Metro, from a MacBook (arm64)

**Board:** Adafruit Metro 328 — ATmega328P, 16 MHz, 5V logic, 20 digital I/O, USB programming.
**TinyGo target:** `arduino-uno` (same MCU, same bootloader, same pin mapping as the Uno).

---

## Read this before you install anything

You picked Go, and Go will work here — but you should know what you're signing up for. The ATmega328P has **32KB of flash and 2KB of RAM**, and TinyGo's AVR backend is the least mature of its targets.

**What you give up on this chip:**

- Goroutines and channels. Technically present, practically unusable at 2KB.
- `fmt`. Use the `println` builtin instead — it works, it's just crude.
- Heap allocation. Anything that allocates (slices that grow, string concatenation, interfaces holding large values) will bite you. Write it like C: fixed-size arrays, no dynamic growth.
- Floating point. Software-emulated, slow, and fat in flash. Use integer math.
- Comfortable margins. You will watch the size output.

**What still works well:** GPIO, timers, the `machine` package, `time`, structs, methods, real type safety, and a compiler that catches things Arduino C++ won't. The starter firmware in `firmware/` is written under all of the above constraints and should land well within budget.

**When to bail out to Arduino C++:** if you hit a RAM wall you can't design around, or if the AVR backend produces something that misbehaves in a way you can't explain. That's not a rewrite — the firmware is deliberately structured to translate almost line-for-line. It's a fallback, not a failure.

Check your budget after every meaningful change:

```bash
tinygo build -o /dev/null -target=arduino-uno -size short .
```

Watch the RAM column. Under ~1.5KB and you're comfortable; past that, start trimming.

---

## Step 1: Go

```bash
brew install go
go version

go version go1.26.5 darwin/arm64
```

TinyGo needs a system Go installation (1.23 or newer for current TinyGo releases) because it compiles against Go's standard library sources.

---

## Step 2: TinyGo

```bash
brew tap tinygo-org/tools
brew trust --formula tinygo-org/tools/tinygo
brew install tinygo
tinygo version

tinygo version 0.41.1 darwin/arm64 (using go version go1.26.5 and LLVM version 20.1.1)
```

The Homebrew formula ships a native arm64 build — no Rosetta.

---

## Step 3: avrdude

This is what actually pushes bytes to the AVR bootloader. TinyGo shells out to it.

```bash
brew install avrdude
avrdude --version

Avrdude version 8.2
Copyright see https://github.com/avrdudes/avrdude/blob/main/AUTHORS
Use https://github.com/avrdudes/avrdude/issues to report bugs and ask questions

```

---

## Step 4: USB-serial driver (only if the port doesn't appear)

The Metro 328 talks to your Mac through a USB-to-serial bridge chip rather than native USB. Plug the board in and check:

```bash
ls -l /dev/cu.*

/dev/cu.usbserial-ADAOKdPsS
```

You want a new entry — typically `/dev/cu.usbserial-XXXX` or `/dev/cu.SLAB_USBtoUART`.

**If nothing new appears:**

1. **Try a different USB cable first.** Charge-only micro-USB cables are extremely common and they look identical to data cables. This is the single most frequent cause of "my board is dead." Rule it out before anything else.
2. Then check which bridge chip your board has — it's the small chip near the USB jack:
   - **CP2102 / CP2104** (Adafruit's usual choice) → install the [Silicon Labs CP210x VCP driver](https://www.silabs.com/developer-tools/usb-to-uart-bridge-vcp-drivers)
   - **CH340 / CH341** (common on clones) → install the WCH CH34x macOS driver
   - **ATmega16U2** (genuine Arduino Uno R3) → no driver needed, it's USB-CDC

You'll also need a **USB-C to USB-A adapter or a USB-C to micro-USB cable**, since the MacBook has no USB-A ports. Worth confirming you have one before you start.

---

## Step 5: GoLand setup

**The problem this solves:** GoLand's indexer resolves `import "machine"` against your *system* Go installation, where no such package exists. Until you tell it otherwise, `machine.LED` and `machine.PinConfig` show as unresolved, autocomplete is dead, and the file sits underlined in red — even though `tinygo build` compiles it without complaint. The `machine` package lives inside TinyGo's own cached GOROOT, and its contents change depending on which target you're building for.

### Install the plugin

JetBrains publishes an official TinyGo plugin. GoLand 2021.1.2 or newer.

**Settings** (`⌘,`) → **Plugins** → **Marketplace** → search `TinyGo` → **Install** → restart GoLand.

### Point it at your toolchain

**Settings** → **Languages & Frameworks** → **Go** → **TinyGo**

| Field | Value |
|---|---|
| Enable TinyGo support | checked |
| TinyGo path | `/opt/homebrew/opt/tinygo` |
| Target | `arduino-uno` |

Confirm the path first — Homebrew moves things around between versions:

```bash
brew --prefix tinygo
```

Selecting the target is the step that matters. The plugin reads TinyGo's build tags and cached GOROOT for `arduino-uno` and feeds them to the indexer, which is what makes `machine` resolve. Change the target later and you must change it here too, or GoLand will happily autocomplete pins that don't exist on your chip.

### Open the right directory

Open `firmware/` (or `fidget_blink/`) as the project root — **not** the parent `enhanced_keyboard_fidget/` folder. GoLand wants `go.mod` at the root. Opening the parent gives you a project with no module and a lot of spurious errors.

Since this repo has two independent modules, either open them as separate projects or add the second one via **Settings** → **Go** → **Go Modules** → **+**.

### Run configuration

The plugin adds a **TinyGo** run configuration type:

**Run** → **Edit Configurations** → **+** → **TinyGo**

| Field | Value |
|---|---|
| Command | `flash` |
| Target | `arduino-uno` |
| Working directory | your module root |

That gives you `⌃R` to build and upload. The terminal (`make flash`) works just as well — use whichever you prefer.

### If `machine` is still red

Verify the plugin actually applied the target:

```bash
tinygo info -target=arduino-uno
```

Note the `build tags:` and `cached GOROOT:` lines. Then check that **Settings** → **Go** → **Build Tags & Vendoring** → *Custom tags* matches. You can paste them in by hand if the plugin didn't populate them — that's the manual equivalent of what the plugin does, and it's a reasonable fallback.

Then **File** → **Invalidate Caches** → *Invalidate and Restart*. GoLand caches resolution aggressively and often needs the nudge after a toolchain change.

---

## Step 6: Verify with a blink

The blink project already exists at `fidget_blink/`.

```bash
cd fidget_blink
```

`fidget_blink/main.go`:

```go
package main

import (
	"machine"
	"time"
)

func main() {
	led := machine.LED // D13
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		led.High()
		time.Sleep(300 * time.Millisecond)
		led.Low()
		time.Sleep(300 * time.Millisecond)
	}
}
```

**Module setup — once, not every build.** TinyGo requires the code to live in a Go module, so there must be a `go.mod` beside `main.go`. This repo already has one; if you're starting a fresh directory:

```bash
go mod init blink     # creates go.mod — run once, ever
```

**Build and upload — this is the actual compile command:**

Uses the output from earlier
```bash

tinygo flash -target=arduino-uno --port /dev/cu.usbserial-ADAOKdPsS .
```

`go mod init` is project setup. `tinygo flash` is the compiler. Don't re-run the first one.

Note that plain `go build` will **not** work on this code — the standard Go toolchain has no `machine` package and no AVR backend. Every build goes through `tinygo`.

**Do not skip this step.** Once the onboard LED blinks, the toolchain, the bootloader, the cable, the driver, and the pin mapping are all confirmed good — which means every problem from that point on is in your wiring or your code. That certainty is worth ten minutes.

### If flashing fails

**`avrdude: stk500_recv(): programmer is not responding`** — the bootloader window was missed. Press RESET and immediately re-run the flash command. On some boards the auto-reset capacitor is flaky and manual timing is the fix.

**`go.mod file not found`** — you're outside a module. Run `go mod init blink` in this directory.

**Serial monitor still open** — it holds the port and blocks avrdude. Close it before flashing.

---

## Step 7: Serial monitor

```bash
tinygo monitor
```

Defaults to 115200 baud, which matches the firmware. Close it before each flash — on this board the monitor and the programmer contend for the same serial port.

---

## Complete install, copy-paste

```bash
brew install go

brew tap tinygo-org/tools
brew trust --formula tinygo-org/tools/tinygo
brew install tinygo

brew install avrdude

go version && tinygo version && avrdude --version 2>&1 | head -1
```

GoLand plugin: **Settings** → **Plugins** → **Marketplace** → `TinyGo` → Install.

---

## Everyday commands

```bash
cd firmware

tinygo flash -target=arduino-uno .                          # build + upload
tinygo build -o /dev/null -target=arduino-uno -size short . # flash/RAM usage
tinygo monitor                                              # serial output
tinygo info -target=arduino-uno                             # build tags, cached GOROOT
gofmt -w .                                                  # format
```

Or use the included `Makefile`: `make flash`, `make size`, `make monitor`.

---

## Sources

- [TinyGo — macOS install guide](https://tinygo.org/getting-started/install/macos/)
- [TinyGo — Arduino Uno target](https://tinygo.org/docs/reference/microcontrollers/boards/arduino-uno/)
- [TinyGo — arduino-uno machine package pin definitions](https://tinygo.org/docs/reference/microcontrollers/machine/arduino-uno/)
- [TinyGo — IntelliJ IDEA / GoLand integration](https://tinygo.org/docs/guides/ide-integration/intellij/)
- [JetBrains TinyGo plugin](https://plugins.jetbrains.com/plugin/16915-tinygo)
- [Adafruit METRO 328 product page](https://www.adafruit.com/product/2488)
