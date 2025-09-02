package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// NewProfileManager creates a profile manager instance
func NewProfileManager() *ZOSMFProfileManager {
	configPath := getZoweConfigPath()
	return &ZOSMFProfileManager{
		configPath: configPath,
	}
}

// NewProfileManagerWithPath creates a profile manager with custom config path
func NewProfileManagerWithPath(configPath string) *ZOSMFProfileManager {
	return &ZOSMFProfileManager{
		configPath: configPath,
	}
}

// GetZOSMFProfile gets a ZOSMF profile by name
func (pm *ZOSMFProfileManager) GetZOSMFProfile(name string) (*ZOSMFProfile, error) {
	config, err := pm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Look for zosmf profile
	zosmfProfile, exists := config.Profiles["zosmf"]
	if !exists {
		return nil, fmt.Errorf("no zosmf profiles found in configuration")
	}

	// Get base profile for inheritance
	var baseProfile *BaseProfile
	if baseProfileData, exists := config.Profiles["global_base"]; exists {
		baseProfile = pm.parseBaseProfile(baseProfileData)
	}

	// Parse the ZOSMF profile
	profile := pm.parseZOSMFProfile(name, zosmfProfile, baseProfile)
	if profile == nil {
		return nil, fmt.Errorf("zosmf profile '%s' not found", name)
	}

	return profile, nil
}

// parseBaseProfile parses the base profile from configuration
func (pm *ZOSMFProfileManager) parseBaseProfile(baseProfileData ZoweProfile) *BaseProfile {
	baseProfile := &BaseProfile{
		RejectUnauthorized: true, // Default to true for security
	}

	properties := baseProfileData.Properties
	if properties != nil {
		if host, ok := properties["host"].(string); ok {
			baseProfile.Host = host
		}
		if port, ok := properties["port"].(float64); ok {
			baseProfile.Port = int(port)
		}
		if user, ok := properties["user"].(string); ok {
			baseProfile.User = user
		}
		if password, ok := properties["password"].(string); ok {
			baseProfile.Password = password
		}
		if rejectUnauthorized, ok := properties["rejectUnauthorized"].(bool); ok {
			baseProfile.RejectUnauthorized = rejectUnauthorized
		}
	}

	return baseProfile
}

// parseZOSMFProfile parses a ZOSMF profile from configuration
func (pm *ZOSMFProfileManager) parseZOSMFProfile(name string, zosmfProfile ZoweProfile, baseProfile *BaseProfile) *ZOSMFProfile {
	// Check if this is the profile we're looking for
	if zosmfProfile.Type != "zosmf" {
		return nil
	}

	// In the new schema, we only have one ZOSMF profile, so we only return it if the name matches
	// For now, we'll return the profile if the name is "zosmf" or if it's the default
	if name != "zosmf" && name != "default" {
		return nil
	}

	profile := &ZOSMFProfile{
		Name:               name,
		RejectUnauthorized: true, // Default to true for security
		Protocol:           "https", // Default protocol
	}

	// Apply base profile properties if available
	if baseProfile != nil {
		if baseProfile.Host != "" {
			profile.Host = baseProfile.Host
		}
		if baseProfile.Port != 0 {
			profile.Port = baseProfile.Port
		}
		if baseProfile.User != "" {
			profile.User = baseProfile.User
		}
		if baseProfile.Password != "" {
			profile.Password = baseProfile.Password
		}
		profile.RejectUnauthorized = baseProfile.RejectUnauthorized
	}

	// Apply ZOSMF profile properties (override base profile)
	properties := zosmfProfile.Properties
	if properties != nil {
		if host, ok := properties["host"].(string); ok {
			profile.Host = host
		}
		if port, ok := properties["port"].(float64); ok {
			profile.Port = int(port)
		}
		if user, ok := properties["user"].(string); ok {
			profile.User = user
		}
		if password, ok := properties["password"].(string); ok {
			profile.Password = password
		}
		if rejectUnauthorized, ok := properties["rejectUnauthorized"].(bool); ok {
			profile.RejectUnauthorized = rejectUnauthorized
		}
		if basePath, ok := properties["basePath"].(string); ok {
			profile.BasePath = basePath
		}
		if protocol, ok := properties["protocol"].(string); ok {
			profile.Protocol = protocol
		}
		if encoding, ok := properties["encoding"].(string); ok {
			profile.Encoding = encoding
		}
		if responseTimeout, ok := properties["responseTimeout"].(float64); ok {
			profile.ResponseTimeout = int(responseTimeout)
		}
		if certFile, ok := properties["certFile"].(string); ok {
			profile.CertFile = certFile
		}
		if certKeyFile, ok := properties["certKeyFile"].(string); ok {
			profile.CertKeyFile = certKeyFile
		}
	}

	return profile
}

// ListZOSMFProfiles returns a list of available ZOSMF profile names
func (pm *ZOSMFProfileManager) ListZOSMFProfiles() ([]string, error) {
	config, err := pm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	zosmfProfile, exists := config.Profiles["zosmf"]
	if !exists {
		return []string{}, nil
	}

	// For the new schema, we need to check if there are sub-profiles
	// or if this is a single ZOSMF profile
	if zosmfProfile.Type == "zosmf" {
		// Single ZOSMF profile, return the default name
		return []string{"zosmf"}, nil
	}

	// Check for sub-profiles
	if zosmfProfile.Profiles != nil {
		var profileNames []string
		for name := range zosmfProfile.Profiles {
			profileNames = append(profileNames, name)
		}
		return profileNames, nil
	}

	return []string{}, nil
}

// SaveZOSMFProfile saves a ZOSMF profile to the configuration
func (pm *ZOSMFProfileManager) SaveZOSMFProfile(profile *ZOSMFProfile) error {
	config, err := pm.loadConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		config = &ZoweConfig{
			Profiles: make(map[string]ZoweProfile),
			Defaults: make(map[string]string),
		}
	}

	// Ensure zosmf profile exists
	if _, exists := config.Profiles["zosmf"]; !exists {
		config.Profiles["zosmf"] = ZoweProfile{
			Type:       "zosmf",
			Properties: make(map[string]interface{}),
		}
	}

	// Convert profile to properties
	properties := map[string]interface{}{
		"host":               profile.Host,
		"port":               profile.Port,
		"user":               profile.User,
		"password":           profile.Password,
		"rejectUnauthorized": profile.RejectUnauthorized,
		"basePath":           profile.BasePath,
		"protocol":           profile.Protocol,
	}

	if profile.Encoding != "" {
		properties["encoding"] = profile.Encoding
	}
	if profile.ResponseTimeout != 0 {
		properties["responseTimeout"] = profile.ResponseTimeout
	}
	if profile.CertFile != "" {
		properties["certFile"] = profile.CertFile
	}
	if profile.CertKeyFile != "" {
		properties["certKeyFile"] = profile.CertKeyFile
	}

	// Update the zosmf profile
	zosmfProfile := config.Profiles["zosmf"]
	zosmfProfile.Properties = properties
	config.Profiles["zosmf"] = zosmfProfile

	return pm.saveConfig(config)
}

// DeleteZOSMFProfile deletes a ZOSMF profile from the configuration
func (pm *ZOSMFProfileManager) DeleteZOSMFProfile(name string) error {
	config, err := pm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	_, exists := config.Profiles["zosmf"]
	if !exists {
		return fmt.Errorf("no zosmf profiles found")
	}

	// For the new schema, we can't delete individual profiles easily
	// since they're properties of the zosmf profile
	// This would require more complex logic to handle sub-profiles
	return fmt.Errorf("deleting individual profiles not supported in new schema")
}

// GetDefaultZOSMFProfile returns the default ZOSMF profile
func (pm *ZOSMFProfileManager) GetDefaultZOSMFProfile() (*ZOSMFProfile, error) {
	config, err := pm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	defaultName, exists := config.Defaults["zosmf"]
	if !exists {
		return nil, fmt.Errorf("no default zosmf profile set")
	}

	return pm.GetZOSMFProfile(defaultName)
}

// loadConfig loads the Zowe configuration from file
func (pm *ZOSMFProfileManager) loadConfig() (*ZoweConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(pm.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("zowe config file not found at %s", pm.configPath)
	}

	// Read the config file directly as JSON
	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ZoweConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// saveConfig saves the Zowe configuration to file
func (pm *ZOSMFProfileManager) saveConfig(config *ZoweConfig) error {
	// Ensure the directory exists
	configDir := filepath.Dir(pm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(pm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getZoweConfigPath returns the path to the Zowe configuration file
func getZoweConfigPath() string {
	var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}

	return filepath.Join(homeDir, ".zowe", "zowe.config.json")
}

// CreateSession creates a session from a profile name
func (pm *ZOSMFProfileManager) CreateSession(profileName string) (*Session, error) {
	profile, err := pm.GetZOSMFProfile(profileName)
	if err != nil {
		return nil, err
	}

	return profile.NewSession()
}

// CreateSessionFromProfile creates a session directly from a profile
func CreateSessionFromProfile(profile *ZOSMFProfile) (*Session, error) {
	return profile.NewSession()
} 