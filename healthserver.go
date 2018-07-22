package healthserver

import (
	"bytes"
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

type ResponseMode int

const (
	// ResponseModePlain displays no body at all for a healthcheck endpoint, just the status code.
	ResponseModePlain = iota
	// ResponseModeStatusName displays a status description in the response body.
	ResponseModeStatusName
	// ResponseModeReport displays a detailed report consisting of all registered check names and
	// their results.
	ResponseModeReport
)

func (rm ResponseMode) String() string {
	switch rm {
	case ResponseModePlain:
		return "ResponseModePlain"
	case ResponseModeStatusName:
		return "ResponseModeStatusName"
	case ResponseModeReport:
		return "ResponseModeReport"
	}
	return "UNKNOWN"
}


// Endpoint specifies a certain endpoint and its health checks.
type Endpoint struct {
	Name         string
	Checks       []Check
	ResponseMode ResponseMode
}

// Check specifies a health check.
type Check struct {
	Name  string
	Check func() error
}

// CheckResult describes the result of a check.
type CheckResult struct {
	Name string
	Err  error
}

// Check invokes all registered checks for the given endpoint.
func (ep Endpoint) Check() []CheckResult {
	type ErrorResult struct {
		CheckName string
		Err       error
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

	var errors []CheckResult
	for _, channel := range channels {
		result := <-channel
		errors = append(errors, CheckResult{Name: result.CheckName, Err: result.Err})
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
			checkResults := endpoint.Check()

			status := http.StatusOK

			for _, checkResult := range checkResults {
				if checkResult.Err != nil {
					status = http.StatusServiceUnavailable
					break
				}
			}

			const contentTypeText = "text/plain"
			const contentTypeHtml = "text/html"

			var body []byte
			var contentType string

			// Generate response body.
			switch endpoint.ResponseMode {
			case ResponseModePlain:
				contentType = contentTypeText
				body = nil
			case ResponseModeStatusName:
				contentType = contentTypeText
				body = []byte(http.StatusText(status))
			case ResponseModeReport:
				contentType = contentTypeHtml
				var temporaryBuffer bytes.Buffer

				err := getReportTemplate().Execute(&temporaryBuffer, checkResults)
				if err != nil {
					panic(err)
				}

				body = temporaryBuffer.Bytes()
			}

			response.Header().Add("Content-Type", contentType+"; charset=utf-8")
			response.WriteHeader(status)
			response.Write(body)
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
