package bluetooth

import "testing"

func TestMockManagerConnectSetsPairedAndConnected(t *testing.T) {
	m := NewMockManager([]Device{{Address: "A1", Name: "Test", Paired: false, Connected: false}})

	err := m.Connect("A1", 10)
	if err != nil {
		t.Fatalf("connect returned error: %v", err)
	}

	d, err := m.DeviceInfo("A1")
	if err != nil {
		t.Fatalf("device info returned error: %v", err)
	}

	if !d.Paired {
		t.Fatalf("expected device to be paired")
	}
	if !d.Connected {
		t.Fatalf("expected device to be connected")
	}
}

func TestMockManagerPairMissingDevice(t *testing.T) {
	m := NewMockManager(nil)

	err := m.Pair("missing", 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
