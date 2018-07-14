# go-healthserver

Simple health-server framework for your Go service.

## Usage

The health server is very easy to configure:

```go
package main

import . "github.com/Makman2/go-healthserver"
import "errors"
import "time"

func main() {
	hs := HealthServer{
		Address:   "localhost:10000",
		Endpoints: []Endpoint{
			{
				Name:   "health",
				Checks: []Check{
					{
						Name:  "time is correct",
						Check: func() error {
							if time.Now().Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
								return errors.New(
									"I'm quite sure that it's the 21st century we're living in...")
							}
							return nil
						},
					},
				},
			},
		},
	}
	
	hs.Start()
	defer hs.Shutdown()
	
	// Run your service...
}
```

A health server is made up of endpoints, which itself run your configured checks. An endpoint's
name is the URL path you can reach the health-endpoint. In the example above you can reach the
configured check via `localhost:10000/health` once the health server was started. If the URL is
queried, all registered checks defined via `Checks` are invoked and tested for errors. Each check
itself has a name (that you can choose to your likings, but they should be unique for an endpoint)
and a check-function. Check-functions are supposed to return an error if the check failed, and `nil`
otherwise. Multiple health checks registered for a single endpoint are executed in parallel.

The health server responds with a `200` *OK* status code if all registered checks for an endpoint are
successful, and with a `503` *Service Unavailable* if not.
