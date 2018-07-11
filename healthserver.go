package healthserver

import "net/http"

type HealthServer struct {
	Address   string
	Endpoints []Endpoint
}

// TODO Feature: Display specific checks / errors in an HTML page. Or only display first failing.
// TODO   Maybe a display-settings struct or enum or so configurable for each endpoint.
type Endpoint struct {
	Name   string
	Checks []Check
}

type Check interface {
	Check() error
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

func (hs *HealthServer) Start() {
	handler := http.NewServeMux()

	for _, endpoint := range hs.Endpoints {
		handler.HandleFunc("/"+endpoint.Name, func(response http.ResponseWriter, request *http.Request) {
			check(endpoint.Checks, response)
		})
	}

	http.ListenAndServe(hs.Address, handler)
}
