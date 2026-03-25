package main

import (
	"fmt"
	"os"

	"github.com/SaadSaid158/fusee-gelee-poc/internal/config"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/device"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/exploit"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/payload"
	"github.com/SaadSaid158/fusee-gelee-poc/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(tui.Warn(fmt.Sprintf("Config load failed (%v) -- using defaults", err)))
		cfg = config.Default()
	}

	pm := payload.NewManager(cfg.DownloadDir)
	dm := device.NewManager()
	defer dm.CloseAll()

	menu := tui.NewMenu("Main Menu")

	menu.Add("Inject Payload", "Send a payload to a connected Switch", func() error {
		return actionInject(pm, dm)
	})
	menu.Add("Download Payloads", "Fetch payloads from GitHub releases", func() error {
		return actionDownload(cfg, pm)
	})
	menu.Add("Detect Devices", "Scan for Tegra X1 devices in RCM mode", func() error {
		return actionDetect(dm)
	})
	menu.Add("Settings", "View and edit configuration", func() error {
		return actionSettings(cfg)
	})

	if err := menu.Run(); err != nil {
		fmt.Println(tui.Errf(err.Error()))
		os.Exit(1)
	}
}

// ── Inject ───────────────────────────────────────────────────────────────────

func actionInject(pm *payload.Manager, dm *device.Manager) error {
	var labels []string
	var payloads []payload.KnownPayload

	for _, p := range payload.Registry {
		status := "[not downloaded]"
		if pm.Exists(p) {
			status = "[ready]"
		}
		labels = append(labels, fmt.Sprintf("%-22s %s", p.Name, status))
		payloads = append(payloads, p)
	}
	labels = append(labels, "Custom file...")

	idx := tui.SelectFromList("Select payload to inject", labels)
	if idx < 0 {
		return nil
	}

	var data []byte

	if idx == len(payloads) {
		path := tui.Prompt("Path to .bin file")
		if path == "" {
			return fmt.Errorf("no path given")
		}
		var err error
		data, err = pm.LoadCustom(path)
		if err != nil {
			return err
		}
		fmt.Println(tui.Success(fmt.Sprintf("Loaded %d bytes from %s", len(data), path)))
	} else {
		p := payloads[idx]
		if !pm.Exists(p) {
			fmt.Println(tui.Warn(fmt.Sprintf("%s is not downloaded yet.", p.Name)))
			if !tui.Confirm("Download it now?") {
				return fmt.Errorf("payload not available")
			}
			if err := pm.Download(p); err != nil {
				return err
			}
		}
		var err error
		data, err = pm.Load(p)
		if err != nil {
			return err
		}
		fmt.Println(tui.Success(fmt.Sprintf("Loaded %s  (%d bytes)", p.Name, len(data))))
	}

	fmt.Println()
	fmt.Println(tui.Info("Scanning for RCM devices..."))
	if err := dm.Scan(); err != nil {
		return fmt.Errorf("device scan: %w", err)
	}

	selectedDev, err := dm.SelectInteractive()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  Device  : %s\n", selectedDev.Label)
	fmt.Printf("  Payload : %d bytes\n", len(data))
	fmt.Println()

	if !tui.Confirm("Proceed with injection?") {
		return fmt.Errorf("injection cancelled")
	}

	fmt.Println()
	fmt.Println(tui.Info("Reading device ID..."))
	deviceID, err := selectedDev.Dev.ReadDeviceID()
	if err != nil {
		return fmt.Errorf("reading device ID: %w", err)
	}
	fmt.Printf("  Device ID: %X\n", deviceID)

	fmt.Println(tui.Info("Building exploit payload..."))
	exploitPayload := exploit.BuildPayload(data)
	fmt.Printf("  Exploit payload: %d bytes\n", len(exploitPayload))

	fmt.Println(tui.Info("Triggering Fusee Gelee..."))
	if err := selectedDev.Dev.TriggerExploit(exploitPayload); err != nil {
		return fmt.Errorf("exploit failed: %w", err)
	}

	fmt.Println(tui.Success("Exploit triggered. Check your Switch screen."))
	return nil
}

// ── Download ─────────────────────────────────────────────────────────────────

func actionDownload(cfg *config.Config, pm *payload.Manager) error {
	if err := cfg.EnsureDownloadDir(); err != nil {
		return err
	}

	var labels []string
	for _, p := range payload.Registry {
		status := "[not downloaded]"
		if pm.Exists(p) {
			status = "[cached]"
		}
		labels = append(labels, fmt.Sprintf("%-22s %s  -- %s", p.Name, status, p.Description))
	}
	labels = append(labels, "Download ALL")

	idx := tui.SelectFromList("Select payload to download", labels)
	if idx < 0 {
		return nil
	}

	if idx == len(payload.Registry) {
		for _, p := range payload.Registry {
			if err := pm.Download(p); err != nil {
				fmt.Println(tui.Warn(fmt.Sprintf("Failed: %s: %v", p.Name, err)))
			}
			fmt.Println()
		}
		return nil
	}

	return pm.Download(payload.Registry[idx])
}

// ── Detect ───────────────────────────────────────────────────────────────────

func actionDetect(dm *device.Manager) error {
	fmt.Println(tui.Info("Scanning USB buses for Tegra X1 in RCM mode..."))
	if err := dm.Scan(); err != nil {
		fmt.Println(tui.Warn(err.Error()))
		return nil
	}
	dm.List()
	return nil
}

// ── Settings ─────────────────────────────────────────────────────────────────

func actionSettings(cfg *config.Config) error {
	fmt.Printf("  %-20s %s\n", "Config file:", cfg.ConfigPath())
	fmt.Printf("  %-20s %s\n", "Download dir:", cfg.DownloadDir)
	fmt.Println()

	settingsMenu := tui.NewMenu("Settings")

	settingsMenu.Add("Change download dir", "Set where payloads are saved", func() error {
		dir := tui.Prompt("New download directory")
		if dir == "" {
			return fmt.Errorf("no directory given")
		}
		cfg.DownloadDir = dir
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(tui.Success(fmt.Sprintf("Download dir set to: %s", dir)))
		return nil
	})

	settingsMenu.Add("Reset config", "Restore all settings to defaults", func() error {
		if !tui.Confirm("Reset all settings to defaults?") {
			return nil
		}
		*cfg = *config.Default()
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(tui.Success("Config reset to defaults."))
		return nil
	})

	return settingsMenu.Run()
}
