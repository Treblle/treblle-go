package treblle

var Config internalConfiguration

// Configuration sets up and customizes communication with the Treblle API
type Configuration struct {
	APIKey     string
	ProjectID  string
	KeysToMask []string
}

// internalConfiguration is used for communication with Treblle API and contains an optimization for
type internalConfiguration struct {
	Configuration
	KeysMap map[string]interface{}
}

func Configure(config Configuration) {
	if config.APIKey != "" {
		Config.APIKey = config.APIKey
	}
	if config.ProjectID != "" {
		Config.ProjectID = config.ProjectID
	}
	if len(config.KeysToMask) > 0 {
		Config.KeysToMask = config.KeysToMask

		// transform the string slice to a map for faster retrieval
		Config.KeysMap = make(map[string]interface{})
		for _, v := range config.KeysToMask {
			Config.KeysMap[v] = nil
		}
	}
}
