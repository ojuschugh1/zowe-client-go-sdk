package profile

import (
	"fmt"
	"os"
)

// CreateZOSMFProfile creates a ZOSMF profile with the given parameters
func CreateZOSMFProfile(name, host string, port int, user, password string) *ZOSMFProfile {
	return &ZOSMFProfile{
		Name:               name,
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		RejectUnauthorized: true,
		BasePath:           "",
		Protocol:           "https",
	}
}

// CreateZOSMFProfileWithOptions creates a ZOSMF profile with extra options
func CreateZOSMFProfileWithOptions(name, host string, port int, user, password string, rejectUnauthorized bool, basePath string) *ZOSMFProfile {
	return &ZOSMFProfile{
		Name:               name,
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		RejectUnauthorized: rejectUnauthorized,
		BasePath:           basePath,
		Protocol:           "https",
	}
}

// CreateSessionDirect creates a session with connection details
func CreateSessionDirect(host string, port int, user, password string) (*Session, error) {
	profile := &ZOSMFProfile{
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		RejectUnauthorized: true,
		BasePath:           "",
	}
	
	return profile.NewSession()
}

// CreateSessionDirectWithOptions creates a session directly with additional options
func CreateSessionDirectWithOptions(host string, port int, user, password string, rejectUnauthorized bool, basePath string) (*Session, error) {
	profile := &ZOSMFProfile{
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		RejectUnauthorized: rejectUnauthorized,
		BasePath:           basePath,
	}
	
	return profile.NewSession()
}

// ValidateProfile validates that a ZOSMF profile has all required fields
func ValidateProfile(profile *ZOSMFProfile) error {
	if profile.Host == "" {
		return fmt.Errorf("host is required")
	}
	if profile.User == "" {
		return fmt.Errorf("user is required")
	}
	if profile.Password == "" {
		return fmt.Errorf("password is required")
	}
	if profile.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	return nil
}

// CloneProfile creates a copy of a ZOSMF profile
func CloneProfile(profile *ZOSMFProfile) *ZOSMFProfile {
	return &ZOSMFProfile{
		Name:               profile.Name,
		Host:               profile.Host,
		Port:               profile.Port,
		User:               profile.User,
		Password:           profile.Password,
		RejectUnauthorized: profile.RejectUnauthorized,
		BasePath:           profile.BasePath,
		Protocol:           profile.Protocol,
		Encoding:           profile.Encoding,
		ResponseTimeout:    profile.ResponseTimeout,
		CertFile:           profile.CertFile,
		CertKeyFile:        profile.CertKeyFile,
	}
}

// WriteTestConfig writes a test configuration to a file
func WriteTestConfig(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
} 