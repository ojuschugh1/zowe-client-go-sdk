package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ZoweConfig represents the structure of Zowe CLI configuration
type ZoweConfig struct {
	Profiles map[string]map[string]interface{} `json:"profiles"`
	Default  map[string]string                 `json:"default"`
}

// NewProfileManager creates a new profile manager instance
func NewProfileManager() *ZOSMFProfileManager {
	configPath := getZoweConfigPath()
	return &ZOSMFProfileManager{
		configPath: configPath,
	}
}

// NewProfileManagerWithPath creates a new profile manager with a custom config path
func NewProfileManagerWithPath(configPath string) *ZOSMFProfileManager {
	return &ZOSMFProfileManager{
		configPath: configPath,
	}
}

// GetZOSMFProfile retrieves a ZOSMF profile by name
func (pm *ZOSMFProfileManager) GetZOSMFProfile(name string) (*ZOSMFProfile, error) {
	config, err := pm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Check if the profile exists
	profiles, exists := config.Profiles["zosmf"]
	if !exists {
		return nil, fmt.Errorf("no zosmf profiles found in configuration")
	}

	profileData, exists := profiles[name]
	if !exists {
		return nil, fmt.Errorf("zosmf profile '%s' not found", name)
	}

	// Convert the profile data to ZOSMFProfile
	profile := &ZOSMFProfile{Name: name}
	
	// Type assert profileData to map[string]interface{}
	if profileMap, ok := profileData.(map[string]interface{}); ok {
		if host, ok := profileMap["host"].(string); ok {
			profile.Host = host
		}
		
		if port, ok := profileMap["port"].(float64); ok {
			profile.Port = int(port)
		}
		
		if user, ok := profileMap["user"].(string); ok {
			profile.User = user
		}
		
		if password, ok := profileMap["password"].(string); ok {
			profile.Password = password
		}
		
		if rejectUnauthorized, ok := profileMap["rejectUnauthorized"].(bool); ok {
			profile.RejectUnauthorized = rejectUnauthorized
		} else {
			// Default to true for security
			profile.RejectUnauthorized = true
		}
		
		if basePath, ok := profileMap["basePath"].(string); ok {
			profile.BasePath = basePath
		}
	}

	return profile, nil
}

// ListZOSMFProfiles returns a list of available ZOSMF profile names
func (pm *ZOSMFProfileManager) ListZOSMFProfiles() ([]string, error) {
	config, err := pm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	profiles, exists := config.Profiles["zosmf"]
	if !exists {
		return []string{}, nil
	}

	var profileNames []string
	for name := range profiles {
		profileNames = append(profileNames, name)
	}

	return profileNames, nil
}

// SaveZOSMFProfile saves a ZOSMF profile to the configuration
func (pm *ZOSMFProfileManager) SaveZOSMFProfile(profile *ZOSMFProfile) error {
	config, err := pm.loadConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		config = &ZoweConfig{
			Profiles: make(map[string]map[string]interface{}),
			Default:  make(map[string]string),
		}
	}

	// Ensure zosmf profiles section exists
	if config.Profiles["zosmf"] == nil {
		config.Profiles["zosmf"] = make(map[string]interface{})
	}

	// Convert profile to map
	profileData := map[string]interface{}{
		"host":               profile.Host,
		"port":               profile.Port,
		"user":               profile.User,
		"password":           profile.Password,
		"rejectUnauthorized": profile.RejectUnauthorized,
		"basePath":           profile.BasePath,
	}

	config.Profiles["zosmf"][profile.Name] = profileData

	return pm.saveConfig(config)
}

// DeleteZOSMFProfile deletes a ZOSMF profile from the configuration
func (pm *ZOSMFProfileManager) DeleteZOSMFProfile(name string) error {
	config, err := pm.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profiles, exists := config.Profiles["zosmf"]
	if !exists {
		return fmt.Errorf("no zosmf profiles found")
	}

	if _, exists := profiles[name]; !exists {
		return fmt.Errorf("zosmf profile '%s' not found", name)
	}

	delete(profiles, name)
	return pm.saveConfig(config)
}

// GetDefaultZOSMFProfile returns the default ZOSMF profile
func (pm *ZOSMFProfileManager) GetDefaultZOSMFProfile() (*ZOSMFProfile, error) {
	config, err := pm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	defaultName, exists := config.Default["zosmf"]
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