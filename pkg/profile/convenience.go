package profile

import (
	"fmt"
)

// CreateZOSMFProfile creates a new ZOSMF profile with the given parameters
func CreateZOSMFProfile(name, host string, port int, user, password string) *ZOSMFProfile {
	return &ZOSMFProfile{
		Name:               name,
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		RejectUnauthorized: true,
		BasePath:           "",
	}
}

// CreateZOSMFProfileWithOptions creates a new ZOSMF profile with additional options
func CreateZOSMFProfileWithOptions(name, host string, port int, user, password string, rejectUnauthorized bool, basePath string) *ZOSMFProfile {
	return &ZOSMFProfile{
		Name:               name,
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		RejectUnauthorized: rejectUnauthorized,
		BasePath:           basePath,
	}
}

// CreateSessionDirect creates a session directly with connection parameters
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
	}
} 