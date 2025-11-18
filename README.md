# Fusée Gelée PoC (CVE-2018-6242)

A Proof of Concept implementation in Go demonstrating the Fusée Gelée vulnerability (CVE-2018-6242) - a coldboot vulnerability in NVIDIA Tegra X1 BootROM that affects devices like the Nintendo Switch.

## ⚠️ Disclaimer

This project is for **educational and research purposes only**. This PoC demonstrates a known vulnerability for security research and portfolio purposes. Only use this on devices you own. The author is not responsible for any misuse.

## 🔍 About CVE-2018-6242

CVE-2018-6242 (Fusée Gelée) is a buffer overflow vulnerability in the USB recovery mode (RCM) of NVIDIA Tegra X1 processors. The vulnerability exists in the BootROM, which cannot be patched, affecting all Tegra X1 devices manufactured before mid-2018.

### Technical Details
- **Type**: Buffer Overflow (CWE-119)
- **Attack Vector**: Physical (USB)
- **CVSS Score**: 6.8 (Medium)
- **Affected**: NVIDIA Tegra X1 BootROM (Nintendo Switch, NVIDIA Shield TV, etc.)

## 🚀 Features

- USB communication with Tegra devices in RCM mode
- Buffer overflow exploit implementation
- Custom payload injection
- Success image display on Nintendo Switch screen
- Cross-platform support (Windows, Linux, macOS)

## 📋 Prerequisites

- Go 1.19 or higher
- libusb (for USB communication)
- A vulnerable NVIDIA Tegra X1 device (Nintendo Switch with unpatched bootrom)
- Device must be in RCM (Recovery Mode)

### Installing libusb

**Linux:**
```bash
sudo apt-get install libusb-1.0-0-dev
```

**macOS:**
```bash
brew install libusb
```

**Windows:**
Download and install from [libusb.info](https://libusb.info/)

## 🛠️ Installation

```bash
git clone https://github.com/SaadSaid158/fusee-gelee-poc.git
cd fusee-gelee-poc
go mod download
go build -o fusee-gelee ./cmd/fusee-gelee
```

## 📖 Usage

1. Put your Nintendo Switch into RCM mode:
   - Power off the device
   - Hold Volume Up + Power while shorting pin 10 (or use an RCM jig)
   - Connect via USB

2. Run the exploit:
```bash
# Linux/macOS (may require sudo for USB access)
sudo ./fusee-gelee

# Windows (run as Administrator)
fusee-gelee.exe
```

3. If successful, you should see a success image on the Switch screen!

## 📁 Project Structure

```
fusee-gelee-poc/
├── cmd/
│   └── fusee-gelee/
│       └── main.go          # Entry point
├── internal/
│   ├── usb/
│   │   └── rcm.go           # USB RCM communication
│   ├── exploit/
│   │   └── payload.go       # Exploit implementation
│   └── display/
│       └── image.go         # Success image payload
├── payloads/
│   └── success.bin          # Display payload binary
├── go.mod
├── go.sum
└── README.md
```

## 🔬 How It Works

1. **Device Detection**: Scans for Tegra devices in RCM mode (VID: 0x0955, PID: 0x7321)
2. **Exploit Trigger**: Sends a specially crafted USB control transfer with an oversized length field
3. **Buffer Overflow**: Overflows the DMA buffer in BootROM, allowing arbitrary code execution
4. **Payload Injection**: Injects a custom payload that displays a success image
5. **Execution**: Device executes the injected payload, bypassing secure boot

## 🎓 Learning Resources

- [Original Fusée Gelée Paper](https://github.com/Qyriad/fusee-launcher/blob/master/report/fusee_gelee.md)
- [NVIDIA Security Bulletin](http://nvidia.custhelp.com/app/answers/detail/a_id/4660)
- [NVD Entry](https://nvd.nist.gov/vuln/detail/CVE-2018-6242)

## 📝 License

MIT License - See LICENSE file for details

## 🤝 Contributing

This is a portfolio/educational project. Feel free to fork and experiment!

## 👤 Author

**Saad Said**
- GitHub: [@SaadSaid158](https://github.com/SaadSaid158)

---

*Built with Go for educational purposes. Stay curious, stay ethical! 🔐*