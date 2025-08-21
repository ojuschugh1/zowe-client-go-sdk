package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateZOSMFProfile(t *testing.T) {
	profile := CreateZOSMFProfile("test", "localhost", 443, "user", "pass")
	
	assert.Equal(t, "test", profile.Name)
	assert.Equal(t, "localhost", profile.Host)
	assert.Equal(t, 443, profile.Port)
	assert.Equal(t, "user", profile.User)
	assert.Equal(t, "pass", profile.Password)
	assert.True(t, profile.RejectUnauthorized)
	assert.Equal(t, "", profile.BasePath)
	assert.Equal(t, "https", profile.Protocol)
}

func TestCreateZOSMFProfileWithOptions(t *testing.T) {
	profile := CreateZOSMFProfileWithOptions("test", "localhost", 443, "user", "pass", false, "/api/v1")
	
	assert.Equal(t, "test", profile.Name)
	assert.Equal(t, "localhost", profile.Host)
	assert.Equal(t, 443, profile.Port)
	assert.Equal(t, "user", profile.User)
	assert.Equal(t, "pass", profile.Password)
	assert.False(t, profile.RejectUnauthorized)
	assert.Equal(t, "/api/v1", profile.BasePath)
	assert.Equal(t, "https", profile.Protocol)
}

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile *ZOSMFProfile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: &ZOSMFProfile{
				Host:     "localhost",
				Port:     443,
				User:     "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			profile: &ZOSMFProfile{
				Port:     443,
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing user",
			profile: &ZOSMFProfile{
				Host:     "localhost",
				Port:     443,
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			profile: &ZOSMFProfile{
				Host: "localhost",
				Port: 443,
				User: "user",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			profile: &ZOSMFProfile{
				Host:     "localhost",
				Port:     0,
				User:     "user",
				Password: "pass",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfile(tt.profile)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewSession(t *testing.T) {
	profile := &ZOSMFProfile{
		Host:               "localhost",
		Port:               443,
		User:               "user",
		Password:           "pass",
		RejectUnauthorized: true,
		BasePath:           "/api/v1",
		Protocol:           "https",
	}

	session, err := profile.NewSession()
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, profile, session.Profile)
	assert.Equal(t, "localhost", session.Host)
	assert.Equal(t, 443, session.Port)
	assert.Equal(t, "user", session.User)
	assert.Equal(t, "pass", session.Password)
	assert.Equal(t, "https://localhost/api/v1", session.BaseURL)
	assert.NotNil(t, session.HTTPClient)
	assert.Equal(t, "application/json", session.Headers["Content-Type"])
	assert.Equal(t, "application/json", session.Headers["Accept"])
}

func TestSessionHeaders(t *testing.T) {
	profile := &ZOSMFProfile{
		Host:     "localhost",
		Port:     443,
		User:     "user",
		Password: "pass",
	}

	session, err := profile.NewSession()
	require.NoError(t, err)

	// Test adding header
	session.AddHeader("X-Custom-Header", "custom-value")
	assert.Equal(t, "custom-value", session.Headers["X-Custom-Header"])

	// Test removing header
	session.RemoveHeader("X-Custom-Header")
	_, exists := session.Headers["X-Custom-Header"]
	assert.False(t, exists)
}

func TestProfileManager(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zowe.config.json")

	// Create test config
	testConfig := ZoweConfig{
		Profiles: map[string]ZoweProfile{
			"zosmf": {
				Type: "zosmf",
				Properties: map[string]interface{}{
					"host":               "testhost.com",
					"port":               float64(443),
					"user":               "testuser",
					"password":           "testpass",
					"rejectUnauthorized": true,
					"basePath":           "/api/v1",
					"protocol":           "https",
				},
			},
			"global_base": {
				Type: "base",
				Properties: map[string]interface{}{
					"host":               "basehost.com",
					"port":               float64(8443),
					"user":               "baseuser",
					"password":           "basepass",
					"rejectUnauthorized": false,
				},
				Secure: []string{"user", "password"},
			},
		},
		Defaults: map[string]string{
			"zosmf": "zosmf",
		},
	}

	// Write test config
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	// Create profile manager with test config
	pm := NewProfileManagerWithPath(configPath)

	// Test listing profiles
	profiles, err := pm.ListZOSMFProfiles()
	require.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Contains(t, profiles, "zosmf")

	// Test getting default profile
	defaultProfile, err := pm.GetDefaultZOSMFProfile()
	require.NoError(t, err)
	assert.Equal(t, "zosmf", defaultProfile.Name)
	assert.Equal(t, "testhost.com", defaultProfile.Host)
	assert.Equal(t, 443, defaultProfile.Port)
	assert.Equal(t, "testuser", defaultProfile.User)
	assert.Equal(t, "testpass", defaultProfile.Password)
	assert.True(t, defaultProfile.RejectUnauthorized)
	assert.Equal(t, "/api/v1", defaultProfile.BasePath)
	assert.Equal(t, "https", defaultProfile.Protocol)

	// Test getting specific profile
	zosmfProfile, err := pm.GetZOSMFProfile("zosmf")
	require.NoError(t, err)
	assert.Equal(t, "zosmf", zosmfProfile.Name)
	assert.Equal(t, "testhost.com", zosmfProfile.Host)
	assert.Equal(t, 443, zosmfProfile.Port)
	assert.Equal(t, "testuser", zosmfProfile.User)
	assert.Equal(t, "testpass", zosmfProfile.Password)
	assert.True(t, zosmfProfile.RejectUnauthorized)
	assert.Equal(t, "/api/v1", zosmfProfile.BasePath)
	assert.Equal(t, "https", zosmfProfile.Protocol)

	// Test getting non-existent profile (in new schema, this should fail)
	_, err = pm.GetZOSMFProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test creating session from profile
	session, err := pm.CreateSession("zosmf")
	require.NoError(t, err)
	assert.Equal(t, "https://testhost.com/api/v1", session.BaseURL)
}

func TestCreateSessionDirect(t *testing.T) {
	session, err := CreateSessionDirect("localhost", 443, "user", "pass")
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, "localhost", session.Host)
	assert.Equal(t, 443, session.Port)
	assert.Equal(t, "user", session.User)
	assert.Equal(t, "pass", session.Password)
	assert.Equal(t, "https://localhost/zosmf", session.BaseURL)
}

func TestCreateSessionDirectWithOptions(t *testing.T) {
	session, err := CreateSessionDirectWithOptions("localhost", 8080, "user", "pass", false, "/api/v1")
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, "localhost", session.Host)
	assert.Equal(t, 8080, session.Port)
	assert.Equal(t, "user", session.User)
	assert.Equal(t, "pass", session.Password)
	assert.Equal(t, "http://localhost:8080/api/v1", session.BaseURL)
}

func TestCloneProfile(t *testing.T) {
	original := &ZOSMFProfile{
		Name:               "original",
		Host:               "localhost",
		Port:               443,
		User:               "user",
		Password:           "pass",
		RejectUnauthorized: true,
		BasePath:           "/api/v1",
		Protocol:           "https",
		Encoding:           "IBM-1047",
		ResponseTimeout:    30,
		CertFile:           "/path/to/cert.pem",
		CertKeyFile:        "/path/to/key.pem",
	}

	cloned := CloneProfile(original)
	
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Host, cloned.Host)
	assert.Equal(t, original.Port, cloned.Port)
	assert.Equal(t, original.User, cloned.User)
	assert.Equal(t, original.Password, cloned.Password)
	assert.Equal(t, original.RejectUnauthorized, cloned.RejectUnauthorized)
	assert.Equal(t, original.BasePath, cloned.BasePath)
	assert.Equal(t, original.Protocol, cloned.Protocol)
	assert.Equal(t, original.Encoding, cloned.Encoding)
	assert.Equal(t, original.ResponseTimeout, cloned.ResponseTimeout)
	assert.Equal(t, original.CertFile, cloned.CertFile)
	assert.Equal(t, original.CertKeyFile, cloned.CertKeyFile)
	
	// Ensure it's a different instance
	assert.NotSame(t, original, cloned)
}

// New tests for missing coverage

func TestSaveZOSMFProfile(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zowe.config.json")

	// Create profile manager
	pm := NewProfileManagerWithPath(configPath)

	// Test saving a new profile
	profile := &ZOSMFProfile{
		Name:               "test",
		Host:               "testhost.com",
		Port:               443,
		User:               "testuser",
		Password:           "testpass",
		RejectUnauthorized: true,
		BasePath:           "/api/v1",
		Protocol:           "https",
		Encoding:           "IBM-1047",
		ResponseTimeout:    30,
	}

	err := pm.SaveZOSMFProfile(profile)
	require.NoError(t, err)

	// Verify the profile was saved
	savedProfile, err := pm.GetZOSMFProfile("zosmf")
	require.NoError(t, err)
	assert.Equal(t, "testhost.com", savedProfile.Host)
	assert.Equal(t, 443, savedProfile.Port)
	assert.Equal(t, "testuser", savedProfile.User)
	assert.Equal(t, "testpass", savedProfile.Password)
	assert.Equal(t, "/api/v1", savedProfile.BasePath)
	assert.Equal(t, "https", savedProfile.Protocol)
	assert.Equal(t, "IBM-1047", savedProfile.Encoding)
	assert.Equal(t, 30, savedProfile.ResponseTimeout)
}

func TestDeleteZOSMFProfile(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zowe.config.json")

	// Create profile manager
	pm := NewProfileManagerWithPath(configPath)

	// Test deleting a profile (should return error in new schema)
	err := pm.DeleteZOSMFProfile("zosmf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestGetZoweConfigPath(t *testing.T) {
	// Test that the function returns a valid path
	configPath := getZoweConfigPath()
	assert.NotEmpty(t, configPath)
	assert.Contains(t, configPath, ".zowe")
	assert.Contains(t, configPath, "zowe.config.json")
}

func TestCreateSessionFromProfile(t *testing.T) {
	profile := &ZOSMFProfile{
		Host:               "localhost",
		Port:               443,
		User:               "user",
		Password:           "pass",
		RejectUnauthorized: true,
		BasePath:           "/api/v1",
		Protocol:           "https",
	}

	session, err := CreateSessionFromProfile(profile)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, "https://localhost/api/v1", session.BaseURL)
	assert.Equal(t, profile, session.Profile)
}

func TestLoadConfigError(t *testing.T) {
	// Test loading config from non-existent file
	pm := NewProfileManagerWithPath("/non/existent/path/config.json")
	
	_, err := pm.GetZOSMFProfile("zosmf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestGetDefaultZOSMFProfileError(t *testing.T) {
	// Create a temporary config file without defaults
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zowe.config.json")

	testConfig := ZoweConfig{
		Profiles: map[string]ZoweProfile{
			"zosmf": {
				Type: "zosmf",
				Properties: map[string]interface{}{
					"host": "testhost.com",
					"port": float64(443),
				},
			},
		},
		// No defaults section
	}

	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	pm := NewProfileManagerWithPath(configPath)
	
	_, err = pm.GetDefaultZOSMFProfile()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no default zosmf profile set")
}

func TestSessionWithDifferentProtocols(t *testing.T) {
	tests := []struct {
		name     string
		profile  *ZOSMFProfile
		expected string
	}{
		{
			name: "https default",
			profile: &ZOSMFProfile{
				Host: "localhost",
				Port: 443,
			},
			expected: "https://localhost/zosmf",
		},
		{
			name: "http port 80",
			profile: &ZOSMFProfile{
				Host: "localhost",
				Port: 80,
			},
			expected: "http://localhost/zosmf",
		},
		{
			name: "http port 8080",
			profile: &ZOSMFProfile{
				Host: "localhost",
				Port: 8080,
			},
			expected: "http://localhost:8080/zosmf",
		},
		{
			name: "custom port",
			profile: &ZOSMFProfile{
				Host: "localhost",
				Port: 8443,
			},
			expected: "https://localhost:8443/zosmf",
		},
		{
			name: "explicit protocol",
			profile: &ZOSMFProfile{
				Host:     "localhost",
				Port:     443,
				Protocol: "http",
			},
			expected: "http://localhost/zosmf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := tt.profile.NewSession()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, session.BaseURL)
		})
	}
}

func TestWriteTestConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test.json")
	content := `{"test": "data"}`

	err := WriteTestConfig(configPath, content)
	require.NoError(t, err)

	// Verify file was written
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
} 