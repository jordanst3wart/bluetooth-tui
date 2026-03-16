package bluetooth

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type commandRunner interface {
	Run(timeout time.Duration, name string, args ...string) (string, error)
}

type systemRunner struct{}

func (systemRunner) Run(timeout time.Duration, name string, args ...string) (string, error) {
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

type BluetoothctlManager struct {
	runner commandRunner
}

func NewBluetoothctlManager() *BluetoothctlManager {
	return &BluetoothctlManager{runner: systemRunner{}}
}

func newBluetoothctlManagerWithRunner(runner commandRunner) *BluetoothctlManager {
	return &BluetoothctlManager{runner: runner}
}

func (m *BluetoothctlManager) PowerState() (bool, error) {
	out, err := m.runCtl(0, "show")
	if err != nil {
		return false, err
	}
	return parsePowerState(out)
}

func (m *BluetoothctlManager) SetPower(on bool) error {
	arg := "off"
	if on {
		arg = "on"
	}
	_, err := m.runCtl(0, "power", arg)
	return err
}

func (m *BluetoothctlManager) Scan(seconds int) ([]Device, error) {
	if seconds <= 0 {
		seconds = 1
	}
	_, err := m.runCtl(time.Duration(seconds+1)*time.Second, "--timeout", fmt.Sprintf("%d", seconds), "scan", "on")
	if err != nil {
		return nil, err
	}
	out, err := m.runCtl(0, "devices")
	if err != nil {
		return nil, err
	}
	devices := parseDevices(out)
	for i := range devices {
		info, infoErr := m.DeviceInfo(devices[i].Address)
		if infoErr != nil {
			continue
		}
		if strings.TrimSpace(info.Name) != "" {
			devices[i].Name = info.Name
		}
		devices[i].Paired = info.Paired
		devices[i].Connected = info.Connected
	}
	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].DisplayName()) > strings.ToLower(devices[j].DisplayName())
	})
	return devices, nil
}

func (m *BluetoothctlManager) KnownDevices() ([]Device, error) {
	out, err := m.runCtl(0, "devices")
	if err != nil {
		return []Device{}, err
	}
	return parseDevices(out), nil
}

func (m *BluetoothctlManager) Pair(address string, timeoutSeconds int) error {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	_, err := m.runCtl(time.Duration(timeoutSeconds+1)*time.Second, "--timeout", fmt.Sprintf("%d", timeoutSeconds), "pair", address)
	return err
}

func (m *BluetoothctlManager) Connect(address string, timeoutSeconds int) error {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	_, err := m.runCtl(time.Duration(timeoutSeconds+1)*time.Second, "--timeout", fmt.Sprintf("%d", timeoutSeconds), "connect", address)
	return err
}

func (m *BluetoothctlManager) DeviceInfo(address string) (Device, error) {
	out, err := m.runCtl(0, "info", address)
	if err != nil {
		return Device{}, err
	}
	info, ok := parseDeviceInfo(out, address)
	if !ok {
		return Device{}, ErrNotFound
	}
	return info, nil
}

func (m *BluetoothctlManager) runCtl(timeout time.Duration, args ...string) (string, error) {
	if _, ok := m.runner.(systemRunner); ok {
		if _, err := exec.LookPath("bluetoothctl"); err != nil {
			return "", fmt.Errorf("bluetoothctl not found: %w", err)
		}
	}
	return m.runner.Run(timeout, "bluetoothctl", args...)
}

func parsePowerState(output string) (bool, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Powered:") {
			return parseOnOff(strings.TrimSpace(strings.TrimPrefix(line, "Powered:")))
		}
		if strings.HasPrefix(line, "PowerState:") {
			return parseOnOff(strings.TrimSpace(strings.TrimPrefix(line, "PowerState:")))
		}
	}
	return false, fmt.Errorf("power state not found")
}

func parseOnOff(v string) (bool, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "yes", "on", "true":
		return true, nil
	case "no", "off", "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid power value %q", v)
	}
}

func parseDevices(output string) []Device {
	var devices []Device
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Device ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		addr := strings.TrimSpace(fields[1])
		if addr == "" {
			continue
		}
		name := ""
		if len(fields) > 2 {
			name = strings.Join(fields[2:], " ")
		}
		devices = append(devices, Device{Address: addr, Name: name})
	}
	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].DisplayName()) > strings.ToLower(devices[j].DisplayName())
	})
	return devices
}

func parseDeviceInfo(output, address string) (Device, bool) {
	device := Device{Address: address}
	found := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Device ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				device.Address = fields[1]
				found = true
			}
			continue
		}
		// TODO fix this
		if strings.HasPrefix(line, "Name:") {
			device.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
			continue
		}
		if strings.HasPrefix(line, "Paired:") {
			paired, err := parseOnOff(strings.TrimSpace(strings.TrimPrefix(line, "Paired:")))
			if err == nil {
				device.Paired = paired
			}
			continue
		}
		if strings.HasPrefix(line, "Connected:") {
			connected, err := parseOnOff(strings.TrimSpace(strings.TrimPrefix(line, "Connected:")))
			if err == nil {
				device.Connected = connected
			}
		}
	}
	if !found {
		return Device{}, false
	}
	return device, true
}
