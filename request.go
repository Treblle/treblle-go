package treblle

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type RequestInfo struct {
	Timestamp string          `json:"timestamp"`
	Ip        string          `json:"ip"`
	Url       string          `json:"url"`
	UserAgent string          `json:"user_agent"`
	Method    string          `json:"method"`
	Headers   json.RawMessage `json:"headers"`
	Body      json.RawMessage `json:"body"`
}

var ErrNotJson = errors.New("request body is not JSON")

// Get details about the request
func getRequestInfo(r *http.Request, startTime time.Time) (RequestInfo, error) {
	defer dontPanic()

	headers := make(map[string]string)
	for k := range r.Header {
		headers[k] = r.Header.Get(k)
	}

	protocol := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
		protocol = "https"
	}
	fullURL := protocol + "://" + r.Host + r.URL.RequestURI()
	ip := extractIP(r.RemoteAddr)

	ri := RequestInfo{
		Timestamp: startTime.Format("2006-01-02 15:04:05"),
		Ip:        ip,
		Url:       fullURL,
		UserAgent: r.UserAgent(),
		Method:    r.Method,
	}

	if r.Body != nil && r.Body != http.NoBody {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return ri, err
		}
		// open 2 NopClosers over the buffer to allow buffer to be read and still passed on
		bodyReaderOriginal := ioutil.NopCloser(bytes.NewBuffer(buf))
		// restore the original request body once done processing
		defer recoverBody(r, ioutil.NopCloser(bytes.NewBuffer(buf)))

		body, err := ioutil.ReadAll(bodyReaderOriginal)
		if err != nil {
			return ri, err
		}

		// mask all the JSON fields listed in Config.FieldsToMask
		sanitizedBody, err := getMaskedJSON(body)
		if err != nil {
			return ri, err
		}

		ri.Body = sanitizedBody

		headersJson, err := json.Marshal(headers)
		if err != nil {
			return ri, err
		}

		sanitizedHeaders, err := getMaskedJSON(headersJson)

		if err != nil {
			return ri, err
		}
		ri.Headers = sanitizedHeaders
	}
	return ri, nil
}

func recoverBody(r *http.Request, bodyReaderCopy io.ReadCloser) {
	r.Body = bodyReaderCopy
}

func getMaskedJSON(payloadToMask []byte) (json.RawMessage, error) {
	jsonMap := make(map[string]interface{})
	if err := json.Unmarshal(payloadToMask, &jsonMap); err != nil {
		// not a valid json request
		return nil, ErrNotJson
	}

	sanitizedJson := make(map[string]interface{})
	copyAndMaskJson(jsonMap, sanitizedJson)
	jsonData, err := json.Marshal(sanitizedJson)
	if err != nil {
		return nil, err
	}

	rawMessage := json.RawMessage(jsonData)

	return rawMessage, nil
}

func copyAndMaskJson(src map[string]interface{}, dest map[string]interface{}) {
	for key, value := range src {
		switch src[key].(type) {
		case map[string]interface{}:
			dest[key] = map[string]interface{}{}
			copyAndMaskJson(src[key].(map[string]interface{}), dest[key].(map[string]interface{}))
		default:
			// if JSON key is in the list of keys to mask, replace it with a * string of the same length
			_, exists := Config.FieldsMap[key]
			if exists {
				maskedValue := maskValue(value.(string), key)
				dest[key] = maskedValue
			} else {
				dest[key] = value
			}
		}
	}
}
func maskValue(valueToMask string, key string) string {
	keyLower := strings.ToLower(key)

	if keyLower == "authorization" && regexp.MustCompile(`(?i)^(bearer|basic)\s+`).MatchString(valueToMask) {
		authParts := strings.SplitN(valueToMask, " ", 2)
		authPrefix := authParts[0]
		authToken := authParts[1]
		maskedAuthToken := strings.Repeat("*", len(authToken))
		maskedValue := authPrefix + " " + maskedAuthToken
		return maskedValue
	}

	return strings.Repeat("*", len(valueToMask))
}

func extractIP(remoteAddr string) string {
	// If RemoteAddr contains both IP and port, split and return the IP
	if strings.Contains(remoteAddr, ":") {
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return ip
		}
	}

	return remoteAddr
}
