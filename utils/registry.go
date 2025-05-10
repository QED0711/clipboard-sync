package utils

import (
	"sync"
)

var once sync.Once
var regCol *RegistryCollection

type ClipboardRegistry struct {
	mu   sync.RWMutex
	Name string
	Data string
}

type RegistryCollection struct {
	mu         sync.RWMutex
	registries map[string]*ClipboardRegistry
}

func GetRegistryCollection() *RegistryCollection {
	once.Do(func() {
		regCol = &RegistryCollection{
			mu:         sync.RWMutex{},
			registries: make(map[string]*ClipboardRegistry),
		}
	})
	return regCol
}

func (rc *RegistryCollection) SetRegistryData(name, data string) (*ClipboardRegistry, bool) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	existingReg, ok := rc.registries[name]
	if !ok {
		newReg := &ClipboardRegistry{
			Name: name,
			Data: data,
			mu:   sync.RWMutex{},
		}
		rc.registries[name] = newReg
		return newReg, true
	} else {
		existingReg.mu.Lock()
		defer existingReg.mu.Unlock()
		existingReg.Data = data
		return existingReg, true
	}
}

func (rc *RegistryCollection) GetRegistryByName(name string) (*ClipboardRegistry, bool) {
	reg, ok := rc.registries[name]
	return reg, ok
}
