package healthserver

import (
	"errors"
		"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"github.com/phayes/freeport"
	"fmt"
)

func TestHealthServer(t *testing.T) {
	SuccessfulCheckFunc := func() error {
		return nil
	}

	UnsuccessfulCheckFunc := func() error {
		return errors.New("test-error")
	}

	type EndpointTest struct {
		EndpointPath string
		StatusCode   int
	}

	testCases := []struct {
		Name          string
		Endpoints     []Endpoint
		EndpointTests []EndpointTest
	}{
		{
			Name:      "No endpoints",
			Endpoints: []Endpoint{},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusNotFound},
			},
		},
		{
			Name: "Single endpoint - Single successful check",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: SuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusOK},
			},
		},
		{
			Name: "Single endpoint - Single failing check",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: UnsuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusServiceUnavailable},
			},
		},
		{
			Name: "Single endpoint - Single successful single unsuccessful check",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
						{Name: "test-check2", Check: UnsuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusServiceUnavailable},
			},
		},
		{
			Name: "Multiple endpoints - Single successful checks",
			Endpoints: []Endpoint{
				{
					Name: "health/test-endpoint1",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
					},
				},
				{
					Name: "health/test-endpoint2",
					Checks: []Check{
						{Name: "test-check2", Check: SuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusNotFound},
				{EndpointPath: "health/test-endpoint1", StatusCode: http.StatusOK},
				{EndpointPath: "health/test-endpoint2", StatusCode: http.StatusOK},
			},
		},
		{
			Name: "Multiple endpoints - Single failing checks",
			Endpoints: []Endpoint{
				{
					Name: "health/test-endpoint1",
					Checks: []Check{
						{Name: "test-check1", Check: UnsuccessfulCheckFunc},
					},
				},
				{
					Name: "health/test-endpoint2",
					Checks: []Check{
						{Name: "test-check2", Check: UnsuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusNotFound},
				{EndpointPath: "health/test-endpoint1", StatusCode: http.StatusServiceUnavailable},
				{EndpointPath: "health/test-endpoint2", StatusCode: http.StatusServiceUnavailable},
			},
		},
		{
			Name: "Multiple endpoints - Successful and failing checks",
			Endpoints: []Endpoint{
				{
					Name: "health/test-endpoint1",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
						{Name: "test-check2", Check: UnsuccessfulCheckFunc},
						{Name: "test-check3", Check: SuccessfulCheckFunc},
					},
				},
				{
					Name: "health/test-endpoint2",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
						{Name: "test-check2", Check: SuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{EndpointPath: "", StatusCode: http.StatusNotFound},
				{EndpointPath: "health", StatusCode: http.StatusNotFound},
				{EndpointPath: "health/test-endpoint1", StatusCode: http.StatusServiceUnavailable},
				{EndpointPath: "health/test-endpoint2", StatusCode: http.StatusOK},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			port, err := freeport.GetFreePort()
			if err != nil {
				panic(err)
			}
			address := fmt.Sprintf("localhost:%d", port)

			healthServer := HealthServer{
				Address:   address,
				Endpoints: testCase.Endpoints,
			}
			err = healthServer.Start()
			defer healthServer.Shutdown()
			require.NoError(t, err)

			for _, endpointTest := range testCase.EndpointTests {
				t.Run(endpointTest.EndpointPath, func(t *testing.T) {
					response, err := http.Get("http://" + address + "/" + endpointTest.EndpointPath)
					require.NoError(t, err)
					require.Equal(t, endpointTest.StatusCode, response.StatusCode)
				})
			}
		})
	}
}
