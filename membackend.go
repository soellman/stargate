package main

import (
	"bytes"
	"expvar"
	"net"
	"sync"
)

var (
	vars = map[string]*expvar.Map{
		"networks": expvar.NewMap("networks"),
		"devices":  expvar.NewMap("devices"),
	}
)

type MemBackend struct {
	networks []Network
	devices  map[string][]Device
	nlock    sync.Mutex
	dlock    sync.Mutex
}

func NewMemBackend() Backend {
	return &MemBackend{
		networks: []Network{},
		devices:  map[string][]Device{},
		nlock:    sync.Mutex{},
		dlock:    sync.Mutex{},
	}
}

func (s *MemBackend) Open() {
	debugf("opened memory store")
}

func (s *MemBackend) Close() {
	debugf("closed memory store")
}

func (s *MemBackend) HWAddrExists(hw net.HardwareAddr) bool {
	s.dlock.Lock()
	defer s.dlock.Unlock()
	for _, dd := range s.devices {
		for _, d := range dd {
			if bytes.Equal(d.HardwareAddr, hw) {
				return true
			}
		}
	}
	return false
}

func (s *MemBackend) Networks() []Network {
	s.nlock.Lock()
	defer s.nlock.Unlock()
	return s.networks
}

func (s *MemBackend) AddNetwork(network Network) {
	s.nlock.Lock()
	defer s.nlock.Unlock()
	s.networks = append(s.networks, network)

	// TODO: this makes invalid expvar json
	vars["networks"].Set(network.Name, &network.IPNet)
	vars["devices"].Set(network.Name, new(expvar.Map).Init())
	debugf("added network %s", network.Name)
}

func (s *MemBackend) RemoveNetwork(network Network) {
	s.nlock.Lock()
	defer s.nlock.Unlock()
	n := []Network{}
	for _, net := range s.networks {
		if net.Name == network.Name {
			n = append(n, network)
		}
	}
	s.networks = n

	debugf("removed network %s", network.Name)
}

func (s *MemBackend) AddDevice(networks []string, device Device) {
	s.dlock.Lock()
	defer s.dlock.Unlock()
	for _, network := range networks {
		s.devices[network] = append(s.devices[network], device)

		dvars := vars["devices"].Get(network).(*expvar.Map)
		dval := new(expvar.String)
		dval.Set(device.Name)
		dvars.Set(device.HardwareAddr.String(), dval)
	}

	debugf("added device %s to networks %v", device.HardwareAddr.String(), networks)
}

func (s *MemBackend) RemoveDevice(device Device) {
	s.dlock.Lock()
	defer s.dlock.Unlock()
	for _, network := range s.networks {
		devices := []Device{}
		for _, d := range s.devices[network.Name] {
			if d.Name != device.Name {
				devices = append(devices, device)
			}
		}
		s.devices[network.Name] = devices
	}

	debugf("removed device %s", device.HardwareAddr.String())
}
