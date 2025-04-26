package redirections

import (
	"cloudCamp/backends"
	"log"
	"net/http"
)

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	backend := backends.GetBackend()
	if backend == nil {
		http.Error(w, "Service currently unavailable", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Redirecting to %s", backend.Host)
	proxy := reverseProxy(backend)
	proxy.ServeHTTP(w, r)
}
