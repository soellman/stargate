# stargate
Easy flexible captive portal. Leave your network open, and stargate can control access for you. Run it on your router, if you dare!

Feedback is wanted. Open issues!

## Features

- Control access to multiple networks as well as the internet
- Tokens can 
- Multiple keys per token

## Installation

1. `go get github.com/soellman/stargate`
1. Move the binary wherever you like.
1. Copy the example config (`example/stargate.yaml`) to `/etc/stargate.yaml`, and edit as you like. Make it 0400.

## Running

Start it up (e.g. `nohup sudo stargate`). You'll need to run as root - it requires iptables and has passwords in the config file.

## Notes

- Make sure you enable ip forwarding: `sysctl -w net.ipv4.ip_forward=1`
- It logs to stdout, redirect as you please.
- When you stop stargate, it will remove all access from the managed network
- Logging in only provides access until the token expires or stargate is stopped/restarted
