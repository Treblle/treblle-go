package treblle

import (
	"runtime"
)

type MetaData struct {
	ApiKey    string   `json:"api_key"`
	ProjectID string   `json:"project_id"`
	Version   string   `json:"version"`
	Sdk       string   `json:"sdk"`
	Data      DataInfo `json:"data"`
}

type DataInfo struct {
	Server   ServerInfo   `json:"server"`
	Language LanguageInfo `json:"language"`
	Request  RequestInfo  `json:"request"`
	Response ResponseInfo `json:"response"`
}

type ServerInfo struct {
	Ip        string `json:"ip"`
	Timezone  string `json:"timezone"`
	Software  string `json:"software"`
	Signature string `json:"signature"`
	Protocol  string `json:"protocol"`
	Os        OsInfo `json:"os"`
}

type OsInfo struct {
	Name         string `json:"name"`
	Release      string `json:"release"`
	Architecture string `json:"architecture"`
}

type LanguageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Get information about the server environment
func getServerInfo() ServerInfo {
	return ServerInfo{
		Ip:        "",
		Timezone:  "UTC",
		Software:  "",
		Signature: "",
		Protocol:  "",
		Os:        getOsInfo(),
	}
}

// Get information about the programming language
func getLanguageInfo() LanguageInfo {
	return LanguageInfo{
		Name:    "go",
		Version: runtime.Version(),
	}
}

// Get information about the operating system that is running on the server
func getOsInfo() OsInfo {
	return OsInfo{
		Name:         runtime.GOOS,
		Release:      "",
		Architecture: runtime.GOARCH,
	}
}
