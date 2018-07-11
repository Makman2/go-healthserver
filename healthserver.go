package healthserver

import (
	"fmt"
	"net/http"
)

type HealthServer struct {
	Address   string
	Endpoints []Endpoint
}

type ResponseMode int

const (
	ResponseModeNone ResponseMode = iota
	ResponseModeSimpleText
	ResponseModeFirstError
	ResponseModeFullReport
)

// TODO Feature: Display specific checks / errors in an HTML page. Or only display first failing.
// TODO   Maybe a display-settings struct or enum or so configurable for each endpoint.
type Endpoint struct {
	Name         string
	Checks       []Check
	ResponseMode ResponseMode
}

// TODO Cached checks: Periodically executed, and if accessed via the endpoint a cached result is returned.
// TODO   For cached checks, it might be problematic to use with Kubernetes, different time rates should be defined until the service is alive.

func (ep Endpoint) Check() []error {
	var errs []error
	for _, check := range ep.Checks {
		err := check.Check()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (ep Endpoint) CheckUntilFirstError() error {
	for _, check := range ep.Checks {
		err := check.Check()
		if err != nil {
			return err
		}
	}
	return nil
}

type Check interface {
	Check() error
}

func check(endpoint Endpoint, response http.ResponseWriter) {
	status := http.StatusOK
	body := ""

	if endpoint.ResponseMode == ResponseModeNone ||
		endpoint.ResponseMode == ResponseModeFirstError ||
		endpoint.ResponseMode == ResponseModeSimpleText {
		err := endpoint.CheckUntilFirstError()
		if err != nil {
			status = http.StatusServiceUnavailable

			if endpoint.ResponseMode == ResponseModeFirstError {
				body = err.Error()
			} else if endpoint.ResponseMode == ResponseModeSimpleText {
				body = "UNAVAILABLE"
			}
		} else {
			if endpoint.ResponseMode == ResponseModeSimpleText {
				body = "OK"
			}
		}
	} else if endpoint.ResponseMode == ResponseModeFullReport {
		errs := endpoint.Check()

		if len(errs) > 0 {
			// TODO
		} else {
			// TODO
		}
	}

	response.WriteHeader(status)
	fmt.Fprint(response, body)
}

func (hs *HealthServer) Start() {
	handler := http.NewServeMux()

	for _, endpoint := range hs.Endpoints {
		handler.HandleFunc("/"+endpoint.Name, func(response http.ResponseWriter, request *http.Request) {
			check(endpoint, response)
		})
	}

	http.ListenAndServe(hs.Address, handler)
}
