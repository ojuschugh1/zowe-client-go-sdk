# Zowe Client Go SDK

A Go SDK for the Zowe framework that provides programmatic APIs to perform basic mainframe tasks on z/OS.

## Features

- **Profile Management**: Compatible with Zowe CLI configuration
- **ZOSMF Profile Support**: Read and manage ZOSMF profiles
- **Session Management**: Multiple sessions to the same mainframe with different users
- **Job Management**: Complete z/OS job operations (submit, monitor, cancel, delete)
- **Dataset Management**: CRUD operations for z/OS datasets (create, read, update, delete)
- **Content Management**: Upload and download content to/from datasets
- **Member Operations**: Manage members in partitioned datasets
- **Validation**: Comprehensive validation for dataset names and parameters

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
    "github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
)

func main() {
    // Create a profile manager
    pm := profile.NewProfileManager()
    
    // Load a profile by name
    zosmfProfile, err := pm.GetZOSMFProfile("default")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create job and dataset managers
    jm, err := jobs.NewJobManagerFromProfile(zosmfProfile)
    if err != nil {
        log.Fatal(err)
    }
    
    dm, err := datasets.NewDatasetManagerFromProfile(zosmfProfile)
    if err != nil {
        log.Fatal(err)
    }
    
    // Submit a job
    jcl := "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A\n//STEP1 EXEC PGM=IEFBR14"
    response, err := jm.SubmitJobStatement(jcl)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a dataset and upload content
    err = dm.CreateSequentialDataset("TEST.DATA")
    if err != nil {
        log.Fatal(err)
    }
    
    err = dm.UploadText("TEST.DATA", "Hello, z/OS!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Job submitted: %s (%s)\n", response.JobName, response.JobID)
    fmt.Println("Dataset created and content uploaded successfully!")
}
```

## Configuration

The SDK reads Zowe CLI configuration from the standard locations:
- `~/.zowe/zowe.config.json` (Unix/Linux/macOS)
- `%USERPROFILE%\.zowe\zowe.config.json` (Windows)

## License

This project is licensed under the Apache License 2.0. 