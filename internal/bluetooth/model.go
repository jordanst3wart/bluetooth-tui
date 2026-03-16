package bluetooth

import (
	"errors"
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
	// KnownDevicesCount() (int, error) // remove...
	Pair(address string, timeoutSeconds int) error
	Connect(address string, timeoutSeconds int) error
	DeviceInfo(address string) (Device, error)
}
