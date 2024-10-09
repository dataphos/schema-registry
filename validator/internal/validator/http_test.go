package validator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestEstimateHTTPTimeout(t *testing.T) {
	tt := []struct {
		name            string
		size            int
		timeout         time.Duration
		adjustedTimeout time.Duration
	}{
		{"lower than base", 99 * 1024, 1 * time.Second, 1 * time.Second},
		{"1 byte", 1, 1 * time.Second, 1 * time.Second},
		{"equal to base", HTTPTimeoutBytesUnit, 1 * time.Second, 1 * time.Second},
		{"closer to 1", 149 * 1024, 1 * time.Second, 1 * time.Second},
		{"closer to 2", 151 * 1024, 1 * time.Second, 2 * time.Second},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			adjusted := EstimateHTTPTimeout(tc.size, tc.timeout)
			if adjusted != tc.adjustedTimeout {
				t.Error("calculated timeout not the same as expected")
			}
		})
	}
}

func TestValidateOverHttp(t *testing.T) {
	tt := []struct {
		name             string
		expectedRequest  validationRequest
		expectedResponse []byte
		statusCode       int
		isValid          bool
	}{
		{
			"valid with status code 200",
			validationRequest{
				Data:   "data sent as the request for validation",
				Schema: "schema for the data to be validated against",
			},
			[]byte("{\"validation\":true,\"info\":\"\"}"),
			http.StatusOK,
			true,
		},
		{
			"invalid with status code 200",
			validationRequest{
				Data:   "data sent as the request for validation",
				Schema: "schema for the data to be validated against",
			},
			[]byte("{\"validation\":false,\"info\":\"\"}"),
			http.StatusOK,
			false,
		},
		{
			"bad request",
			validationRequest{
				Data:   "data sent as the request for validation",
				Schema: "schema for the data to be validated against",
			},
			[]byte("{\"validation\":false,\"info\":\"\"}"),
			http.StatusBadRequest,
			false,
		},
		{
			"valid but malformed response because json is missing closing bracket",
			validationRequest{
				Data:   "data sent as the request for validation",
				Schema: "schema for the data to be validated against",
			},
			[]byte("{\"validation\":true,\"info\":\"\""),
			http.StatusBadRequest,
			false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				defer request.Body.Close()

				var receivedRequest validationRequest
				if err := json.NewDecoder(request.Body).Decode(&receivedRequest); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(receivedRequest, tc.expectedRequest) {
					t.Fatal("expected and actual request not the same")
				}

				writer.WriteHeader(tc.statusCode)
				writer.Write(tc.expectedResponse)
			})
			srv := httptest.NewServer(handler)
			defer srv.Close()

			isValid, err := ValidateOverHTTP(context.Background(), []byte(tc.expectedRequest.Data), []byte(tc.expectedRequest.Schema), srv.URL)
			if err != nil {
				if tc.statusCode == http.StatusOK {
					t.Fatal("error not expected", err)
				}
			}

			if isValid != tc.isValid {
				t.Fatal("expected and actual validation result not the same")
			}
		})
	}
}
