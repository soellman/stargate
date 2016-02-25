package main

// This program provides authentication/authorization for open wireless networks.
// On startup, it will install an iptables rule to receive all port 80 traffic
// from devices not already seen, and on receiving such traffic, display
// a page including an AUP and the option to include an access code to grant
// the device access to other local networks.
