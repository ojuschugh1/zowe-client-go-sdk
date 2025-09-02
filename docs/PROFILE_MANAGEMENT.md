# Profile Management

The Zowe Go SDK provides comprehensive profile management functionality that is compatible with the Zowe CLI configuration format. This allows you to manage ZOSMF profiles and create sessions for connecting to mainframe systems.

## Overview

The profile management system consists of the following key components:

- **ZOSMFProfile**: Represents a ZOSMF profile configuration
- **Session**: Represents an active connection to a mainframe
- **ZOSMFProfileManager**: Manages profiles and provides CRUD operations
- **ProfileManager**: Interface for profile management operations

## Features

- ✅ Compatible with Zowe CLI configuration format
- ✅ Read profiles from `~/.zowe/zowe.config.json` (Unix/Linux/macOS) or `%USERPROFILE%\.zowe\zowe.config.json` (Windows)
- ✅ Create profiles programmatically
- ✅ Manage multiple profiles
- ✅ Create sessions from profiles
- ✅ Direct session creation without profiles
- ✅ Profile validation
- ✅ Profile cloning
- ✅ Session header management

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
    // Create a profile manager
    pm := profile.NewProfileManager()
    
    // Get a profile by name
    zosmfProfile, err := pm.GetZOSMFProfile("default")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a session
    session, err := zosmfProfile.NewSession()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Connected to %s as %s\n", session.Host, session.User)
}
```

### Creating Profiles Programmatically

```go
// Create a basic profile
profile := profile.CreateZOSMFProfile("myprofile", "mainframe.example.com", 443, "myuser", "mypassword")

// Create a profile with custom options
customProfile := profile.CreateZOSMFProfileWithOptions(
    "custom",
    "dev-mainframe.example.com",
    8080,
    "devuser",
    "devpass",
    false, // Allow insecure connections
    "/api/v1",
)
```

### Direct Session Creation

```go
// Create a session directly without a profile
session, err := profile.CreateSessionDirect("mainframe.example.com", 443, "myuser", "mypassword")
if err != nil {
    log.Fatal(err)
}

// Create a session with custom options
session, err := profile.CreateSessionDirectWithOptions(
    "mainframe.example.com",
    8080,
    "myuser",
    "mypassword",
    false, // Allow insecure connections
    "/api/v1",
)
```

## API Reference

### ZOSMFProfile

Represents a ZOSMF profile configuration.

```go
type ZOSMFProfile struct {
    Name               string `json:"name"`
    Host               string `json:"host"`
    Port               int    `json:"port"`
    User               string `json:"user"`
    Password           string `json:"password"`
    RejectUnauthorized bool   `json:"rejectUnauthorized"`
    BasePath           string `json:"basePath"`
}
```

#### Methods

- `NewSession() (*Session, error)`: Creates a new session from the profile

### Session

Represents an active connection to a mainframe.

```go
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
```

#### Methods

- `GetBaseURL() string`: Returns the base URL for the session
- `GetHTTPClient() *http.Client`: Returns the HTTP client for the session
- `GetHeaders() map[string]string`: Returns the headers for the session
- `AddHeader(key, value string)`: Adds a header to the session
- `RemoveHeader(key string)`: Removes a header from the session

### ZOSMFProfileManager

Manages ZOSMF profiles and provides CRUD operations.

#### Constructor Functions

- `NewProfileManager() *ZOSMFProfileManager`: Creates a new profile manager using the default config path
- `NewProfileManagerWithPath(configPath string) *ZOSMFProfileManager`: Creates a new profile manager with a custom config path

#### Methods

- `GetZOSMFProfile(name string) (*ZOSMFProfile, error)`: Retrieves a ZOSMF profile by name
- `ListZOSMFProfiles() ([]string, error)`: Returns a list of available ZOSMF profile names
- `SaveZOSMFProfile(profile *ZOSMFProfile) error`: Saves a ZOSMF profile to the configuration
- `DeleteZOSMFProfile(name string) error`: Deletes a ZOSMF profile from the configuration
- `GetDefaultZOSMFProfile() (*ZOSMFProfile, error)`: Returns the default ZOSMF profile
- `CreateSession(profileName string) (*Session, error)`: Creates a session from a profile name

### Convenience Functions

- `CreateZOSMFProfile(name, host string, port int, user, password string) *ZOSMFProfile`: Creates a new ZOSMF profile
- `CreateZOSMFProfileWithOptions(name, host string, port int, user, password string, rejectUnauthorized bool, basePath string) *ZOSMFProfile`: Creates a new ZOSMF profile with additional options
- `CreateSessionDirect(host string, port int, user, password string) (*Session, error)`: Creates a session directly with connection parameters
- `CreateSessionDirectWithOptions(host string, port int, user, password string, rejectUnauthorized bool, basePath string) (*Session, error)`: Creates a session directly with additional options
- `ValidateProfile(profile *ZOSMFProfile) error`: Validates that a ZOSMF profile has all required fields
- `CloneProfile(profile *ZOSMFProfile) *ZOSMFProfile`: Creates a copy of a ZOSMF profile

## Configuration Format

The SDK reads Zowe CLI configuration from JSON files with the following structure:

```json
{
  "profiles": {
    "zosmf": {
      "default": {
        "host": "mainframe.example.com",
        "port": 443,
        "user": "myuser",
        "password": "mypassword",
        "rejectUnauthorized": true,
        "basePath": "/zosmf/api/v1"
      },
      "dev": {
        "host": "dev-mainframe.example.com",
        "port": 8080,
        "user": "devuser",
        "password": "devpass",
        "rejectUnauthorized": false,
        "basePath": "/api/v1"
      }
    }
  },
  "default": {
    "zosmf": "default"
  }
}
```

### Configuration Locations

- **Unix/Linux/macOS**: `~/.zowe/zowe.config.json`
- **Windows**: `%USERPROFILE%\.zowe\zowe.config.json`

## Error Handling

The SDK provides comprehensive error handling with descriptive error messages:

```go
profile, err := pm.GetZOSMFProfile("nonexistent")
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        fmt.Println("Profile does not exist")
    } else if strings.Contains(err.Error(), "failed to load config") {
        fmt.Println("Configuration file not found or invalid")
    }
}
```

## Security Considerations

- Passwords are stored in plain text in the configuration file
- Consider using environment variables or secure credential storage for production use
- The `RejectUnauthorized` flag controls TLS certificate validation
- Default value for `RejectUnauthorized` is `true` for security

## Examples

### Working with Multiple Profiles

```go
pm := profile.NewProfileManager()

// List all available profiles
profiles, err := pm.ListZOSMFProfiles()
if err != nil {
    log.Fatal(err)
}

for _, profileName := range profiles {
    profile, err := pm.GetZOSMFProfile(profileName)
    if err != nil {
        log.Printf("Error loading profile %s: %v", profileName, err)
        continue
    }
    
    session, err := profile.NewSession()
    if err != nil {
        log.Printf("Error creating session for %s: %v", profileName, err)
        continue
    }
    
    fmt.Printf("Profile: %s, URL: %s\n", profileName, session.GetBaseURL())
}
```

### Profile Validation

```go
profile := &profile.ZOSMFProfile{
    Name: "test",
    // Missing required fields
}

err := profile.ValidateProfile(profile)
if err != nil {
    fmt.Printf("Profile validation failed: %v\n", err)
    // Handle validation error
}
```

### Session Header Management

```go
session, err := profile.NewSession()
if err != nil {
    log.Fatal(err)
}

// Add custom headers
session.AddHeader("X-Custom-Header", "custom-value")
session.AddHeader("Authorization", "Bearer token123")

// Remove headers
session.RemoveHeader("X-Custom-Header")

// Get all headers
headers := session.GetHeaders()
for key, value := range headers {
    fmt.Printf("%s: %s\n", key, value)
}
```

## Testing

The SDK includes comprehensive tests for all functionality:

```bash
# Run all tests
go test -v ./pkg/profile/

# Run tests with coverage
go test -v -coverprofile=coverage.out ./pkg/profile/
go tool cover -html=coverage.out -o coverage.html
```

## Best Practices

1. **Use Environment Variables**: For production applications, consider using environment variables for sensitive information
2. **Validate Profiles**: Always validate profiles before using them
3. **Handle Errors**: Implement proper error handling for all profile operations
4. **Secure Storage**: Use secure credential storage for production environments
5. **Connection Pooling**: Consider implementing connection pooling for multiple sessions
6. **Logging**: Implement appropriate logging for debugging and monitoring

## Troubleshooting

### Common Issues

1. **Configuration File Not Found**: Ensure the Zowe CLI configuration file exists in the correct location
2. **Profile Not Found**: Verify the profile name exists in the configuration
3. **Connection Issues**: Check network connectivity and firewall settings
4. **TLS Certificate Issues**: Adjust the `RejectUnauthorized` setting if needed

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
// Set log level for debugging
log.SetLevel(log.DebugLevel)
``` 
