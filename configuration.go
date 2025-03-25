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
	SDK_TOKEN               string
	API_KEY                 string
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
	IgnoredEnvironments     []string
}

func Configure(config Configuration) {
	if config.SDK_TOKEN != "" {
		Config.APIKey = config.SDK_TOKEN
	}
	if config.API_KEY != "" {
		Config.ProjectID = config.API_KEY
	}
	if config.Endpoint != "" {
		Config.Endpoint = config.Endpoint
	}

	// Set debug mode
	Config.Debug = config.Debug

	// Initialize server and language info
	Config.serverInfo = GetServerInfo(nil)
	Config.languageInfo = GetLanguageInfo()

	// Initialize default masking settings
	Config.MaskingEnabled = true

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
		Config.MaxConcurrentProcessing = 10
	}

	Config.AsyncShutdownTimeout = config.AsyncShutdownTimeout
	if Config.AsyncShutdownTimeout <= 0 {
		Config.AsyncShutdownTimeout = 5 * time.Second
	}

	// Initialize batch error collector if enabled
	if config.BatchErrorEnabled {
		if Config.batchErrorCollector != nil {
			Config.batchErrorCollector.Close()
		}
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
		defaultIgnoredEnvs := []string{"dev", "test", "testing"}
		Config.IgnoredEnvironments = getEnvAsSlice("TREBLLE_IGNORED_ENV", defaultIgnoredEnvs)
	}

	Config.FieldsMap = generateFieldsToMask(Config.DefaultFieldsToMask, Config.AdditionalFieldsToMask)
}

func getEnvMaskedFields() []string {
	fieldsStr := os.Getenv("TREBLLE_MASKED_FIELDS")
	if fieldsStr == "" {
		return nil
	}
	return strings.Split(fieldsStr, ",")
}

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

func getEnvOrDefault(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

func GetSDKInfo() map[string]string {
	return map[string]string{
		"SDK Name":    Config.SDKName,
		"SDK Version": strconv.FormatFloat(Config.SDKVersion, 'f', 2, 64),
	}
}

func getEnvAsSlice(envKey string, defaultValues []string) []string {
	if value := os.Getenv(envKey); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValues
}

func IsEnvironmentIgnored() bool {
	currentEnv := os.Getenv("GO_ENV")
	if currentEnv == "" {
		currentEnv = os.Getenv("ENV")
		if currentEnv == "" {
			currentEnv = os.Getenv("ENVIRONMENT")
			if currentEnv == "" {
				currentEnv = os.Getenv("APP_ENV")
			}
		}
	}

	if currentEnv == "" {
		return false
	}

	for _, ignoredEnv := range Config.IgnoredEnvironments {
		if strings.TrimSpace(currentEnv) == strings.TrimSpace(ignoredEnv) {
			return true
		}
	}

	return false
}
