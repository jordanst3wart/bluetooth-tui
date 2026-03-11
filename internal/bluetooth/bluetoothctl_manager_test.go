package bluetooth

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	responses map[string]string
	errors    map[string]error
	calls     []string
}

func (f *fakeRunner) Run(timeout time.Duration, name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, key)
	if err, ok := f.errors[key]; ok {
		return "", err
	}
	if out, ok := f.responses[key]; ok {
		return out, nil
	}
	return "", fmt.Errorf("missing response for %s", key)
}

func TestParseDevices(t *testing.T) {
	out := "Device 11:22:33:44:55:66 Headset\nDevice AA:BB:CC:DD:EE:FF Office Speaker\n"
	devices := parseDevices(out)
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].Address != "11:22:33:44:55:66" || devices[0].Name != "Headset" {
		t.Fatalf("unexpected first device: %+v", devices[0])
	}
}

func TestParseDeviceInfo(t *testing.T) {
	out := "Device 11:22:33:44:55:66\nName: Headset\nPaired: yes\nConnected: no\n"
	d, ok := parseDeviceInfo(out, "11:22:33:44:55:66")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if d.Name != "Headset" || !d.Paired || d.Connected {
		t.Fatalf("unexpected parsed device: %+v", d)
	}
}

func TestBluetoothctlManagerScanAndConnect(t *testing.T) {
	runner := &fakeRunner{
		responses: map[string]string{
			"bluetoothctl show":                                   "Powered: yes\n",
			"bluetoothctl --timeout 3 scan on":                    "Discovery started\n",
			"bluetoothctl devices":                                "Device 11:22:33:44:55:66 Headset\n",
			"bluetoothctl info 11:22:33:44:55:66":                 "Device 11:22:33:44:55:66\nName: Headset\nPaired: no\nConnected: no\n",
			"bluetoothctl --timeout 10 pair 11:22:33:44:55:66":    "Pairing successful\n",
			"bluetoothctl --timeout 10 connect 11:22:33:44:55:66": "Connection successful\n",
		},
		errors: map[string]error{},
	}

	m := newBluetoothctlManagerWithRunner(runner)

	devices, err := m.Scan(3)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected one device, got %d", len(devices))
	}
	if devices[0].Name != "Headset" {
		t.Fatalf("unexpected device name: %s", devices[0].Name)
	}

	if err := m.Pair("11:22:33:44:55:66", 10); err != nil {
		t.Fatalf("pair failed: %v", err)
	}
	if err := m.Connect("11:22:33:44:55:66", 10); err != nil {
		t.Fatalf("connect failed: %v", err)
	}
}
