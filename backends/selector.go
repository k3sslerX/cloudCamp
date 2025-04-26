package backends

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type backendsStruct struct {
	Backends []string `json:"backends"`
}

var selectorMu sync.Mutex
var backends []*url.URL
var currentBack int

func init() {
	serversFile, err := os.Open("./servers.json")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = serversFile.Close()
	}()

	data, err := io.ReadAll(serversFile)
	if err != nil {
		log.Fatal(err)
	}
	backendStruct := backendsStruct{}

	err = json.Unmarshal(data, &backendStruct)
	if err != nil {
		log.Fatal(err)
	}

	err = parseBackends(backendStruct)
	if err != nil {
		log.Fatal(err)
	}
}

func parseBackends(backs backendsStruct) error {
	for _, backend := range backs.Backends {
		u, err := url.Parse(backend)
		if err != nil {
			return err
		}
		backends = append(backends, u)
	}
	return nil
}

func checkBackendHealth(u *url.URL) bool {
	resp, err := http.Get(u.String() + "/health")
	if err != nil {
		log.Printf("backend %s unavailable: %v\n", u.Host, err)
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return resp.StatusCode == http.StatusOK
}

func GetBackend() *url.URL {
	selectorMu.Lock()
	defer selectorMu.Unlock()

	start := currentBack
	for {
		backend := backends[currentBack]
		currentBack = (currentBack + 1) % len(backends)

		if checkBackendHealth(backend) {
			return backend
		}

		if currentBack == start {
			log.Println("all backends unavailable")
			return nil
		}
	}
}
