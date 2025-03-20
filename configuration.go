package treblle

import (
	"os"
	"strconv"
	"strings"
	"time"
)

var Config internalConfiguration

// Configuration sets up and customizes communication with the Treblle API
type Configuration struct {
	APIKey                  string
	ProjectID               string
	AdditionalFieldsToMask  []string
	DefaultFieldsToMask     []string
	MaskingEnabled          bool
	Endpoint                string        // Custom endpoint for testing
	BatchErrorEnabled       bool          // Enable batch error collection
	BatchErrorSize          int           // Size of error batch before sending
	BatchFlushInterval      time.Duration // Interval to flush errors if batch size not reached
	SDKName                 string        // Defaults to "go"
	SDKVersion              float64       // Defaults to 2.0
	AsyncProcessingEnabled  bool          // Enable asynchronous request processing
	MaxConcurrentProcessing int           // Maximum number of concurrent async operations (default: 10)
	AsyncShutdownTimeout    time.Duration // Timeout for async shutdown (default: 5s)
	IgnoredEnvironments     []string      // Environments where Treblle does not track requests
	Debug                   bool          // Enable debug mode to see what's being sent to Treblle
}

// internalConfiguration is used for communication with Treblle API and contains optimizations
type internalConfiguration struct {
	APIKey                  string
	ProjectID               string
	AdditionalFieldsToMask  []string
	DefaultFieldsToMask     []string
	MaskingEnabled          bool
	Endpoint                string
	FieldsMap               map[string]bool
	serverInfo              ServerInfo
	languageInfo            LanguageInfo
	Debug                   bool
	batchErrorCollector     *BatchErrorCollector
	SDKName                 string
	SDKVersion              float64
	AsyncProcessingEnabled  bool
	MaxConcurrentProcessing int
	AsyncShutdownTimeout    time.Duration
	IgnoredEnvironments     []string // Environments where Treblle does not track requests
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
	}

	// Set debug mode
	Config.Debug = config.Debug

	// Initialize server and language info
	Config.serverInfo = GetServerInfo(nil) // Pass nil as request, protocol will be updated in middleware
	Config.languageInfo = GetLanguageInfo()

	// Initialize default masking settings
	Config.MaskingEnabled = true // Enable by default

	// Set SDK Name and Version (Can be overridden via ENV)
	sdkName := "go"
	if config.SDKName != "" {
		sdkName = config.SDKName
	}

	sdkVersion := 2.0
	if config.SDKVersion != 0 {
		sdkVersion = config.SDKVersion
	}

	sdkVersionEnv, err := strconv.ParseFloat(os.Getenv("TREBLLE_SDK_VERSION"), 64)
	if err == nil {
		sdkVersion = sdkVersionEnv
	}

	Config.SDKName = getEnvOrDefault("TREBLLE_SDK_NAME", sdkName)
	Config.SDKVersion = sdkVersion

	// Configure async processing
	Config.AsyncProcessingEnabled = config.AsyncProcessingEnabled
	Config.MaxConcurrentProcessing = config.MaxConcurrentProcessing
	if Config.MaxConcurrentProcessing <= 0 {
		Config.MaxConcurrentProcessing = 10 // Default to 10 concurrent operations
	}

	Config.AsyncShutdownTimeout = config.AsyncShutdownTimeout
	if Config.AsyncShutdownTimeout <= 0 {
		Config.AsyncShutdownTimeout = 5 * time.Second // Default to 5 seconds
	}

	// Initialize batch error collector if enabled
	if config.BatchErrorEnabled {
		// Close existing collector if any
		if Config.batchErrorCollector != nil {
			Config.batchErrorCollector.Close()
		}
		// Create new batch error collector
		Config.batchErrorCollector = NewBatchErrorCollector(config.BatchErrorSize, config.BatchFlushInterval)
	}

	// Load default fields to mask if not specified
	if len(config.DefaultFieldsToMask) == 0 {
		Config.DefaultFieldsToMask = getDefaultFieldsToMask()
	} else {
		Config.DefaultFieldsToMask = config.DefaultFieldsToMask
	}

	// Check for additional fields to mask from environment variables
	envMaskedFields := getEnvMaskedFields()
	if len(envMaskedFields) > 0 {
		Config.AdditionalFieldsToMask = append(Config.AdditionalFieldsToMask, envMaskedFields...)
	} else if len(config.AdditionalFieldsToMask) > 0 {
		Config.AdditionalFieldsToMask = config.AdditionalFieldsToMask
	}

	// Load ignored environments from config or environment variable
	if len(config.IgnoredEnvironments) > 0 {
		Config.IgnoredEnvironments = config.IgnoredEnvironments
	} else {
		// Default ignored environments: dev, test, testing
		defaultIgnoredEnvs := []string{"dev", "test", "testing"}
		Config.IgnoredEnvironments = getEnvAsSlice("TREBLLE_IGNORED_ENV", defaultIgnoredEnvs)
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
		"api_key",
		"apiKey",
		"credit_card",
		"creditCard",
		"authorization",
		"authorizationHeader",
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

// Utility function to get env variable or return default
func getEnvOrDefault(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

// GetSDKInfo returns SDK name and version (for debugging)
func GetSDKInfo() map[string]string {
	return map[string]string{
		"SDK Name":    Config.SDKName,
		"SDK Version": strconv.FormatFloat(Config.SDKVersion, 'f', 2, 64),
	}
}

// getEnvAsSlice reads a comma-separated environment variable and returns it as a slice
// If the environment variable is not set, it returns the default values
func getEnvAsSlice(envKey string, defaultValues []string) []string {
	if value := os.Getenv(envKey); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValues
}

// IsEnvironmentIgnored checks if the current environment should be ignored
func IsEnvironmentIgnored() bool {
	// Get the current environment
	currentEnv := os.Getenv("GO_ENV") // Default Go environment variable
	if currentEnv == "" {
		// Try alternative environment variables if GO_ENV is not set
		currentEnv = os.Getenv("ENV")
		if currentEnv == "" {
			currentEnv = os.Getenv("ENVIRONMENT")
			if currentEnv == "" {
				currentEnv = os.Getenv("APP_ENV")
			}
		}
	}

	// If no environment is set, don't ignore
	if currentEnv == "" {
		return false
	}

	// Check if the environment is in the ignored list
	for _, ignoredEnv := range Config.IgnoredEnvironments {
		if strings.TrimSpace(currentEnv) == strings.TrimSpace(ignoredEnv) {
			return true
		}
	}

	return false
}
