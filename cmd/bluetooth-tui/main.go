package main

import (
	"fmt"
	"os"

	"bluetooth-tui2/internal/bluetooth"
	"bluetooth-tui2/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.NewModel(bluetooth.NewBluetoothctlManager())
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start tui: %v\n", err)
		os.Exit(1)
	}
}
