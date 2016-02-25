package main

import (
	"log"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"syscall"

	"github.com/soellman/pidfile"
)

func main() {
	// check for linux
	if runtime.GOOS != "linux" {
		log.Fatalf("Sorry, only linux is supported at this time.")
	}

	// check for root
	if !isRoot() {
		log.Fatalf("Must run as root.")
	}

	// parse config
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatalf("Configuration file didn't parse: %v\n", err)
	}

	// do runtime validation
	if err := cfg.runtimeValidate(); err != nil {
		log.Fatalf("Runtime validation failed: %v\n", err)
	}

	// check for pid
	if err = pidfile.Write(pfile); err != nil {
		log.Fatalf("Error writing pid file: %#v", err)
	}

	// prepare for the end
	done := make(chan error, 1)
	trapSignals(done)

	// start the backend and sync nets from the config
	backend := NewIPTablesBackend(cfg.backendConfig())
	backend.Open()
	SyncNetworks(backend, cfg)

	// start up the server
	scfg := cfg.serverConfig()
	s := NewServer(scfg, backend)
	go func() {
		log.Printf("stargate opening at address %s\n", scfg.listenIP)
		done <- s.ListenAndServe()
	}()

	// we're done
	err = <-done
	status := 0
	if err != nil {
		status = 1
		log.Printf("stargate problem: %v\n", err)
	}

	// close up shop
	backend.Close()
	pidfile.Remove("/var/run/stargate.pid")
	log.Printf("stargate is closed\n")
	os.Exit(status)
}

func isRoot() bool {
	u, err := user.Current()
	if err != nil {
		log.Fatalf("Can't determine current user: %v\n", err)
	}
	return u.Uid == "0"
}

func trapSignals(done chan error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; done <- nil }()
}
