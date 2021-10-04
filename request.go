package treblle

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type RequestInfo struct {
	Timestamp string                 `json:"timestamp"`
	Ip        string                 `json:"ip"`
	Url       string                 `json:"url"`
	UserAgent string                 `json:"user_agent"`
	Method    string                 `json:"method"`
	Headers   map[string]string      `json:"headers"`
	Body      map[string]interface{} `json:"body"`
}

// Get details about the request
func getRequestInfo(r *http.Request, startTime time.Time) (RequestInfo, error) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return RequestInfo{}, err
	}
	bodyReaderOriginal := ioutil.NopCloser(bytes.NewBuffer(buf))
	bodyReaderCopy := ioutil.NopCloser(bytes.NewBuffer(buf))

	body, err := ioutil.ReadAll(bodyReaderOriginal)
	if err != nil {
		return RequestInfo{}, err
	}

	sanitizedJsonString, err := getMaskedJSON(body)
	if err != nil {
		return RequestInfo{}, err
	}

	headers := make(map[string]string)
	for k, _ := range r.Header {
		headers[k] = r.Header.Get(k)
	}

	ri := RequestInfo{
		Timestamp: startTime.Format("2006-01-02 15:04:05"),
		Ip:        r.RemoteAddr,
		Url:       r.RequestURI,
		UserAgent: r.UserAgent(),
		Method:    r.Method,
		Headers:   headers,
		Body:      sanitizedJsonString,
	}

	r.Body = bodyReaderCopy

	return ri, nil
}

func getMaskedJSON(body []byte) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	if err := json.Unmarshal(body, &jsonMap); err != nil {
		// not a valid json request
		return nil, ErrNotJson
	}

	sanitizedJson := make(map[string]interface{})
	copyAndMaskJson(jsonMap, sanitizedJson)

	return sanitizedJson, nil
}

func copyAndMaskJson(src map[string]interface{}, dest map[string]interface{}) {
	for key, value := range src {
		switch src[key].(type) {
		case map[string]interface{}:
			dest[key] = map[string]interface{}{}
			copyAndMaskJson(src[key].(map[string]interface{}), dest[key].(map[string]interface{}))
		default:
			// if JSON key is in the list of keys to mask, replace it with a * string of the same length
			_, exists := Config.KeysMap[key]
			if exists {
				re := regexp.MustCompile(".")
				maskedValue := re.ReplaceAllString(value.(string), "*")
				dest[key] = maskedValue
			} else {
				dest[key] = value
			}
		}
	}
}
