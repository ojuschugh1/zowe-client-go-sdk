package profile

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// ZOSMFProfile represents a ZOSMF profile configuration
type ZOSMFProfile struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	RejectUnauthorized bool `json:"rejectUnauthorized"`
	BasePath string `json:"basePath"`
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

// NewSession creates a new session from a ZOSMF profile
func (p *ZOSMFProfile) NewSession() (*Session, error) {
	// Create HTTP client with TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !p.RejectUnauthorized,
	}
	
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	
	// Build base URL
	protocol := "https"
	if p.Port == 80 {
		protocol = "http"
	} else if p.Port == 8080 {
		protocol = "http"
	}
	
	baseURL := protocol + "://" + p.Host
	if p.Port != 0 && p.Port != 80 && p.Port != 443 {
		baseURL += ":" + fmt.Sprintf("%d", p.Port)
	}
	
	if p.BasePath != "" {
		baseURL += p.BasePath
	}
	
	// Set default headers
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	
	return &Session{
		Profile:    p,
		Host:       p.Host,
		Port:       p.Port,
		User:       p.User,
		Password:   p.Password,
		BaseURL:    baseURL,
		HTTPClient: client,
		Headers:    headers,
	}, nil
}

// GetBaseURL returns the base URL for the session
func (s *Session) GetBaseURL() string {
	return s.BaseURL
}

// GetHTTPClient returns the HTTP client for the session
func (s *Session) GetHTTPClient() *http.Client {
	return s.HTTPClient
}

// GetHeaders returns the headers for the session
func (s *Session) GetHeaders() map[string]string {
	return s.Headers
}

// AddHeader adds a header to the session
func (s *Session) AddHeader(key, value string) {
	s.Headers[key] = value
}

// RemoveHeader removes a header from the session
func (s *Session) RemoveHeader(key string) {
	delete(s.Headers, key)
} 