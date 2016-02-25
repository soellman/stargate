package main

import "log"

func debugf(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}

// SyncNetworks copies networks from dst to src
func SyncNetworks(dst Networks, src ListNetworks) {
	// delete unused networks
	for _, dstnet := range dst.Networks() {
		for _, srcnet := range src.Networks() {
			if dstnet.Name == srcnet.Name {
				continue
			}
			dst.RemoveNetwork(dstnet)
		}
	}

	// add remaining networks
	for _, srcnet := range src.Networks() {
		exists := false
		for _, dstnet := range dst.Networks() {
			if srcnet.Name == dstnet.Name {
				exists = true
				break
			}
		}
		if !exists {
			dst.AddNetwork(srcnet)
		}
	}
}
