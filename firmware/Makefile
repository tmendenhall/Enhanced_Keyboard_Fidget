# Adafruit Metro 328 (ATmega328P) -> TinyGo target "arduino-uno"
TARGET ?= arduino-uno

# Override if auto-detection picks the wrong device:
#   make flash PORT=/dev/cu.usbserial-1420
PORT ?=
PORTFLAG = $(if $(PORT),-port=$(PORT),)

.PHONY: build flash monitor size fmt clean

build:
	tinygo build -o build/fidget.hex -target=$(TARGET) .

flash:
	tinygo flash -target=$(TARGET) $(PORTFLAG) .

monitor:
	tinygo monitor $(PORTFLAG)

# Watch the RAM column. The '328 has 2KB RAM and 32KB flash.
size:
	tinygo build -o /dev/null -target=$(TARGET) -size short .

# Per-package breakdown — use this when something needs trimming.
size-full:
	tinygo build -o /dev/null -target=$(TARGET) -size full .

fmt:
	gofmt -w .

clean:
	rm -rf build
