package treblle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// getMaskedQueryString masks sensitive query parameters
func getMaskedQueryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}

	// Create a copy of the query values to avoid modifying the original
	maskedQuery := make(url.Values)
	for key, values := range query {
		if shouldMaskField(key) {
			maskedValues := make([]string, len(values))
			for i := range values {
				maskedValues[i] = maskValue(values[i], key).(string)
			}
			maskedQuery[key] = maskedValues
		} else {
			maskedQuery[key] = values
		}
	}

	return maskedQuery.Encode()
}

// getMaskedJSON masks sensitive fields in JSON data
func getMaskedJSON(data []byte) (json.RawMessage, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// Return the original error from json.Unmarshal
		return nil, err
	}

	maskedData := maskData(jsonData)
	maskedJSON, err := json.Marshal(maskedData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal masked data: %v", err)
	}

	return maskedJSON, nil
}

// maskMap masks sensitive fields in a map based on configuration
func maskMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		// Check if this key should be masked
		if shouldMaskField(strings.ToLower(key)) {
			switch v := value.(type) {
			case string:
				result[key] = maskValue(v, key)
			case []interface{}:
				// If it's an array of strings, mask each element
				strArray := make([]string, len(v))
				for i, elem := range v {
					if str, ok := elem.(string); ok {
						strArray[i] = maskValue(str, key).(string)
					}
				}
				result[key] = strArray
			default:
				// For non-string values that need masking, convert to JSON string and mask
				if jsonStr, err := json.Marshal(v); err == nil {
					result[key] = strings.Repeat("*", len(string(jsonStr)))
				} else {
					result[key] = "****"
				}
			}
		} else {
			result[key] = maskData(value)
		}
	}
	return result
}

// maskValue masks a string value based on its type
func maskValue(value interface{}, key string) interface{} {
	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			return v
		}
		// Special handling for Authorization header
		if strings.EqualFold(key, "Authorization") {
			// Check for common auth types
			authTypes := []string{"Bearer", "Basic", "ApiKey", "Token"}
			for _, authType := range authTypes {
				if strings.HasPrefix(v, authType+" ") {
					return authType + " " + strings.Repeat("*", 9)
				}
			}
			// No auth type prefix found, mask entire value
			return strings.Repeat("*", 9)
		}
		return strings.Repeat("*", 9)
	case []string:
		maskedValues := make([]interface{}, len(v))
		for i := range v {
			if str := v[i]; len(str) > 0 {
				maskedValues[i] = strings.Repeat("*", 9)
			} else {
				maskedValues[i] = str
			}
		}
		return maskedValues
	case []interface{}:
		maskedValues := make([]interface{}, len(v))
		for i := range v {
			if str, ok := v[i].(string); ok && len(str) > 0 {
				maskedValues[i] = strings.Repeat("*", 9)
			} else {
				maskedValues[i] = ""
			}
		}
		return maskedValues
	default:
		return strings.Repeat("*", 9)
	}
}

// maskData recursively masks data in different formats
func maskData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return maskMap(v)
	case []interface{}:
		return maskArray(v)
	default:
		return v
	}
}

// maskArray handles masking of JSON arrays
func maskArray(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = maskData(v)
	}
	return result
}

// shouldMaskField checks if a field should be masked based on configuration
func shouldMaskField(fieldName string) bool {
	// Convert field name to lowercase for consistent matching
	fieldName = strings.ToLower(fieldName)

	// Check direct match
	if _, exists := Config.FieldsMap[fieldName]; exists {
		return true
	}

	// Check with common prefixes
	prefixes := []string{"x-", "x_"}
	for _, prefix := range prefixes {
		if _, exists := Config.FieldsMap[prefix+fieldName]; exists {
			return true
		}
	}

	return false
}
