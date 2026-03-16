package ui

import (
	"fmt"
	"strings"
	"time"

	"bluetooth-tui2/internal/bluetooth"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultScanSeconds = 5
	defaultOpTimeout   = 10
)

type initResultMsg struct {
	err error
}

type scanResultMsg struct {
	devices []bluetooth.Device
	err     error
}

type scanTickMsg struct {
	count   time.Time
	devices []bluetooth.Device
	err     error
}

type connectResultMsg struct {
	address string
	err     error
}

type Model struct {
	manager       *bluetooth.BluetoothctlManager
	devices       []bluetooth.Device
	selected      int
	status        string
	loading       bool
	scanning      bool
	width         int
	height        int
	scanSeconds   int
	scanRemaining int
	timeoutSecond int
}

func NewModel(manager *bluetooth.BluetoothctlManager) Model {
	return Model{
		manager:       manager,
		status:        "Starting Bluetooth...", // probably not needed...
		loading:       true,
		scanSeconds:   defaultScanSeconds,
		timeoutSecond: defaultOpTimeout,
	}
}

func (m Model) Init() tea.Cmd {
	return m.initCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case initResultMsg:
		if msg.err != nil {
			m.loading = false
			m.status = "Failed to initialize: " + msg.err.Error()
			return m, nil
		}
		return m.startScan()

	case scanResultMsg:
		m.loading = false
		m.scanning = false
		m.scanRemaining = 0
		if msg.err != nil {
			m.status = "Scan failed: " + msg.err.Error()
			return m, nil
		}
		m.devices = msg.devices
		if m.selected >= len(m.devices) {
			m.selected = max(0, len(m.devices)-1)
		}
		if len(m.devices) == 0 {
			m.status = "No devices found."
		} else {
			m.status = fmt.Sprintf("%d devices", len(m.devices))
		}
		return m, nil

	case scanTickMsg:
		if msg.err != nil {
			m.status = "Scan errored: " + msg.err.Error()
			return m, nil
		}
		if !m.scanning {
			// shouldn't happened
			return m, nil
		}
		if m.scanRemaining > 0 {
			m.scanRemaining--
		}
		if m.scanRemaining == 0 {
			return m, nil
		}
		m.devices = msg.devices
		return m, tea.Batch(m.scanTickCmd())

	case connectResultMsg:
		m.loading = false
		if msg.err != nil {
			m.status = "Connection failed: " + msg.err.Error()
			return m, nil
		}
		for i := range m.devices {
			if m.devices[i].Address == msg.address {
				m.devices[i].Connected = true
				m.devices[i].Paired = true
				m.status = "Connected to " + m.devices[i].DisplayName()
				break
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.devices)-1 {
				m.selected++
			}
		case "r":
			return m.startScan()
		case "enter":
			if len(m.devices) == 0 || m.loading {
				return m, nil
			}
			target := m.devices[m.selected]
			if target.Connected {
				m.status = "Already connected to " + target.DisplayName()
				return m, nil
			}
			m.loading = true
			m.status = "Pairing and connecting to " + target.DisplayName() + "..."
			return m, m.connectCmd(target.Address)
		}
	}

	return m, nil
}

func (m Model) View() string {
	theme := newTheme()
	// TODO move heading to somewhere better
	header := theme.header.Render("Bluetooth Control")
	subtitle := fmt.Sprintf("%s select | %s connect | %s quit | %s rescan", theme.highlight.Render("↑/↓"), theme.highlight.Render("⏎"), theme.highlight.Render("q"), theme.highlight.Render("r"))

	var rows []string
	if len(m.devices) == 0 {
		rows = append(rows, theme.muted.Render("No devices yet"))
	} else {
		for i, d := range m.devices {
			state := "available"
			if d.Connected {
				state = "connected"
			} else if d.Paired {
				state = "paired"
			}

			row := fmt.Sprintf("%-25s", theme.deviceName.Render(d.DisplayName()))
			row += "  " + renderState(theme, state)
			if i == m.selected {
				rows = append(rows, theme.selectedRow.Render(row))
			} else {
				rows = append(rows, row)
			}
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	status := theme.status.Render(m.status)
	if m.scanning {
		// can remove known count
		status = theme.status.Render(fmt.Sprintf("%d devices, Scanning... %ds left", len(m.devices), m.scanRemaining))
	} else if m.loading {
		status = theme.status.Render("Working... " + m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, status, "", body, "", subtitle)
	return theme.base.Render(content)
}

func (m Model) initCmd() tea.Cmd {
	return func() tea.Msg {
		powered, err := m.manager.PowerState()
		if err != nil {
			return initResultMsg{err: err}
		}
		if !powered {
			if err := m.manager.SetPower(true); err != nil {
				return initResultMsg{err: err}
			}
		}

		return initResultMsg{err: nil}
	}
}

func (m Model) startScan() (Model, tea.Cmd) {
	m.loading = true
	m.scanning = true
	m.scanRemaining = m.scanSeconds
	m.status = "Scanning for devices..."
	return m, tea.Batch(m.scanCmd(), m.scanTickCmd())
}

func (m Model) scanCmd() tea.Cmd {
	return func() tea.Msg {
		devices, err := m.manager.Scan(m.scanSeconds)
		return scanResultMsg{devices: devices, err: err}
	}
}

func (m Model) scanTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		devices, err := m.manager.KnownDevices()
		return scanTickMsg{
			count:   t,
			devices: devices,
			err:     err,
		}
	})
}

func (m Model) connectCmd(address string) tea.Cmd {
	return func() tea.Msg {
		info, err := m.manager.DeviceInfo(address)
		if err != nil {
			return connectResultMsg{address: address, err: err}
		}
		if info.Connected {
			return connectResultMsg{address: address}
		}
		if !info.Paired {
			if err := m.manager.Pair(address, m.timeoutSecond); err != nil {
				return connectResultMsg{address: address, err: err}
			}
		}
		if err := m.manager.Connect(address, m.timeoutSecond); err != nil {
			return connectResultMsg{address: address, err: err}
		}
		return connectResultMsg{address: address}
	}
}

type theme struct {
	base        lipgloss.Style
	header      lipgloss.Style
	muted       lipgloss.Style
	highlight   lipgloss.Style
	selectedRow lipgloss.Style
	deviceName  lipgloss.Style
	deviceMeta  lipgloss.Style
	status      lipgloss.Style
	cursor      lipgloss.Style
	tagOK       lipgloss.Style
	tagWarn     lipgloss.Style
	tagIdle     lipgloss.Style
}

func newTheme() theme {
	return theme{
		base: lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			Foreground(lipgloss.Color("254")),
		header:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")),
		muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
		highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("230")),
		selectedRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("60")),
		deviceName: lipgloss.NewStyle().Bold(true),
		deviceMeta: lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		status:     lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Bold(true),
		cursor:     lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Bold(true),
		tagOK: lipgloss.NewStyle().
			Foreground(lipgloss.Color("120")),
		tagWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("222")),
		tagIdle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
	}
}

func renderState(t theme, state string) string {
	switch strings.ToLower(state) {
	case "connected":
		return t.tagOK.Render("connected")
	case "paired":
		return t.tagWarn.Render("paired")
	default:
		return t.tagIdle.Render("available")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
