package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/mostlygeek/arp"
)

// Server represents the portal server
type Server struct {
	*http.Server
	*http.ServeMux
	ServerConfig
	templates *template.Template
	backend   Backend
}

// NewServer creates a server from a config and backend
func NewServer(c ServerConfig, b Backend) *Server {
	s := &Server{}
	s.ServerConfig = c
	s.backend = b
	s.templates = getTemplates()

	s.ServeMux = http.DefaultServeMux
	s.HandleFunc("/", s.Handler)
	s.Server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", c.listenIP, c.ports.HTTP),
		Handler: s,
	}
	return s
}

// IsLocal determines if the remote IP is part of the local network
func (s Server) IsLocal(remote string) bool {
	host := strings.Split(remote, ":")[0]
	ip := net.ParseIP(host)
	return s.localnet.Contains(ip)
}

// Redirect redirects the user to the server's redirect URL
func (s Server) Redirect(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, s.redirect, http.StatusFound)
}

// DisplayMessage renders the login page with a specified message
func (s Server) DisplayMessage(w http.ResponseWriter, message string) {
	s.templates.ExecuteTemplate(w, "index.html", struct{ Message string }{Message: message})
}

// Handler allows server to satisfy the http.Handler interface
func (s Server) Handler(w http.ResponseWriter, req *http.Request) {
	// Redirect any non-local requests
	if !s.IsLocal(req.RemoteAddr) {
		debugf("redirecting non-local request from %s", req.RemoteAddr)
		s.Redirect(w, req)
		return
	}

	switch req.Method {
	case "GET":
		// Redirect authorized devices
		hw, _ := HardwareAddr(req.RemoteAddr)
		if s.backend.HWAddrExists(hw) {
			debugf("redirecting GET from authorized device %s", hw)
			s.DisplayMessage(w, "you're already logged in")
			return
		}
		s.templates.ExecuteTemplate(w, "index.html", nil)
		return

	case "POST":
		// TODO: instead of unauth, redirect to index with error page
		hw, err := HardwareAddr(req.RemoteAddr)
		if err != nil {
			debugf("rejecting request with indeterminate mac: %v", err)
			s.DisplayMessage(w, "unauthorized")
			return
		}

		// Redirect authorized devices
		if s.backend.HWAddrExists(hw) {
			debugf("redirecting POST from authorized device %s", hw)
			s.Redirect(w, req)
			return
		}

		// Reject unauthorized devices
		token, err := s.Token(req.PostFormValue("key"))
		if err != nil {
			debugf("rejecting invalid key: %v\n", err)
			s.DisplayMessage(w, "unauthorized")
			return
		}

		// Authorize and redirect new device
		s.backend.AddDevice(token.NetworkNames, Device{HardwareAddr: hw})
		log.Printf("authorized %s as %s", hw, token.Name)

		http.Redirect(w, req, "", http.StatusFound)
		return

	default:
		// Disallow other methods
		debugf("rejecting disallowed %s method", req.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// Token returns a token which matches the provided key
func (s Server) Token(key string) (t Token, err error) {
	for _, t = range s.tokens {
		for _, k := range t.Keys {
			if key == k {
				return
			}
		}
	}
	err = errors.New("no token found")
	return
}

// HardwareAddr returns the mac addr for a local IP, or an error
func HardwareAddr(remote string) (hw net.HardwareAddr, err error) {
	addr := strings.Split(remote, ":")
	ip := net.ParseIP(addr[0])
	mac := arp.Search(ip.String())
	if mac == "" {
		return nil, fmt.Errorf("unable to resolve hardware address for %s", ip.String())
	}
	return net.ParseMAC(mac)
}
