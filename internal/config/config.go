package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".fusee-gelee.json"

// Config holds all persistent settings for the tool.
type Config struct {
	// LastDevice is the USB device string last used, for convenience.
	LastDevice string `json:"last_device,omitempty"`

	// Favourites is an ordered list of payload names the user has starred.
	Favourites []string `json:"favourites,omitempty"`

	// DownloadDir is where payloads are saved. Defaults to ./payloads.
	DownloadDir string `json:"download_dir,omitempty"`

	// AutoVerify controls whether SHA256 verification runs automatically.
	AutoVerify bool `json:"auto_verify"`

	// path is the resolved config file path — not serialised.
	path string
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		DownloadDir: "payloads",
		AutoVerify:  true,
	}
}

// Load reads the config from the user's home directory.
// If the file doesn't exist, a default config is returned (not an error).
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		cfg.path = path
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.path = path
	return cfg, nil
}

// Save writes the config back to disk.
func (c *Config) Save() error {
	if c.path == "" {
		path, err := configPath()
		if err != nil {
			return err
		}
		c.path = path
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// AddFavourite adds a payload name to favourites (no duplicates).
func (c *Config) AddFavourite(name string) {
	for _, f := range c.Favourites {
		if f == name {
			return
		}
	}
	c.Favourites = append(c.Favourites, name)
}

// RemoveFavourite removes a payload name from favourites.
func (c *Config) RemoveFavourite(name string) {
	out := c.Favourites[:0]
	for _, f := range c.Favourites {
		if f != name {
			out = append(out, f)
		}
	}
	c.Favourites = out
}

// IsFavourite returns true if name is in the favourites list.
func (c *Config) IsFavourite(name string) bool {
	for _, f := range c.Favourites {
		if f == name {
			return true
		}
	}
	return false
}

// ConfigPath returns the path to the config file for display purposes.
func (c *Config) ConfigPath() string {
	return c.path
}

// EnsureDownloadDir creates the download directory if it doesn't exist.
func (c *Config) EnsureDownloadDir() error {
	return os.MkdirAll(c.DownloadDir, 0755)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, configFileName), nil
}
