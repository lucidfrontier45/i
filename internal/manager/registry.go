package manager

import "github.com/lucidfrontier45/i/internal/types"

var drivers []types.Driver

func Register(d types.Driver) {
	drivers = append(drivers, d)
}

func All() []types.Driver {
	return drivers
}

func Lookup(name string) types.Driver {
	for _, d := range drivers {
		if d.Name() == name {
			return d
		}
	}
	return nil
}

func Detect() []types.Driver {
	var detected []types.Driver
	for _, d := range drivers {
		if d.Detect() {
			detected = append(detected, d)
		}
	}
	return detected
}
