package tui

import (
	"fmt"
	"strings"
)

const barWidth = 40

// ProgressBar renders a plain ASCII inline progress bar to stdout.
type ProgressBar struct {
	label string
	total int64
	done  int64
}

// NewProgressBar creates a new bar with a label and total byte count.
func NewProgressBar(label string, total int64) *ProgressBar {
	return &ProgressBar{label: label, total: total}
}

// Update sets the current progress and redraws the bar.
func (p *ProgressBar) Update(done int64) {
	p.done = done
	p.render()
}

// Add increments progress by n bytes and redraws.
func (p *ProgressBar) Add(n int64) {
	p.done += n
	p.render()
}

// Finish marks the bar as complete and prints a newline.
func (p *ProgressBar) Finish() {
	p.done = p.total
	p.render()
	fmt.Println()
}

func (p *ProgressBar) render() {
	pct := 0.0
	if p.total > 0 {
		pct = float64(p.done) / float64(p.total)
		if pct > 1.0 {
			pct = 1.0
		}
	}

	filled := int(pct * barWidth)
	empty := barWidth - filled

	bar := strings.Repeat("#", filled) + strings.Repeat("-", empty)

	fmt.Printf("\r  %-20s [%s] %5.1f%%  %s",
		p.label,
		bar,
		pct*100,
		formatBytes(p.done),
	)
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B ", b)
	}
}
