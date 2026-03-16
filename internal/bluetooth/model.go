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
