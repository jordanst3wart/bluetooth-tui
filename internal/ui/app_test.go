package ui

import (
	"strings"
	"testing"

	"bluetooth-tui2/internal/bluetooth"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitLoadsDevices(t *testing.T) {
	mgr := bluetooth.NewMockManager([]bluetooth.Device{
		{Address: "AA", Name: "Headset", Paired: false, Connected: false},
		{Address: "BB", Name: "Speaker", Paired: true, Connected: false},
	})

	m := NewModel(mgr)
	msg := m.Init()()
	updated, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatalf("expected scan start command")
	}
	next := updated.(Model)
	if !next.loading || !next.scanning {
		t.Fatalf("expected scanning to be in progress")
	}

	scanMsg := next.scanCmd()()
	updated, _ = next.Update(scanMsg)
	next = updated.(Model)

	if len(next.devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(next.devices))
	}
	if next.loading {
		t.Fatalf("expected loading false")
	}
	if !strings.Contains(next.status, "Found") {
		t.Fatalf("unexpected status: %s", next.status)
	}
}

func TestEnterConnectsSelectedDevice(t *testing.T) {
	mgr := bluetooth.NewMockManager([]bluetooth.Device{
		{Address: "AA", Name: "Headset", Paired: false, Connected: false},
	})

	m := NewModel(mgr)
	initMsg := m.Init()()
	updated, _ := m.Update(initMsg)
	s := updated.(Model)
	scanMsg := s.scanCmd()()
	updated, _ = s.Update(scanMsg)
	s = updated.(Model)

	updated, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected connect command")
	}

	connectMsg := cmd()
	updated, _ = updated.(Model).Update(connectMsg)
	next := updated.(Model)

	if len(next.devices) != 1 || !next.devices[0].Connected {
		t.Fatalf("expected selected device to be connected")
	}
	if !strings.Contains(next.status, "Connected") {
		t.Fatalf("unexpected status: %s", next.status)
	}
}

func TestScanTickUpdatesCountdownAndView(t *testing.T) {
	mgr := bluetooth.NewMockManager([]bluetooth.Device{{Address: "AA", Name: "Headset"}})
	m := NewModel(mgr)

	updated, _ := m.Update(initResultMsg{})
	s := updated.(Model)
	if !s.scanning {
		t.Fatalf("expected scanning true")
	}

	updated, _ = s.Update(scanTickMsg{})
	next := updated.(Model)
	if next.scanRemaining != s.scanRemaining-1 {
		t.Fatalf("expected scanRemaining to decrement")
	}
	view := next.View()
	if !strings.Contains(view, "Scanning...") {
		t.Fatalf("expected scanning status in view")
	}
}

func TestViewContainsHeaderAndDeviceName(t *testing.T) {
	mgr := bluetooth.NewMockManager([]bluetooth.Device{{Address: "AA", Name: "Headset"}})
	m := NewModel(mgr)
	msg := m.Init()()
	updated, _ := m.Update(msg)
	s := updated.(Model)
	scanMsg := s.scanCmd()()
	updated, _ = s.Update(scanMsg)
	view := updated.(Model).View()

	if !strings.Contains(view, "Bluetooth Control") {
		t.Fatalf("expected header in view")
	}
	if !strings.Contains(view, "Headset") {
		t.Fatalf("expected device name in view")
	}
}
