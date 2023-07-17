package treblle

var Config internalConfiguration

// Configuration sets up and customizes communication with the Treblle API
type Configuration struct {
	APIKey       string
	ProjectID    string
	FieldsToMask []string
}

// internalConfiguration is used for communication with Treblle API and contains optimizations
type internalConfiguration struct {
	Configuration
	FieldsMap    map[string]interface{}
	serverInfo   ServerInfo
	languageInfo LanguageInfo
}

func Configure(config Configuration) {
	if config.APIKey != "" {
		Config.APIKey = config.APIKey
	}
	if config.ProjectID != "" {
		Config.ProjectID = config.ProjectID
	}
	if len(config.FieldsToMask) > 0 {
		Config.FieldsToMask = config.FieldsToMask

		// transform the string slice to a map for faster retrieval
		Config.FieldsMap = make(map[string]interface{})
		for _, v := range config.FieldsToMask {
			Config.FieldsMap[v] = nil
		}
	}

	Config.serverInfo = getServerInfo()
	Config.languageInfo = getLanguageInfo()
}
