package treblle

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type MetaData struct {
	ApiKey    string   `json:"api_key"`    // Renamed internally but kept the same JSON field name
	ProjectID string   `json:"project_id"` // Renamed internally but kept the same JSON field name
	Version   float64  `json:"version"`
	Sdk       string   `json:"sdk"`
	Data      DataInfo `json:"data"`
}

type DataInfo struct {
	Server   ServerInfo   `json:"server"`
	Language LanguageInfo `json:"language"`
	Request  RequestInfo  `json:"request"`
	Response ResponseInfo `json:"response"`
	Errors   []ErrorInfo  `json:"errors,omitempty"`
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
func GetServerInfo(r *http.Request) ServerInfo {
	// Get local timezone
	_, offset := time.Now().Zone()
	tzInfo := fmt.Sprintf("UTC%+d", offset/3600) // Simplified timezone format like the old SDK

	// Get OS version with timeout
	osVersion := GetOSVersion()

	return ServerInfo{
		Ip:        SelectFirstValidIPv4("127.0.0.1"), // Default to localhost, will be updated with actual IP in middleware
		Timezone:  tzInfo,
		Software:  runtime.Version(),
		Signature: "Treblle Go SDK",
		Protocol:  DetectProtocol(r), // Use the DetectProtocol function to determine the protocol
		Os:        GetOSInfo(osVersion),
	}
}

// GetOSVersion returns the OS version with a timeout
func GetOSVersion() string {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Prepare command based on OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "sw_vers", "-productVersion")
	case "linux":
		cmd = exec.CommandContext(ctx, "uname", "-r")
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "ver")
	default:
		return "unknown"
	}

	// Run command with timeout
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Clean and return output
	return strings.TrimSpace(string(out))
}

// GetOSInfo returns information about the operating system that is running on the server
func GetOSInfo(version string) OsInfo {
	return OsInfo{
		Name:         runtime.GOOS,
		Release:      version,
		Architecture: runtime.GOARCH,
	}
}

func GetLanguageInfo() LanguageInfo {
	return LanguageInfo{
		Name:    "go",
		Version: runtime.Version(),
	}
}

// SelectFirstValidIPv4 ensures only the first valid IPv4 address is returned
// This function takes a comma-separated list of IPs (like those in X-Forwarded-For)
// and returns only the first valid IPv4 address found
func SelectFirstValidIPv4(ipList string) string {
	// If empty, return localhost
	if ipList == "" {
		return "127.0.0.1"
	}

	// Split by comma if multiple IPs are provided
	ips := strings.Split(ipList, ",")

	// Check each IP
	for _, ipRaw := range ips {
		// Clean up the IP (remove spaces, etc.)
		ip := strings.TrimSpace(ipRaw)

		// Parse the IP to validate it
		parsedIP := net.ParseIP(ip)

		// Check if it's a valid IPv4 address
		if parsedIP != nil && parsedIP.To4() != nil {
			return ip
		}
	}

	// If no valid IPv4 found, return the first one anyway or localhost
	if len(ips) > 0 {
		return strings.TrimSpace(ips[0])
	}

	return "127.0.0.1"
}

// DetectProtocol determines the HTTP protocol version from the request
// Returns the protocol string (e.g., "HTTP/1.1", "HTTP/2.0")
func DetectProtocol(r *http.Request) string {
	if r == nil {
		return "HTTP/1.1" // Default to HTTP/1.1 if no request is provided
	}

	// The Proto field contains the protocol version (e.g., "HTTP/1.1", "HTTP/2.0")
	if r.Proto != "" {
		return r.Proto
	}

	// Fallback to ProtoMajor and ProtoMinor if Proto is not available
	if r.ProtoMajor > 0 {
		return fmt.Sprintf("HTTP/%d.%d", r.ProtoMajor, r.ProtoMinor)
	}

	// Default fallback
	return "HTTP/1.1"
}
