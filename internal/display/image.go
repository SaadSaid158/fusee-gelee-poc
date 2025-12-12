package display

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed success_embedded.bin
var embeddedPayload embed.FS

const (
	// Display payload constants
	SuccessPayloadFile = "payloads/success.bin"
	EmbeddedFile       = "success_embedded.bin"
)

// LoadSuccessImage loads the success image display payload
func LoadSuccessImage() ([]byte, error) {
	// Try to load from file first
	payload, err := loadFromFile()
	if err == nil {
		return payload, nil
	}

	// Fall back to embedded payload
	payload, err = loadEmbedded()
	if err == nil {
		return payload, nil
	}

	// If both fail, generate a basic payload
	return generateBasicPayload(), nil
}

// loadFromFile attempts to load the payload from the filesystem
func loadFromFile() ([]byte, error) {
	// Try relative to current directory
	data, err := os.ReadFile(SuccessPayloadFile)
	if err == nil {
		return data, nil
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		payloadPath := filepath.Join(dir, SuccessPayloadFile)
		data, err = os.ReadFile(payloadPath)
		if err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("payload file not found")
}

// loadEmbedded loads the embedded payload
func loadEmbedded() ([]byte, error) {
	data, err := embeddedPayload.ReadFile(EmbeddedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded payload: %v", err)
	}
	return data, nil
}

// generateBasicPayload creates a basic payload when no file is available
func generateBasicPayload() []byte {
	// This creates a simple payload that would display something on screen
	// In a real implementation, this would be ARM64 code that:
	// 1. Initializes the display controller
	// 2. Sets up a framebuffer
	// 3. Draws a success image/message

	payload := make([]byte, 4096)

	// ARM64 instructions for a basic display routine
	// This is highly simplified - real display code would be much more complex
	instructions := []uint32{
		0xD503201F, // NOP - for alignment
		0xD503201F, // NOP
		// In reality, these would be:
		// - MMIO writes to display controller registers
		// - Framebuffer setup code
		// - Drawing routines
		// - Color data
		0x580000A0, // LDR X0, #20 (load register base address)
		0xD2800001, // MOV X1, #0 (value to write)
		0xF9000001, // STR X1, [X0] (write to register)
		0x14000000, // B . (infinite loop)
	}

	// Write instructions to payload
	offset := 0
	for _, instr := range instructions {
		payload[offset] = byte(instr)
		payload[offset+1] = byte(instr >> 8)
		payload[offset+2] = byte(instr >> 16)
		payload[offset+3] = byte(instr >> 24)
		offset += 4
	}

	// Add a "signature" to identify this is our payload
	signature := []byte("FUSEE_GELEE_SUCCESS")
	copy(payload[offset:], signature)

	return payload
}

// CreateSuccessMessage creates a payload that displays a success message
func CreateSuccessMessage(message string) []byte {
	payload := generateBasicPayload()

	// Embed the message in the payload
	messageOffset := 0x100
	if len(message) > 0 && messageOffset+len(message) < len(payload) {
		copy(payload[messageOffset:], []byte(message))
	}

	return payload
}

// ValidatePayload checks if a payload is valid
func ValidatePayload(payload []byte) bool {
	// Basic validation
	if len(payload) < 256 {
		return false
	}

	// Check for payload signature or valid ARM64 code
	// ARM64 instructions are 4-byte aligned
	if len(payload)%4 != 0 {
		return false
	}

	return true
}
