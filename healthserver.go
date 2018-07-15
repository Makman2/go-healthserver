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

func (ep Endpoint) Check() map[string]error {
	type ErrorResult struct {
		CheckName string
		Err   error
	}

	var channels []chan ErrorResult
	for _, check := range ep.Checks {
		channel := make(chan ErrorResult)
		channels = append(channels, channel)

		go func(check Check) {
			channel <- ErrorResult{CheckName: check.Name, Err: check.Check()}
			close(channel)
		}(check)
	}

	errors := make(map[string]error)
	for _, channel := range channels {
		result := <- channel
		errors[result.CheckName] = result.Err
	}

	return errors
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
			errors := endpoint.Check()

			response.Header().Add("Content-Type", "text/plain; charset=utf-8")

			for _, err := range errors {
				if err != nil {
					response.WriteHeader(http.StatusServiceUnavailable)
					return
				}
			}

			response.WriteHeader(http.StatusOK)
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
