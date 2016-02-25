package main

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/coreos/go-iptables/iptables"
)

type chain struct {
	name  string
	table string
	hook  string
	rules []string
}

// Chains returns a list of the main chains with their embedded configuration
func (b *IPTablesBackend) chains() []chain {
	// Add rules to govern what ports are available on the gate host
	rules := []string{}
	for _, port := range b.config.ports.TCP {
		rules = append(rules, fmt.Sprintf("-p %s --dport %d -j RETURN", "tcp", port))
	}
	for _, port := range b.config.ports.UDP {
		rules = append(rules, fmt.Sprintf("-p %s --dport %d -j RETURN", "udp", port))
	}
	return []chain{
		{"captive_check", "mangle", "PREROUTING",
			[]string{
				"-j captive_allowed",
				"-j MARK --set-mark 99",
			}},
		{"captive_redirect", "nat", "PREROUTING",
			[]string{
				fmt.Sprintf("-m mark --mark 99 -p tcp --dport 80 -j DNAT --to-destination %s:%d", b.config.ip, b.config.ports.HTTP),
				fmt.Sprintf("-m mark --mark 99 -p tcp --dport 443 -j DNAT --to-destination %s:%d", b.config.ip, b.config.ports.HTTPS),
			}},
		{"captive_return", "nat", "POSTROUTING",
			[]string{
				fmt.Sprintf("-d $NET -p tcp --sport %d -j SNAT --to-source :80", b.config.ports.HTTP),
				fmt.Sprintf("-d $NET -p tcp --sport %d -j SNAT --to-source :443", b.config.ports.HTTPS),
			}},
		{"captive_input", "filter", "INPUT",
			append(rules, []string{
				fmt.Sprintf("-p tcp -m multiport --dports %d,%d -j RETURN", b.config.ports.HTTP, b.config.ports.HTTPS),
				"-j REJECT",
			}...)},
		{"captive_forward", "filter", "FORWARD",
			[]string{
				"-p udp --dport 53 -j RETURN",
				"-m mark --mark 99 -j REJECT",
			}},
	}
}

// IPTablesBackend represents a portal backend supporting iptables
type IPTablesBackend struct {
	ipt      *iptables.IPTables
	config   BackendConfig
	networks []Network
	devices  []Device
	nlock    sync.Mutex
	dlock    sync.Mutex
}

// NewIPTablesBackend returns a backend provided a config
func NewIPTablesBackend(cfg BackendConfig) Backend {
	i, err := iptables.New()
	if err != nil {
		panic("iptables not supported")
	}
	return &IPTablesBackend{
		ipt:      i,
		config:   cfg,
		networks: []Network{},
		devices:  []Device{},
		nlock:    sync.Mutex{},
		dlock:    sync.Mutex{},
	}
}

// Open will initialize iptables by defining chains
// and inserting them into the built-in chains
func (b *IPTablesBackend) Open() {
	b.ipt.NewChain("mangle", "captive_allowed")
	for _, c := range b.chains() {
		b.ipt.NewChain(c.table, c.name)
		for _, t := range c.rules {
			rule := strings.Split(t, " ")
			b.ipt.AppendUnique(c.table, c.name, rule...)
		}
		b.ipt.AppendUnique(c.table, c.hook, "-s", b.config.net, "-j", c.name)
	}

	b.ipt.AppendUnique("nat", "POSTROUTING", "-j", "MASQUERADE")

	// Drop the main gate
	b.ipt.Delete("filter", "INPUT", "-s", b.config.net, "-j", "REJECT")
	b.ipt.Delete("filter", "FORWARD", "-s", b.config.net, "-j", "REJECT")

	debugf("opened iptables backend")
}

// Close will remove the portal chains from the built-in chains
// and remove the chains themselves
// Close will also insert basic rules to firewall the managed network
func (b *IPTablesBackend) Close() {
	b.ipt.Delete("nat", "POSTROUTING", "-j", "MASQUERADE")

	for _, n := range b.networks {
		b.RemoveNetwork(n)
	}

	for _, c := range b.chains() {
		b.ipt.Delete(c.table, c.hook, "-s", b.config.net, "-j", c.name)
		b.ipt.ClearChain(c.table, c.name)
		b.ipt.DeleteChain(c.table, c.name)
	}

	b.ipt.ClearChain("mangle", "captive_allowed")
	b.ipt.DeleteChain("mangle", "captive_allowed")

	// Add rules to keep the hordes at bay
	b.ipt.AppendUnique("filter", "INPUT", "-s", b.config.net, "-j", "REJECT")
	b.ipt.AppendUnique("filter", "FORWARD", "-s", b.config.net, "-j", "REJECT")

	debugf("closed iptables backend")
}

// HWAddrExists checks if the specified mac addr is known to the portal
func (b *IPTablesBackend) HWAddrExists(hw net.HardwareAddr) bool {
	b.dlock.Lock()
	defer b.dlock.Unlock()
	for _, d := range b.devices {
		if bytes.Equal(d.HardwareAddr, hw) {
			return true
		}
	}
	return false
}

// Networks fulfills the ListNetworks interface
func (b *IPTablesBackend) Networks() []Network {
	b.nlock.Lock()
	defer b.nlock.Unlock()
	return b.networks
}

// AddNetwork fulfills the Networks interface
func (b *IPTablesBackend) AddNetwork(network Network) {
	b.nlock.Lock()
	defer b.nlock.Unlock()
	b.networks = append(b.networks, network)

	b.ipt.NewChain("filter", "access_"+network.Name)
	b.ipt.AppendUnique("filter", "FORWARD", "-s", b.config.net, "-d", network.String(), "-j", "access_"+network.Name)
	b.ipt.AppendUnique("filter", "FORWARD", "-s", b.config.net, "-d", network.String(), "-j", "DROP")

	debugf("network %s added", network.Name)
}

// RemoveNetwork fulfills the Networks interface
func (b *IPTablesBackend) RemoveNetwork(network Network) {
	b.nlock.Lock()
	defer b.nlock.Unlock()
	networks := []Network{}
	for _, n := range b.networks {
		if n.Name == network.Name {
			networks = append(networks, network)
		}
	}
	b.networks = networks

	b.ipt.Delete("filter", "FORWARD", "-s", b.config.net, "-d", network.String(), "-j", "DROP")
	b.ipt.Delete("filter", "FORWARD", "-s", b.config.net, "-d", network.String(), "-j", "access_"+network.Name)
	b.ipt.ClearChain("filter", "access_"+network.Name)
	b.ipt.DeleteChain("filter", "access_"+network.Name)

	debugf("network %s removed", network.Name)
}

// AddDevice fulfills the Device interface
func (b *IPTablesBackend) AddDevice(networks []string, device Device) {
	b.dlock.Lock()
	defer b.dlock.Unlock()
	b.devices = append(b.devices, device)

	b.ipt.AppendUnique("mangle", "captive_allowed", "-m", "mac", "--mac-source", device.HardwareAddr.String(), "-j", "ACCEPT")
	for _, n := range networks {
		b.ipt.AppendUnique("filter", "access_"+n, "-m", "mac", "--mac-source", device.HardwareAddr.String(), "-j", "ACCEPT")
	}

	debugf("added device %s to networks %v", device.HardwareAddr.String(), networks)
}

// RemoveDevice fulfills the Device interface
func (b *IPTablesBackend) RemoveDevice(device Device) {
	b.dlock.Lock()
	defer b.dlock.Unlock()
	devices := []Device{}
	for _, d := range b.devices {
		if d.Name != device.Name {
			devices = append(devices, device)
		}
	}
	b.devices = devices

	b.ipt.Delete("mangle", "captive_allowed", "-m", "mac", "--mac-source", device.HardwareAddr.String(), "-j", "ACCEPT")
	for _, n := range b.networks {
		b.ipt.Delete("filter", "access_"+n.Name, "-m", "mac", "--mac-source", device.HardwareAddr.String(), "-j", "ACCEPT")
	}

	debugf("removed device %s", device.HardwareAddr.String())
}
