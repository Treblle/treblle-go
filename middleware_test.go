package treblle

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	testServer *httptest.Server
	router     *chi.Mux

	treblleMockMux    *http.ServeMux
	treblleMockServer *httptest.Server
}

func TestTreblleTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.router = chi.NewRouter()
	s.router.Use(Middleware)
	s.testServer = httptest.NewServer(s.router)
	s.treblleMockMux = http.NewServeMux()
	s.treblleMockServer = httptest.NewServer(s.treblleMockMux)
	Configure(Configuration{
		APIKey:    "test-api-key",
		ProjectID: "test-project",
		DefaultFieldsToMask: []string{
			"password",
			"api_key",
			"credit_card",
			"authorization",
		},
	})
}

func (s *TestSuite) TearDownTest() {
	if s.testServer != nil {
		s.testServer.Close()
	}
	if s.treblleMockServer != nil {
		s.treblleMockServer.Close()
	}
}

func (s *TestSuite) TestJsonFormat() {
	sampleData := map[string]interface{}{
		"api_key":    "",
		"project_id": "",
		"version":    "0.6",
		"sdk":        "laravel",
		"data": map[string]interface{}{
			"server": map[string]interface{}{
				"ip":        "18.194.223.176",
				"timezone":  "UTC",
				"software":  "Apache",
				"signature": "Apache/2.4.2",
				"protocol":  "HTTP/1.1",
				"os": map[string]interface{}{
					"name":         "Linux",
					"release":      "4.14.186-110.268.amzn1.x86_64",
					"architecture": "x86_64",
				},
			},
		},
	}

	content, err := json.Marshal(sampleData)
	s.Require().NoError(err)

	var treblleMetadata MetaData
	err = json.Unmarshal(content, &treblleMetadata)
	s.Require().NoError(err)
}

func (s *TestSuite) testRequest(method, path, body string, headers map[string]string) (*http.Response, string) {
	var bodyReader *bytes.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	}

	req, err := http.NewRequest(method, s.testServer.URL+path, bodyReader)
	s.Require().NoError(err)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	respBody, err := ioutil.ReadAll(resp.Body)
	s.Require().NoError(err)
	resp.Body.Close()

	return resp, string(respBody)
}

func (s *TestSuite) TestCRUDMasking() {
	s.router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var requestBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		s.Require().NoError(err)

		// Mask sensitive data in response
		response := map[string]interface{}{
			"id":       1,
			"password": maskValue("should-be-masked", "password"),
			"api_key":  maskValue("should-be-masked", "api_key"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})

	// Test POST request with sensitive data
	resp, body := s.testRequest("POST", "/users", `{
		"username": "test",
		"password": "secret123",
		"api_key": "key123"
	}`, map[string]string{
		"Content-Type": "application/json",
	})

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var responseBody map[string]interface{}
	err := json.Unmarshal([]byte(body), &responseBody)
	s.Require().NoError(err)
	s.Require().Equal("*********", responseBody["password"])
	s.Require().Equal("*********", responseBody["api_key"])
}

func (s *TestSuite) TestMiddleware() {
	testCases := map[string]struct {
		requestJson        string
		responseJson       string
		requestHeaderKey   string
		requestHeaderValue string
		respHeaderKey      string
		respHeaderValue    string
		status             int
		treblleCalled      bool
	}{
		"happy-path": {
			requestJson:   `{"id":1}`,
			responseJson:  `{"id":1}`,
			status:        http.StatusOK,
			treblleCalled: true,
		},
		"invalid-request-json": {
			requestJson:   `{"id":`,
			responseJson:  `{"error":"bad request"}`,
			status:        http.StatusBadRequest,
			treblleCalled: false,
		},
		"non-json-response": {
			requestJson:   `{"id":5}`,
			responseJson:  `Hello, World!`,
			status:        http.StatusOK,
			treblleCalled: true,
		},
	}

	for tn, tc := range testCases {
		s.SetupTest()
		treblleMuxCalled := false

		mockURL := s.treblleMockServer.URL
		log.Printf("Test case: %s, Mock URL: %s", tn, mockURL)

		Configure(Configuration{
			APIKey:              "test-api-key",
			ProjectID:           "test-project-id",
			DefaultFieldsToMask: []string{"password"},
			Endpoint:            mockURL,
		})

		s.treblleMockMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Mock server received request to: %s", r.URL.String())
			var treblleMetadata MetaData
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&treblleMetadata)
			if err != nil {
				log.Printf("Error decoding request body in mock server: %v", err)
				return
			}
			log.Printf("Received metadata - APIKey: %s, ProjectID: %s", treblleMetadata.ApiKey, treblleMetadata.ProjectID)
			s.Require().Equal("test-api-key", treblleMetadata.ApiKey)
			s.Require().Equal("test-project-id", treblleMetadata.ProjectID)

			if tn == "non-json-response" {
				// For non-JSON responses, the body should be a JSON string
				expectedBody, _ := json.Marshal("Hello, World!")
				s.Require().Equal(string(expectedBody), string(treblleMetadata.Data.Response.Body))
			}

			treblleMuxCalled = true
			w.WriteHeader(http.StatusOK)
		})

		s.router.Use(Middleware)
		s.router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request headers: %+v", r.Header)
			if tc.requestHeaderKey != "" {
				s.Require().Equal(tc.requestHeaderValue, r.Header.Get(tc.requestHeaderKey))
			}
			if tc.respHeaderKey != "" {
				w.Header().Set(tc.respHeaderKey, tc.respHeaderValue)
			}
			if tn == "non-json-response" {
				w.Header().Set("Content-Type", "text/plain")
			} else {
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(tc.status)
			w.Write([]byte(tc.responseJson))
		})

		requestHeaders := map[string]string{}
		if tc.requestHeaderKey != "" {
			requestHeaders[tc.requestHeaderKey] = tc.requestHeaderValue
		}
		requestHeaders["Content-Type"] = "application/json"

		resp, body := s.testRequest(http.MethodGet, "/test", tc.requestJson, requestHeaders)
		log.Printf("Response status: %d, body: %s", resp.StatusCode, body)

		s.Require().Equal(tc.status, resp.StatusCode, tn)
		s.Require().Equal(tc.responseJson, body, tn)
		if tc.respHeaderKey != "" {
			s.Require().Equal(tc.respHeaderValue, resp.Header.Get(tc.respHeaderKey), tn)
		}

		// Wait for the async Treblle call to finish
		time.Sleep(1 * time.Second)
		log.Printf("After sleep - treblleMuxCalled: %v, expected: %v", treblleMuxCalled, tc.treblleCalled)
		s.Require().Equal(tc.treblleCalled, treblleMuxCalled, tn)

		s.TearDownTest()
	}
}
