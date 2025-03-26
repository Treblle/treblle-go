package treblle

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectFirstValidIPv4(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "127.0.0.1",
		},
		{
			name:     "single valid IPv4",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "multiple IPv4 addresses",
			input:    "192.168.1.1, 10.0.0.1, 172.16.0.1",
			expected: "192.168.1.1",
		},
		{
			name:     "IPv6 and IPv4 mixed",
			input:    "2001:db8::1, 192.168.1.1, 10.0.0.1",
			expected: "192.168.1.1",
		},
		{
			name:     "only IPv6 addresses",
			input:    "2001:db8::1, 2001:db8::2",
			expected: "2001:db8::1", // Returns first one even if not IPv4
		},
		{
			name:     "with spaces and invalid entries",
			input:    " 192.168.1.1 , invalid, 10.0.0.1",
			expected: "192.168.1.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SelectFirstValidIPv4(tc.input)
			assert.Equal(t, tc.expected, result, "SelectFirstValidIPv4 returned unexpected result")
		})
	}
}

func TestDetectProtocol(t *testing.T) {
	testCases := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name:     "nil request",
			request:  nil,
			expected: "HTTP/1.1",
		},
		{
			name: "HTTP/1.1 request",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil)
				req.Proto = "HTTP/1.1"
				req.ProtoMajor = 1
				req.ProtoMinor = 1
				return req
			}(),
			expected: "HTTP/1.1",
		},
		{
			name: "HTTP/2.0 request",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil)
				req.Proto = "HTTP/2.0"
				req.ProtoMajor = 2
				req.ProtoMinor = 0
				return req
			}(),
			expected: "HTTP/2.0",
		},
		{
			name: "empty Proto but valid ProtoMajor/Minor",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil)
				req.Proto = ""
				req.ProtoMajor = 1
				req.ProtoMinor = 0
				return req
			}(),
			expected: "HTTP/1.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectProtocol(tc.request)
			assert.Equal(t, tc.expected, result, "DetectProtocol returned unexpected result")
		})
	}
}
