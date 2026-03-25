package device

import (
	"fmt"

	"github.com/SaadSaid158/fusee-gelee-poc/internal/tui"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/usb"
)

// RCMDevice holds a discovered device plus its index for display.
type RCMDevice struct {
	Index  int
	Label  string
	USBCtx *usb.Context
	Dev    *usb.Device
}

// Manager handles multi-device detection and selection.
type Manager struct {
	devices []*RCMDevice
}

// NewManager creates an empty device manager.
func NewManager() *Manager {
	return &Manager{}
}

// Scan probes all USB buses for Tegra X1 devices in RCM mode.
// Multiple Switches on the same host are fully supported.
func (m *Manager) Scan() error {
	m.devices = nil

	// We open one context per found device because gousb's FindDevices
	// iterates all buses — we collect all matches.
	ctx, err := usb.NewContext()
	if err != nil {
		return fmt.Errorf("USB init failed: %w", err)
	}

	// FindAllRCMDevices returns all matching Tegra devices on the bus.
	devs, err := ctx.FindAllRCMDevices()
	if err != nil {
		ctx.Close()
		return err
	}

	for i, d := range devs {
		m.devices = append(m.devices, &RCMDevice{
			Index:  i + 1,
			Label:  fmt.Sprintf("Switch #%d — %s", i+1, d.String()),
			USBCtx: ctx,
			Dev:    d,
		})
	}

	return nil
}

// Count returns the number of detected devices.
func (m *Manager) Count() int {
	return len(m.devices)
}

// List prints all detected devices to stdout.
func (m *Manager) List() {
	if len(m.devices) == 0 {
		fmt.Println(tui.Warn("No Tegra RCM devices found. Is the Switch in RCM mode?"))
		return
	}
	fmt.Println(tui.BoldS(fmt.Sprintf("  Found %d device(s):", len(m.devices))))
	for _, d := range m.devices {
		fmt.Printf("    %s %s\n",
			tui.Colorf(tui.BrightYellow, "[%d]", d.Index),
			tui.Colorf(tui.BrightWhite, "%s", d.Label),
		)
	}
}

// Select returns a device by 1-based index.
func (m *Manager) Select(index int) (*RCMDevice, error) {
	if index < 1 || index > len(m.devices) {
		return nil, fmt.Errorf("invalid device index %d", index)
	}
	return m.devices[index-1], nil
}

// SelectInteractive shows a list and prompts the user to pick one.
// If there is exactly one device, it is returned automatically.
func (m *Manager) SelectInteractive() (*RCMDevice, error) {
	if len(m.devices) == 0 {
		return nil, fmt.Errorf("no RCM devices found")
	}
	if len(m.devices) == 1 {
		fmt.Println(tui.Info(fmt.Sprintf("Using %s", m.devices[0].Label)))
		return m.devices[0], nil
	}

	labels := make([]string, len(m.devices))
	for i, d := range m.devices {
		labels[i] = d.Label
	}

	idx := tui.SelectFromList("Select device", labels)
	if idx < 0 {
		return nil, fmt.Errorf("device selection cancelled")
	}
	return m.devices[idx], nil
}

// CloseAll releases all open device handles and USB contexts.
func (m *Manager) CloseAll() {
	seen := map[*usb.Context]bool{}
	for _, d := range m.devices {
		if d.Dev != nil {
			d.Dev.Close()
		}
		if d.USBCtx != nil && !seen[d.USBCtx] {
			d.USBCtx.Close()
			seen[d.USBCtx] = true
		}
	}
	m.devices = nil
}
