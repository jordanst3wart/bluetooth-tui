package bluetooth

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrNotFound = errors.New("device not found")

type Device struct {
	Address   string
	Name      string
	Connected bool
	Paired    bool
}

func (d Device) DisplayName() string {
	name := strings.TrimSpace(d.Name)
	if name == "" {
		return d.Address
	}
	return name
}

type Manager interface {
	PowerState() (bool, error)
	SetPower(on bool) error
	Scan(seconds int) ([]Device, error)
	KnownDevicesCount() (int, error)
	Pair(address string, timeoutSeconds int) error
	Connect(address string, timeoutSeconds int) error
	DeviceInfo(address string) (Device, error)
}

type MockManager struct {
	Powered bool
	Devices []Device
}

func NewMockManager(devices []Device) *MockManager {
	cp := append([]Device(nil), devices...)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].Address < cp[j].Address
	})
	return &MockManager{Powered: true, Devices: cp}
}

func (m *MockManager) PowerState() (bool, error) {
	return m.Powered, nil
}

func (m *MockManager) SetPower(on bool) error {
	m.Powered = on
	return nil
}

func (m *MockManager) Scan(seconds int) ([]Device, error) {
	if seconds < 0 {
		return nil, fmt.Errorf("invalid scan duration: %d", seconds)
	}
	return append([]Device(nil), m.Devices...), nil
}

func (m *MockManager) KnownDevicesCount() (int, error) {
	return len(m.Devices), nil
}

func (m *MockManager) Pair(address string, timeoutSeconds int) error {
	idx := m.indexOf(address)
	if idx == -1 {
		return ErrNotFound
	}
	m.Devices[idx].Paired = true
	return nil
}

func (m *MockManager) Connect(address string, timeoutSeconds int) error {
	idx := m.indexOf(address)
	if idx == -1 {
		return ErrNotFound
	}
	if !m.Devices[idx].Paired {
		m.Devices[idx].Paired = true
	}
	m.Devices[idx].Connected = true
	return nil
}

func (m *MockManager) DeviceInfo(address string) (Device, error) {
	idx := m.indexOf(address)
	if idx == -1 {
		return Device{}, ErrNotFound
	}
	return m.Devices[idx], nil
}

func (m *MockManager) indexOf(address string) int {
	for i, d := range m.Devices {
		if d.Address == address {
			return i
		}
	}
	return -1
}
