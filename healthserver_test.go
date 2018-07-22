package healthserver

import (
	"errors"
	"fmt"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
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
		Body         string
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
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
			},
		},
		{
			Name: "Single endpoint - Single successful check - ResponseModePlain (default)",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: SuccessfulCheckFunc},
					},
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusOK,
					Body:         "",
				},
			},
		},
		{
			Name: "Single endpoint - Single successful check - ResponseModeStatusName",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: SuccessfulCheckFunc},
					},
					ResponseMode: ResponseModeStatusName,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusOK,
					Body:         "OK",
				},
			},
		},
		{
			Name: "Single endpoint - Single successful check - ResponseModeReport",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: SuccessfulCheckFunc},
					},
					ResponseMode: ResponseModeReport,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusOK,
					Body: "<meta charset=utf-8>" +
						"<title>Health Status</title>" +
						"<style>" +
						"table{border-collapse:collapse}" +
						"tr{height:2em}" +
						"td{padding-left:.7em;padding-right:.7em}" +
						".status{text-align:center}" +
						".failing{background-color:red}" +
						".passing{background-color:#7cfc00}" +
						"</style>" +
						"<table>" +
						"<tr class=\"passing\">" +
						"<td class=status>&#x2714" +
						"<td>test-check\n" +
						"</table>",
				},
			},
		},
		{
			Name: "Single endpoint - Single failing check - ResponseModePlain",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: UnsuccessfulCheckFunc},
					},
					ResponseMode: ResponseModePlain,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "",
				},
			},
		},
		{
			Name: "Single endpoint - Single failing check - ResponseModeStatusName",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: UnsuccessfulCheckFunc},
					},
					ResponseMode: ResponseModeStatusName,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "Service Unavailable",
				},
			},
		},
		{
			Name: "Single endpoint - Single failing check - ResponseModeReport",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check", Check: UnsuccessfulCheckFunc},
					},
					ResponseMode: ResponseModeReport,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusServiceUnavailable,
					Body: "<meta charset=utf-8>" +
						"<title>Health Status</title>" +
						"<style>" +
						"table{border-collapse:collapse}" +
						"tr{height:2em}" +
						"td{padding-left:.7em;padding-right:.7em}" +
						".status{text-align:center}" +
						".failing{background-color:red}" +
						".passing{background-color:#7cfc00}" +
						"</style>" +
						"<table>" +
						"<tr class=\"failing\">" +
						"<td class=status>&#x2718" +
						"<td>test-check\n" +
						"</table>",
				},
			},
		},
		{
			Name: "Single endpoint - Single successful single unsuccessful check - ResponseModePlain",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
						{Name: "test-check2", Check: UnsuccessfulCheckFunc},
					},
					ResponseMode: ResponseModePlain,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "",
				},
			},
		},
		{
			Name: "Single endpoint - Single successful single unsuccessful check - ResponseModeStatusName",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
						{Name: "test-check2", Check: UnsuccessfulCheckFunc},
					},
					ResponseMode: ResponseModeStatusName,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "Service Unavailable",
				},
			},
		},
		{
			Name: "Single endpoint - Single successful single unsuccessful check - ResponseModeReport",
			Endpoints: []Endpoint{
				{
					Name: "health",
					Checks: []Check{
						{Name: "test-check1", Check: SuccessfulCheckFunc},
						{Name: "test-check2", Check: UnsuccessfulCheckFunc},
					},
					ResponseMode: ResponseModeReport,
				},
			},
			EndpointTests: []EndpointTest{
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusServiceUnavailable,
					Body: "<meta charset=utf-8>" +
						"<title>Health Status</title>" +
						"<style>" +
						"table{border-collapse:collapse}" +
						"tr{height:2em}" +
						"td{padding-left:.7em;padding-right:.7em}" +
						".status{text-align:center}" +
						".failing{background-color:red}" +
						".passing{background-color:#7cfc00}" +
						"</style>" +
						"<table>" +
						"<tr class=\"passing\">" +
						"<td class=status>&#x2714" +
						"<td>test-check1\n" +
						"<tr class=\"failing\">" +
						"<td class=status>&#x2718" +
						"<td>test-check2\n" +
						"</table>",
				},
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
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health/test-endpoint1",
					StatusCode:   http.StatusOK,
					Body:         "",
				},
				{
					EndpointPath: "health/test-endpoint2",
					StatusCode:   http.StatusOK,
					Body:         "",
				},
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
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health/test-endpoint1",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "",
				},
				{
					EndpointPath: "health/test-endpoint2",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "",
				},
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
				{
					EndpointPath: "",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health",
					StatusCode:   http.StatusNotFound,
					Body:         "404 page not found\n",
				},
				{
					EndpointPath: "health/test-endpoint1",
					StatusCode:   http.StatusServiceUnavailable,
					Body:         "",
				},
				{
					EndpointPath: "health/test-endpoint2",
					StatusCode:   http.StatusOK,
					Body:         "",
				},
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
				testName := endpointTest.EndpointPath
				if len(testName) == 0 {
					testName = "\"\""
				}

				t.Run(testName, func(t *testing.T) {
					response, err := http.Get("http://" + address + "/" + endpointTest.EndpointPath)
					require.NoError(t, err)
					require.Equal(t, endpointTest.StatusCode, response.StatusCode)

					body, err := ioutil.ReadAll(response.Body)
					require.NoError(t, err)
					require.Equal(t, endpointTest.Body, string(body))
				})
			}
		})
	}
}
