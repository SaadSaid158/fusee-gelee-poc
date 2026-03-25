# Fusée Gelée PoC (CVE-2018-6242)

Go-based implementation of the Fusée Gelée vulnerability — a coldboot exploit in the NVIDIA Tegra X1 BootROM, affecting Nintendo Switch units manufactured before mid-2018.

## Prerequisites

- Go 1.19+
- libusb-1.0 (`sudo apt install libusb-1.0-0-dev` / `brew install libusb`)
- A first-generation Nintendo Switch (unpatched Tegra X1 bootrom)
- Device in RCM mode (hold VOL+ and press HOME with RCM jig inserted)

## Build

```bash
go build -o fusee-gelee ./cmd/fusee-gelee
```

Cross-compile examples:

> **Note:** `gousb` uses cgo and links against libusb, so cross-compilation requires a cross-toolchain **and** target-arch libusb headers. Plain `GOOS/GOARCH` without the correct `CC` and sysroot will fail with `undefined` symbol errors.

```bash
# Linux arm64 (requires gcc-aarch64-linux-gnu + libusb-1.0-0-dev:arm64)
sudo apt install gcc-aarch64-linux-gnu libusb-1.0-0-dev:arm64
CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o fusee-gelee-arm64 ./cmd/fusee-gelee

# Windows amd64 (requires MinGW + libusb Windows binaries)
# Easiest to build natively on Windows with MSYS2 + libusb installed
GOOS=windows GOARCH=amd64 go build -o fusee-gelee.exe ./cmd/fusee-gelee
```

## Usage

```bash
sudo ./fusee-gelee        # sudo required for raw USB access on Linux
```

On Linux you can avoid sudo by adding a udev rule:

```
SUBSYSTEM=="usb", ATTRS{idVendor}=="0955", ATTRS{idProduct}=="7321", MODE="0666"
```

## Project structure

```
cmd/fusee-gelee/main.go         — entry point + all menu actions
internal/tui/                   — colours, progress bar, interactive menu
internal/config/config.go       — JSON config + favourites persistence
internal/payload/manager.go     — payload registry, download, SHA256 verify
internal/device/manager.go      — multi-device USB detection and selection
internal/usb/rcm.go             — USB/RCM protocol, exploit trigger
internal/exploit/payload.go     — payload construction, intermezzo, stack spray
```

## Technical details

CVE-2018-6242 is a buffer overflow in the Tegra X1 BootROM USB recovery mode (RCM).
The bootrom reads a 4-byte length field from the incoming payload and passes it
directly to a DMA engine without bounds-checking it. By setting this field to
`0x7000` while the DMA bounce buffer is only `0x1000` bytes, `0x6000` bytes of
IRAM stack are overwritten. The saved link register (LR) is replaced with the
address of a small ARM64 intermezzo stub embedded in the payload, which then
branches to the user's payload (Hekate, Atmosphere, etc.).

The vulnerability is in mask ROM and cannot be patched by software. Nintendo
mitigated it in hardware on units produced from mid-2018 onwards (the "Mariko" /
"patched" revision).

## Learning Resources

- [Original Fusée Gelée Paper](https://github.com/Qyriad/fusee-launcher/blob/master/report/fusee_gelee.md)
- [NVIDIA Security Bulletin](http://nvidia.custhelp.com/app/answers/detail/a_id/4660)
- [NVD Entry](https://nvd.nist.gov/vuln/detail/CVE-2018-6242)

## Legal

This software is intended for use on hardware you own. The author assumes no responsibility for any damage resulting from its use. Distributed under the MIT License — see LICENSE for details.