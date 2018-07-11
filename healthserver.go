package healthserver

import "net/http"

// TODO Feature: Display specific checks / errors in an HTML page. Or only display first failing.
type Endpoint struct {
	Name   string
	Checks []Check
}

type Check interface {
	Check() error
}

type HealthServer struct {
	Address   string
	Endpoints []Endpoint
}

/*
func (hs *HealthServer) AddCheck(check Check, endpoint string) {
	hs.checks[endpoint] = append(hs.checks[endpoint], check)
}
*/

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
		handler.HandleFunc("health/" + endpoint.Name, func(response http.ResponseWriter, request *http.Request) {
			check(endpoint.Checks, response)
		})
	}

	http.ListenAndServe(hs.Address, handler)
}
