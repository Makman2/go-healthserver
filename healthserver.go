package healthserver

import (
	"context"
	"net"
	"net/http"
)

// HealthServer allows you to configure and run the health server.
type HealthServer struct {
	Address   string
	Endpoints []Endpoint
	server    *http.Server
}

// Endpoint specifies a certain endpoint and its health checks.
type Endpoint struct {
	Name   string
	Checks []Check
}

// Check specifies a health check.
type Check struct {
	Name  string
	Check func() error
}

func check(checks []Check, response http.ResponseWriter) {
	for _, check := range checks {
		err := check.Check()
		if err != nil {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}

	response.WriteHeader(http.StatusOK)
}

// Start starts the health server asynchronously, thus the function exits immediately. To shut it
// down again, use Shutdown. The health server is immediately able to handle connections when the
// function has finished.
func (hs *HealthServer) Start() error {
	handler := http.NewServeMux()
	server := &http.Server{Handler: handler}

	for i := range hs.Endpoints {
		endpoint := hs.Endpoints[i]
		handler.HandleFunc("/"+endpoint.Name, func(response http.ResponseWriter, request *http.Request) {
			check(endpoint.Checks, response)
		})
	}

	hs.server = server

	listener, err := net.Listen("tcp", hs.Address)
	if err != nil {
		return err
	}

	go server.Serve(listener)

	return nil
}

// Shutdown stops the health server gracefully. If the health server hasn't been started yet, this
// function is a no-op.
func (hs *HealthServer) Shutdown() {
	if hs.server != nil {
		hs.server.Shutdown(context.Background())
		hs.server = nil
	}
}
