package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/ghodss/yaml"
)

var (
	debug bool
	cfile string
	pfile string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "debug logging")
	flag.StringVar(&cfile, "config", "/etc/stargate.yaml", "config file path")
	flag.StringVar(&pfile, "pidfile", "/var/run/stargate.pid", "pid file path")
	flag.Parse()
}

var (
	defaultHTTP     = 7676
	defaultHTTPS    = 7677
	defaultRedirect = "https://google.com"
	defaultTCP      = []int{}
	defaultUDP      = []int{67}
)

// Config represents the configuration object
type Config struct {
	ListenIP string `json:"listen"`
	Ports    struct {
		HTTP  int   `json:"http"`
		HTTPS int   `json:"https"`
		TCP   []int `json:"tcp"`
		UDP   []int `json:"udp"`
	} `json:"ports"`
	Redirect string `json:"redirect"`
	Nets     []struct {
		Name string `json:"name"`
		CIDR string `json:"network"`
	} `json:"networks"`
	Tokens []Token `json:"tokens"`

	networks []Network
	ipnet    *net.IPNet
}

// BackendConfig configures the portal backends
type BackendConfig struct {
	ports struct {
		HTTP  int
		HTTPS int
		TCP   []int
		UDP   []int
	}
	net string
	ip  string
}

// ServerConfig configures the portal server
type ServerConfig struct {
	ports struct {
		HTTP  string
		HTTPS string
	}
	listenIP string
	redirect string
	localnet *net.IPNet
	tokens   []Token
}

// ParseConfig parses file configuration and returns a Config
func ParseConfig() (c *Config, err error) {
	return ParseConfigFile(cfile)
}

// ParseConfigFile parses file configuration from filename and returns a Config
func ParseConfigFile(filename string) (c *Config, err error) {

	var data []byte
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	c = &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return
	}

	c.applyDefaults()

	err = c.validate()

	return
}

// Networks Satisfy the ListNetworks interface
func (c *Config) Networks() []Network {
	return c.networks
}

// Fill in default values
func (c *Config) applyDefaults() {
	if c.Ports.HTTP == 0 {
		c.Ports.HTTP = defaultHTTP
	}
	if c.Ports.HTTPS == 0 {
		c.Ports.HTTPS = defaultHTTPS
	}
	if len(c.Ports.TCP) == 0 {
		c.Ports.TCP = defaultTCP
	}
	if len(c.Ports.UDP) == 0 {
		c.Ports.UDP = defaultUDP
	}
	if c.Redirect != "" {
		c.Redirect = defaultRedirect
	}
}

// Validate the raw input from the config file
func (c *Config) validate() error {
	if err := c.parseNetworks(); err != nil {
		return err
	}

	if err := c.parseTokens(); err != nil {
		return err
	}

	if _, err := url.Parse(c.Redirect); err != nil {
		return err
	}

	return nil
}

// Runtime validation validates the config according to the runtime
func (c *Config) runtimeValidate() error {
	var err error
	c.ipnet, err = c.determineIPNet()
	return err
}

// Verify that the provided listen addr is bound to an interface
// and return the *net.IPNet struct
func (c *Config) determineIPNet() (*net.IPNet, error) {
	ip := net.ParseIP(c.ListenIP)
	if ip == nil {
		return nil, errors.New("listen address can't be parsed as ip:host")
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if addr, ok := addr.(*net.IPNet); ok {
			if addr.IP.String() == ip.String() {
				return addr, nil
			}
		}
	}

	return nil, errors.New("listen address is not bound to a local interface")
}

// Parse the networks supplied in the file input
func (c *Config) parseNetworks() error {
	nets := []Network{}
	for _, network := range c.Nets {
		ip, ipnet, err := net.ParseCIDR(network.CIDR)
		if err != nil {
			return err
		}
		if ip.String() != ipnet.IP.String() {
			msg := fmt.Sprintf("network address %s for name %s specifies an IP, not a network.", network.CIDR, network.Name)
			return errors.New(msg)
		}
		nets = append(nets, Network{
			Name: network.Name,
			IPNet: net.IPNet{
				IP:   ip,
				Mask: ipnet.Mask,
			},
		})
	}

	c.networks = nets
	return nil
}

// Parse the tokens supplied in the file input
func (c *Config) parseTokens() error {
	for i, t := range c.Tokens {
		if t.Duration != "" {
			d, err := time.ParseDuration(t.Duration)
			if err != nil {
				return err
			}
			c.Tokens[i].duration = d
		}
	}
	return nil
}

// Construct a backend config
func (c *Config) backendConfig() (b BackendConfig) {
	b.ports.HTTP = c.Ports.HTTP
	b.ports.HTTPS = c.Ports.HTTPS
	b.ports.TCP = c.Ports.TCP
	b.ports.UDP = c.Ports.UDP
	b.ip = c.ListenIP
	b.net = c.ipnet.String()
	return
}

// Construct a server config
func (c *Config) serverConfig() (s ServerConfig) {
	s.tokens = c.Tokens
	s.redirect = c.Redirect
	s.listenIP = c.ListenIP
	s.ports.HTTP = strconv.Itoa(c.Ports.HTTP)
	s.ports.HTTPS = strconv.Itoa(c.Ports.HTTPS)
	s.localnet = c.ipnet
	s.redirect = c.Redirect
	return
}
