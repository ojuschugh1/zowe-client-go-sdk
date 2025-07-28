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
		Profiles: map[string]map[string]interface{}{
			"zosmf": {
				"default": map[string]interface{}{
					"host":               "testhost.com",
					"port":               float64(443),
					"user":               "testuser",
					"password":           "testpass",
					"rejectUnauthorized": true,
					"basePath":           "/api/v1",
				},
				"dev": map[string]interface{}{
					"host":               "devhost.com",
					"port":               float64(8080),
					"user":               "devuser",
					"password":           "devpass",
					"rejectUnauthorized": false,
					"basePath":           "",
				},
			},
		},
		Default: map[string]string{
			"zosmf": "default",
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
	assert.Len(t, profiles, 2)
	assert.Contains(t, profiles, "default")
	assert.Contains(t, profiles, "dev")

	// Test getting default profile
	defaultProfile, err := pm.GetDefaultZOSMFProfile()
	require.NoError(t, err)
	assert.Equal(t, "default", defaultProfile.Name)
	assert.Equal(t, "testhost.com", defaultProfile.Host)
	assert.Equal(t, 443, defaultProfile.Port)
	assert.Equal(t, "testuser", defaultProfile.User)
	assert.Equal(t, "testpass", defaultProfile.Password)
	assert.True(t, defaultProfile.RejectUnauthorized)
	assert.Equal(t, "/api/v1", defaultProfile.BasePath)

	// Test getting specific profile
	devProfile, err := pm.GetZOSMFProfile("dev")
	require.NoError(t, err)
	assert.Equal(t, "dev", devProfile.Name)
	assert.Equal(t, "devhost.com", devProfile.Host)
	assert.Equal(t, 8080, devProfile.Port)
	assert.Equal(t, "devuser", devProfile.User)
	assert.Equal(t, "devpass", devProfile.Password)
	assert.False(t, devProfile.RejectUnauthorized)
	assert.Equal(t, "", devProfile.BasePath)

	// Test getting non-existent profile
	_, err = pm.GetZOSMFProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test creating session from profile
	session, err := pm.CreateSession("default")
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
	assert.Equal(t, "https://localhost", session.BaseURL)
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
	}

	cloned := CloneProfile(original)
	
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Host, cloned.Host)
	assert.Equal(t, original.Port, cloned.Port)
	assert.Equal(t, original.User, cloned.User)
	assert.Equal(t, original.Password, cloned.Password)
	assert.Equal(t, original.RejectUnauthorized, cloned.RejectUnauthorized)
	assert.Equal(t, original.BasePath, cloned.BasePath)
	
	// Ensure it's a different instance
	assert.NotSame(t, original, cloned)
} 