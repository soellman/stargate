package main

import (
	"net"
	"time"
)

// Network represents a network that can be accessed through the portal
type Network struct {
	Name string
	net.IPNet
}

// Device represents a device that can access the portal
type Device struct {
	Name string
	net.HardwareAddr
}

// ListNetworks can enumnerate its networks
type ListNetworks interface {
	Networks() []Network
}

// Networks can ListNetworks and add/remove networks
type Networks interface {
	ListNetworks
	AddNetwork(network Network)
	RemoveNetwork(network Network)
}

// Devices can manage devices
type Devices interface {
	HWAddrExists(hw net.HardwareAddr) bool
	AddDevice(networks []string, device Device) // error?
	RemoveDevice(device Device)                 // error?
}

// Backend represents a firewall interface, e.g. iptables
type Backend interface {
	Networks
	Devices
	Open()
	Close()
}

// Token represents a token which can be used to gain access to networks by devices
type Token struct {
	Name         string   `json:"name"`
	Duration     string   `json:"duration"`
	Keys         []string `json:"keys"`
	NetworkNames []string `json:"networks"`

	duration time.Duration
}
