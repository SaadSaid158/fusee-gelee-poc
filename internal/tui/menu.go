package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ClearScreen prints enough newlines to push old content off screen.
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// MenuItem represents a single menu entry.
type MenuItem struct {
	Label       string
	Description string
	Action      func() error
}

// Menu is a numbered interactive terminal menu.
type Menu struct {
	Title string
	Items []MenuItem
}

// NewMenu creates a menu with a title.
func NewMenu(title string) *Menu {
	return &Menu{Title: title}
}

// Add appends a menu item.
func (m *Menu) Add(label, description string, action func() error) {
	m.Items = append(m.Items, MenuItem{
		Label:       label,
		Description: description,
		Action:      action,
	})
}

// Run displays the menu in a loop until the user quits.
func (m *Menu) Run() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		ClearScreen()
		m.render()
		fmt.Print("  > ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)

		if line == "q" || line == "quit" || line == "exit" {
			ClearScreen()
			fmt.Println("  Goodbye.")
			fmt.Println()
			return nil
		}

		choice, err := strconv.Atoi(line)
		if err != nil || choice < 1 || choice > len(m.Items) {
			fmt.Printf("\n  Invalid choice -- enter 1-%d or q to quit.\n", len(m.Items))
			fmt.Print("  Press Enter to continue...")
			reader.ReadString('\n')
			continue
		}

		item := m.Items[choice-1]
		ClearScreen()
		fmt.Printf("  %s\n", item.Label)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println()

		if err := item.Action(); err != nil {
			fmt.Println()
			fmt.Println(Errf(err.Error()))
		}

		fmt.Println()
		fmt.Print("  Press Enter to continue...")
		reader.ReadString('\n')
	}
}

func (m *Menu) render() {
	PrintBanner()
	fmt.Printf("  %s\n\n", m.Title)
	for i, item := range m.Items {
		fmt.Printf("  [%d] %-24s %s\n", i+1, item.Label, item.Description)
	}
	fmt.Println()
	fmt.Println("  [q] Quit")
	fmt.Println()
}

// Confirm asks a yes/no question. Returns true on y/Y/yes.
func Confirm(prompt string) bool {
	fmt.Printf("  %s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

// SelectFromList shows a numbered list and returns the chosen index (0-based).
// Returns -1 if the user cancels.
func SelectFromList(title string, items []string) int {
	fmt.Printf("  %s\n", title)
	fmt.Println(strings.Repeat("-", 50))
	for i, item := range items {
		fmt.Printf("  [%d] %s\n", i+1, item)
	}
	fmt.Println("  [0] Cancel")
	fmt.Print("  > ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	choice, err := strconv.Atoi(line)
	if err != nil || choice < 1 || choice > len(items) {
		return -1
	}
	return choice - 1
}

// Prompt reads a single line of input with a label.
func Prompt(label string) string {
	fmt.Printf("  %s: ", label)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
