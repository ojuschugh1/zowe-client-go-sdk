package profile

import (
	"net/http"
)



// ZoweConfig represents the complete Zowe CLI configuration structure
type ZoweConfig struct {
	Schema    string                 `json:"$schema,omitempty"`
	Profiles  map[string]ZoweProfile `json:"profiles"`
	Defaults  map[string]string      `json:"defaults"`
	AutoStore bool                   `json:"autoStore,omitempty"`
}

// ZoweProfile represents a profile in the Zowe CLI configuration
type ZoweProfile struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Secure     []string               `json:"secure,omitempty"`
	Profiles   map[string]ZoweProfile `json:"profiles,omitempty"`
}

// ZOSMFProfile represents a ZOSMF profile configuration
type ZOSMFProfile struct {
	Name               string `json:"name"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	User               string `json:"user"`
	Password           string `json:"password"`
	RejectUnauthorized bool   `json:"rejectUnauthorized"`
	BasePath           string `json:"basePath"`
	Protocol           string `json:"protocol"`
	Encoding           string `json:"encoding,omitempty"`
	ResponseTimeout    int    `json:"responseTimeout,omitempty"`
	CertFile           string `json:"certFile,omitempty"`
	CertKeyFile        string `json:"certKeyFile,omitempty"`
}

// BaseProfile represents the global base profile properties
type BaseProfile struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	User               string `json:"user"`
	Password           string `json:"password"`
	RejectUnauthorized bool   `json:"rejectUnauthorized"`
	TokenType          string `json:"tokenType,omitempty"`
	TokenValue         string `json:"tokenValue,omitempty"`
	CertFile           string `json:"certFile,omitempty"`
	CertKeyFile        string `json:"certKeyFile,omitempty"`
}

// Session represents a connection to a specific mainframe
type Session struct {
	Profile    *ZOSMFProfile
	Host       string
	Port       int
	User       string
	Password   string
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
}

// ProfileManager interface for managing profiles
type ProfileManager interface {
	GetZOSMFProfile(name string) (*ZOSMFProfile, error)
	ListZOSMFProfiles() ([]string, error)
	SaveZOSMFProfile(profile *ZOSMFProfile) error
	DeleteZOSMFProfile(name string) error
}

// ZOSMFProfileManager implements ProfileManager for ZOSMF profiles
type ZOSMFProfileManager struct {
	configPath string
} 