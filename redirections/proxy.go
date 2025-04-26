package redirections

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func reverseProxy(backend *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(backend)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("reverse proxy error: %v\n", err)
		w.WriteHeader(http.StatusBadGateway)
		_, err = w.Write([]byte("Service currently unavailable"))
		if err != nil {
			return
		}
	}

	return proxy
}
