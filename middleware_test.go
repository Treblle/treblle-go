package treblle

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSubTest() {
	s.router = chi.NewRouter()
	s.testServer = httptest.NewServer(s.router)
	s.treblleMockMux = http.NewServeMux()
	s.treblleMockServer = httptest.NewServer(s.treblleMockMux)
	Configure(Configuration{APIKey: "key", ProjectID: "project_id"})
}

func (s *TestSuite) TearDownSubTest() {
	s.testServer.Close()
	s.treblleMockServer.Close()
}

func (s *TestSuite) Test_JsonFormat() {
	content, err := ioutil.ReadFile("sample.json")
	s.Require().NoError(err)
	var treblleMetadata MetaData
	err = json.Unmarshal(content, &treblleMetadata)
	s.Require().NoError(err)

}

func (s *TestSuite) Test_Middleware() {
	testCases := map[string]struct {
		requestJson        string
		responseJson       string
		status             int
		requestHeaderKey   string
		requestHeaderValue string
		respHeaderKey      string
		respHeaderValue    string
		treblleCalled      bool
	}{
		"happy-path": {

			requestJson:        `{"id":2}`,
			responseJson:       `{"id":2, "name":"test"}`,
			status:             http.StatusOK,
			requestHeaderKey:   "Req-K-200",
			requestHeaderValue: "Req-V-200",
			respHeaderKey:      "Resp-K-200",
			respHeaderValue:    "Resp-V-200",
			treblleCalled:      true,
		},
		"status-nok": {
			requestJson:        `{"id":3}`,
			responseJson:       `{"id":2, "name":"test", "errors":true}`,
			status:             http.StatusConflict,
			requestHeaderKey:   "Req-K-409",
			requestHeaderValue: "Req-V-409",
			respHeaderKey:      "Resp-K-409",
			respHeaderValue:    "Resp-V-409",
			treblleCalled:      true,
		},
		"req-not-json": {
			requestJson:   `{"id4`,
			responseJson:  `{}`,
			status:        http.StatusOK,
			treblleCalled: false,
		},
		"resp-not-json": {
			requestJson:   `{"id":5}`,
			responseJson:  `{"`,
			status:        http.StatusOK,
			treblleCalled: true,
		},
	}

	for tn, tc := range testCases {
		s.SetupSubTest()
		treblleMuxCalled := false

		s.treblleMockMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			var treblleMetadata MetaData
			decoder := json.NewDecoder(r.Body)
			s.Require().NoError(decoder.Decode(&treblleMetadata))
			s.Require().Equal(getOsInfo(), treblleMetadata.Data.Server.Os)
			treblleMuxCalled = true
		})

		s.router.With(Middleware).Get("/test", func(w http.ResponseWriter, r *http.Request) {
			s.Require().Equal(tc.requestHeaderValue, r.Header.Get(tc.requestHeaderKey))
			w.Header()[tc.respHeaderKey] = []string{tc.respHeaderValue}
			w.WriteHeader(tc.status)
			_, err := w.Write([]byte(tc.responseJson))
			if err != nil {
				return
			}
		})

		requestHeaders := map[string]string{}
		if len(tc.requestHeaderKey) > 0 {
			requestHeaders[tc.requestHeaderKey] = tc.requestHeaderValue
		}
		resp, respBody := s.testRequest(http.MethodGet, "/test", tc.requestJson, requestHeaders)
		s.Require().Equal(tc.responseJson, respBody, tn)
		s.Require().Equal(tc.status, resp.StatusCode, tn)
		s.Require().Equal(tc.respHeaderValue, resp.Header.Get(tc.respHeaderKey), tn)

		// wait  the async treblle call to finish
		time.Sleep(1 * time.Second)

		s.Require().Equal(tc.treblleCalled, treblleMuxCalled, tn)
		s.TearDownSubTest()
	}
}

func (s *TestSuite) testRequest(method, path, body string, headers map[string]string) (*http.Response, string) {
	req, err := http.NewRequest(method, s.testServer.URL+path, bytes.NewBuffer([]byte(body)))
	s.Require().NoError(err)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	respBody, err := ioutil.ReadAll(resp.Body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	return resp, string(respBody)
}
