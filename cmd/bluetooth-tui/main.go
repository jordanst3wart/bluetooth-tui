package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jordanst3wart/bluetooth-tui/internal/bluetooth"
	"github.com/jordanst3wart/bluetooth-tui/internal/ui"
)

func main() {
	m := ui.NewModel(bluetooth.NewBluetoothctlManager())
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start tui: %v\n", err)
		os.Exit(1)
	}
}
