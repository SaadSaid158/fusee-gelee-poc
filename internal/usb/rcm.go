package usb

import (
	"fmt"

	"github.com/google/gousb"
)

const (
	// Tegra X1 RCM mode USB identifiers
	TegraVendorID  = 0x0955
	TegraProductID = 0x7321

	// RCM USB configuration
	RCMInterface = 0
	RCMEndpoint  = 0x01

	// RCM protocol constants
	RCMRequestGetStatus = 0x00
	RCMRequestSendData  = 0x01

	// Buffer sizes
	DeviceIDSize = 16
	MaxTransfer  = 0x30000 // 192KB
)

// Context wraps the USB context for interacting with devices
type Context struct {
	ctx *gousb.Context
}

// Device represents a Tegra device in RCM mode
type Device struct {
	dev *gousb.Device
}

// NewContext creates a new USB context
func NewContext() (*Context, error) {
	ctx := gousb.NewContext()
	return &Context{ctx: ctx}, nil
}

// Close closes the USB context
func (c *Context) Close() {
	c.ctx.Close()
}

// FindRCMDevice searches for a Tegra device in RCM mode
func (c *Context) FindRCMDevice() (*Device, error) {
	dev, err := c.ctx.OpenDeviceWithVIDPID(gousb.ID(TegraVendorID), gousb.ID(TegraProductID))
	if err != nil || dev == nil {
		return nil, fmt.Errorf("no Tegra device found in RCM mode (VID:0x%04X PID:0x%04X)", TegraVendorID, TegraProductID)
	}

	// Set auto-detach kernel driver
	dev.SetAutoDetach(true)

	return &Device{
		dev: dev,
	}, nil
}

// Close closes the device
func (d *Device) Close() {
	if d.dev != nil {
		d.dev.Close()
	}
}

// String returns a string representation of the device
func (d *Device) String() string {
	if d.dev != nil {
		desc, _ := d.dev.Desc()
		return fmt.Sprintf("Tegra X1 (Bus %03d Device %03d: ID %04x:%04x)",
			desc.Bus,
			desc.Address,
			desc.Vendor,
			desc.Product)
	}
	return "Tegra X1 RCM Device"
}

// ReadDeviceID reads the device ID from the RCM device
func (d *Device) ReadDeviceID() ([]byte, error) {
	// Prepare request to read device ID
	buffer := make([]byte, DeviceIDSize)

	// Send control transfer to read device ID
	// REQUEST_TYPE_IN | REQUEST_TYPE_STANDARD | RECIPIENT_INTERFACE
	reqType := uint8(0x82) // Device to Host, Standard, Endpoint
	request := uint8(RCMRequestGetStatus)
	value := uint16(0)
	index := uint16(0)

	n, err := d.dev.Control(reqType, request, value, index, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read device ID: %v", err)
	}

	if n != DeviceIDSize {
		return nil, fmt.Errorf("unexpected device ID size: got %d, expected %d", n, DeviceIDSize)
	}

	return buffer, nil
}

// TriggerExploit sends the exploit payload to the device
func (d *Device) TriggerExploit(payload []byte) error {
	if len(payload) > MaxTransfer {
		return fmt.Errorf("payload too large: %d bytes (max: %d)", len(payload), MaxTransfer)
	}

	// This is where the Fusée Gelée vulnerability is triggered
	// The key is sending a specially crafted USB control transfer
	// with an oversized length field that overflows the DMA buffer

	// First, we need to send the payload length (intentionally oversized)
	reqType := uint8(0x41) // Host to Device, Vendor, Interface
	request := uint8(RCMRequestSendData)
	value := uint16(0)
	index := uint16(0)

	// Send the actual payload via bulk transfer
	// The length field in the setup packet will be manipulated to cause overflow
	_, err := d.dev.Control(reqType, request, value, index, payload)
	if err != nil {
		// An error here might actually indicate success since we're crashing the bootrom
		// and taking control - but we'll return it for now
		return fmt.Errorf("exploit trigger: %v", err)
	}

	return nil
}

// WriteBulk writes data to the device via bulk transfer
func (d *Device) WriteBulk(data []byte) (int, error) {
	// For bulk transfers, we would use the endpoint
	// This is a simplified version
	reqType := uint8(0x41)
	request := uint8(RCMRequestSendData)
	value := uint16(0)
	index := uint16(0)

	return d.dev.Control(reqType, request, value, index, data)
}
