package backends

import (
	"log"
	"time"
)

func HealthChecker() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			for _, backend := range backends {
				isAlive := checkBackendHealth(backend)
				log.Printf("backend %s available: %v\n", backend.Host, isAlive)
			}
		}
	}
}
