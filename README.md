# Zowe Client Go SDK

A Go SDK for the Zowe framework that provides programmatic APIs to perform basic mainframe tasks on z/OS.

## Features

- **Profile Management**: Compatible with Zowe CLI configuration
- **ZOSMF Profile Support**: Read and manage ZOSMF profiles
- **Session Management**: Multiple sessions to the same mainframe with different users
- **Dataset Management**: CRUD operations for z/OS datasets (planned)
- **Job Management**: z/OS job operations (planned)

## Installation

```bash
go get github.com/zowe/zowe-client-go-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/zowe/zowe-client-go-sdk/pkg/profile"
)

func main() {
    // Create a profile manager
    pm := profile.NewProfileManager()
    
    // Load a profile by name
    zosmfProfile, err := pm.GetZOSMFProfile("default")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a session
    session, err := zosmfProfile.CreateSession()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Connected to %s as %s\n", session.Host, session.User)
}
```

## Configuration

The SDK reads Zowe CLI configuration from the standard locations:
- `~/.zowe/zowe.config.json` (Unix/Linux/macOS)
- `%USERPROFILE%\.zowe\zowe.config.json` (Windows)

## License

This project is licensed under the Apache License 2.0. 