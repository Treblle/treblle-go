package treblle

import (
	"log"
	"os"
	"strings"
)

var Config internalConfiguration

// Configuration sets up and customizes communication with the Treblle API
type Configuration struct {
	APIKey                 string
	ProjectID              string
	AdditionalFieldsToMask []string
	DefaultFieldsToMask    []string
	MaskingEnabled         bool
	Endpoint               string // Custom endpoint for testing
}

// internalConfiguration is used for communication with Treblle API and contains optimizations
type internalConfiguration struct {
	APIKey                 string
	ProjectID              string
	AdditionalFieldsToMask []string
	DefaultFieldsToMask    []string
	MaskingEnabled         bool
	Endpoint               string
	FieldsMap              map[string]bool
	serverInfo             ServerInfo
	languageInfo           LanguageInfo
	Debug                  bool
}

func Configure(config Configuration) {
	if config.APIKey != "" {
		Config.APIKey = config.APIKey
	}
	if config.ProjectID != "" {
		Config.ProjectID = config.ProjectID
	}
	if config.Endpoint != "" {
		Config.Endpoint = config.Endpoint
		log.Printf("Setting Treblle endpoint to: %s", config.Endpoint)
	}

	// Initialize default masking settings
	Config.MaskingEnabled = true // Enable by default
	if len(config.DefaultFieldsToMask) > 0 {
		Config.DefaultFieldsToMask = config.DefaultFieldsToMask
	} else {
		// Load from environment variable if available
		if envFields := getEnvMaskedFields(); len(envFields) > 0 {
			Config.DefaultFieldsToMask = envFields
		} else {
			Config.DefaultFieldsToMask = getDefaultFieldsToMask()
		}
	}

	if len(config.AdditionalFieldsToMask) > 0 {
		Config.AdditionalFieldsToMask = config.AdditionalFieldsToMask
	}

	Config.FieldsMap = generateFieldsToMask(Config.DefaultFieldsToMask, Config.AdditionalFieldsToMask)
}

// getEnvMaskedFields reads masked fields from environment variable
func getEnvMaskedFields() []string {
	fieldsStr := os.Getenv("TREBLLE_MASKED_FIELDS")
	if fieldsStr == "" {
		return nil
	}
	return strings.Split(fieldsStr, ",")
}

// getDefaultFieldsToMask returns the default list of fields to mask
func getDefaultFieldsToMask() []string {
	return []string{
		"password",
		"pwd",
		"secret",
		"password_confirmation",
		"passwordConfirmation",
		"cc",
		"card_number",
		"cardNumber",
		"ccv",
		"ssn",
		"credit_score",
		"creditScore",
	}
}

func generateFieldsToMask(defaultFields, additionalFields []string) map[string]bool {
	fields := append(defaultFields, additionalFields...)
	fieldsToMask := make(map[string]bool)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			fieldsToMask[field] = true
		}
	}
	return fieldsToMask
}

// shouldMaskField checks if a field should be masked based on configuration
func shouldMaskField(field string) bool {
	_, exists := Config.FieldsMap[strings.ToLower(field)]
	return exists
}
