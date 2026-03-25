package tui

import "fmt"

// No ANSI colours — plain ASCII output for maximum compatibility.
// These stubs keep the call-sites unchanged while removing all escape codes.

const (
	Reset         = ""
	Bold          = ""
	Dim           = ""
	Red           = ""
	Green         = ""
	Yellow        = ""
	Blue          = ""
	Magenta       = ""
	Cyan          = ""
	White         = ""
	BrightRed     = ""
	BrightGreen   = ""
	BrightYellow  = ""
	BrightBlue    = ""
	BrightMagenta = ""
	BrightCyan    = ""
	BrightWhite   = ""
)

func Colorf(color, format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func Success(msg string) string { return "[OK]  " + msg }
func Warn(msg string) string    { return "[!!]  " + msg }
func Errf(msg string) string    { return "[ERR] " + msg }
func Info(msg string) string    { return "[..] " + msg }
func BoldS(msg string) string   { return msg }

func PrintBanner() {
	fmt.Println("  ------------------------------------------------")
	fmt.Println("  fusee-gelee  //  CVE-2018-6242  //  Tegra X1 RCM")
	fmt.Println("  For use on hardware you own only.")
	fmt.Println("  ------------------------------------------------")
	fmt.Println()
}
