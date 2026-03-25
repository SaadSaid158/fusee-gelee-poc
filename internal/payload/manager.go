package payload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SaadSaid158/fusee-gelee-poc/internal/tui"
)

// KnownPayload describes a supported payload with its download source and hash.
type KnownPayload struct {
	Name        string
	Description string
	// DownloadURL is a direct .bin URL, OR a GitHub API releases/latest URL
	// (https://api.github.com/repos/OWNER/REPO/releases/latest) in which case
	// the asset whose name matches AssetPattern is downloaded automatically.
	DownloadURL  string
	AssetPattern string // substring match against asset name; only used with API URLs
	SHA256       string // expected hash; "" = print hash but don't fail
	Filename     string
}

// Registry is the list of all supported payloads.
var Registry = []KnownPayload{
	{
		Name:         "Hekate",
		Description:  "Nintendo Switch bootloader",
		DownloadURL:  "https://api.github.com/repos/CTCaer/hekate/releases/latest",
		AssetPattern: ".bin",
		SHA256:       "",
		Filename:     "hekate.bin",
	},
	{
		Name:         "Atmosphere (fusee)",
		Description:  "Atmosphère CFW fusee stage-1 loader",
		DownloadURL:  "https://github.com/Atmosphere-NX/Atmosphere/releases/latest/download/fusee.bin",
		AssetPattern: "",
		SHA256:       "",
		Filename:     "fusee.bin",
	},
	{
		Name:         "TegraExplorer",
		Description:  "Filesystem explorer and script runner from RCM",
		DownloadURL:  "https://api.github.com/repos/suchmememanyskill/TegraExplorer/releases/latest",
		AssetPattern: ".bin",
		SHA256:       "",
		Filename:     "tegraexplorer.bin",
	},
}

// Manager handles local payload storage, downloading, and verification.
type Manager struct {
	dir string
}

// NewManager creates a Manager that stores payloads in dir.
func NewManager(dir string) *Manager {
	return &Manager{dir: dir}
}

// LocalPath returns the full path for a payload filename.
func (m *Manager) LocalPath(filename string) string {
	return filepath.Join(m.dir, filename)
}

// Exists returns true if the payload file is already on disk.
func (m *Manager) Exists(p KnownPayload) bool {
	_, err := os.Stat(m.LocalPath(p.Filename))
	return err == nil
}

// Load reads a payload from disk and returns its bytes.
func (m *Manager) Load(p KnownPayload) ([]byte, error) {
	path := m.LocalPath(p.Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return data, nil
}

// LoadCustom reads an arbitrary .bin file from a user-supplied path.
func (m *Manager) LoadCustom(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading custom payload %s: %w", path, err)
	}
	return data, nil
}

// resolveURL returns the actual download URL for a payload.
// If DownloadURL points to a GitHub API releases/latest endpoint, it fetches
// the release metadata and picks the first asset whose name contains AssetPattern.
// Otherwise the DownloadURL is returned unchanged.
func resolveURL(p KnownPayload) (string, error) {
	if p.AssetPattern == "" {
		return p.DownloadURL, nil
	}

	resp, err := http.Get(p.DownloadURL)
	if err != nil {
		return "", fmt.Errorf("fetching release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing release JSON: %w", err)
	}

	for _, a := range release.Assets {
		if strings.Contains(a.Name, p.AssetPattern) &&
			!strings.Contains(a.Name, "ram8GB") &&
			!strings.Contains(a.Name, ".zip") {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no asset matching %q found in latest release", p.AssetPattern)
}

// Download fetches a payload, saves it to disk, then auto-verifies its SHA256.
func (m *Manager) Download(p KnownPayload) error {
	if err := os.MkdirAll(m.dir, 0755); err != nil {
		return fmt.Errorf("creating payload dir: %w", err)
	}

	fmt.Println(tui.Info(fmt.Sprintf("Downloading %s...", p.Name)))

	url, err := resolveURL(p)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	destPath := m.LocalPath(p.Filename)
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	bar := tui.NewProgressBar(p.Name, resp.ContentLength)

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return fmt.Errorf("writing to disk: %w", werr)
			}
			bar.Add(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
	}
	bar.Finish()
	fmt.Println(tui.Success(fmt.Sprintf("Saved to %s", destPath)))

	// Auto-verify
	fmt.Println(tui.Info("Verifying integrity..."))
	time.Sleep(1500 * time.Millisecond)
	if err := m.verifyFile(p, destPath); err != nil {
		return err
	}

	return nil
}

// Verify checks the SHA256 of a file on disk against the expected hash.
// If KnownPayload.SHA256 is empty, the file hash is printed but not checked.
func (m *Manager) Verify(p KnownPayload) error {
	return m.verifyFile(p, m.LocalPath(p.Filename))
}

// verifyFile is the internal implementation used by both Verify and Download.
func (m *Manager) verifyFile(p KnownPayload, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file for verification: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hashing file: %w", err)
	}
	got := hex.EncodeToString(h.Sum(nil))

	if p.SHA256 == "" {
		fmt.Printf("  SHA256: %s\n", got)
		fmt.Println(tui.Success("Integrity check done (no expected hash to compare)."))
		return nil
	}

	if got != p.SHA256 {
		return fmt.Errorf("SHA256 mismatch for %s:\n  expected: %s\n  got:      %s", p.Name, p.SHA256, got)
	}

	fmt.Println(tui.Success(fmt.Sprintf("SHA256 OK: %s...", got[:16])))
	return nil
}

// VerifyBytes checks the SHA256 of an in-memory payload against an expected hash.
func VerifyBytes(data []byte, expected string) error {
	if expected == "" {
		return nil
	}
	h := sha256.Sum256(data)
	got := hex.EncodeToString(h[:])
	if got != expected {
		return fmt.Errorf("SHA256 mismatch:\n  expected: %s\n  got:      %s", expected, got)
	}
	return nil
}

// HashBytes returns the SHA256 hex string of the given data.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
