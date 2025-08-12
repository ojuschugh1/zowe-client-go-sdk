# Zowe Client Go SDK

A Go SDK for the Zowe framework that provides programmatic APIs to perform basic mainframe tasks on z/OS.

## Features

- **Profile Management**: Compatible with Zowe CLI configuration
- **ZOSMF Profile Support**: Read and manage ZOSMF profiles
- **Session Management**: Multiple sessions to the same mainframe with different users
- **Job Management**: Complete z/OS job operations (submit, monitor, cancel, delete)
- **Dataset Management**: CRUD operations for z/OS datasets (planned)

## Installation

```bash
go get github.com/ojuschugh1/zowe-client-go-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
    "github.com/ojuschugh1/zowe-client-go-sdk/pkg/jobs"
)

func main() {
    // Create a profile manager
    pm := profile.NewProfileManager()
    
    // Load a profile by name
    zosmfProfile, err := pm.GetZOSMFProfile("default")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a job manager
    jm, err := jobs.NewJobManagerFromProfile(zosmfProfile)
    if err != nil {
        log.Fatal(err)
    }
    
    // Submit a simple job
    jcl := "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A\n//STEP1 EXEC PGM=IEFBR14"
    response, err := jm.SubmitJobStatement(jcl)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Job submitted: %s (%s)\n", response.JobName, response.JobID)
}
```

## Configuration

The SDK reads Zowe CLI configuration from the standard locations:
- `~/.zowe/zowe.config.json` (Unix/Linux/macOS)
- `%USERPROFILE%\.zowe\zowe.config.json` (Windows)

## License

This project is licensed under the Apache License 2.0. 