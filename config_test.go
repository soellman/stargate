package main

import "testing"

func TestConfig(t *testing.T) {
	_, err := ParseConfigFile("example/stargate.yaml")
	if err != nil {
		t.Errorf("file failed: %v", err)
	}
}
