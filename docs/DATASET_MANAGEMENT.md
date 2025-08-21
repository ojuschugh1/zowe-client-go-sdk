# Dataset Management

The Zowe Go SDK provides comprehensive dataset management functionality for z/OS dataset operations. This allows you to create, manage, and manipulate datasets on mainframe systems through ZOSMF REST APIs.

## Overview

The dataset management system consists of the following key components:

- **Dataset**: Represents a z/OS dataset with its properties and metadata
- **DatasetMember**: Represents a member in a partitioned dataset
- **ZOSMFDatasetManager**: Manages dataset operations and provides REST API integration
- **DatasetManager**: Interface for dataset management operations

## Features

- ✅ Create sequential and partitioned datasets
- ✅ Delete datasets and dataset members
- ✅ Upload content to datasets and members
- ✅ Download content from datasets and members
- ✅ List datasets with filtering options
- ✅ List members in partitioned datasets
- ✅ Copy and rename datasets
- ✅ Dataset validation and naming conventions
- ✅ Space allocation management
- ✅ Comprehensive error handling

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
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
    
    // Create a dataset manager
    dm, err := datasets.NewDatasetManagerFromProfile(zosmfProfile)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a sequential dataset
    err = dm.CreateSequentialDataset("TEST.DATA")
    if err != nil {
        log.Fatal(err)
    }
    
    // Upload content
    err = dm.UploadText("TEST.DATA", "Hello, z/OS!")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Dataset created and content uploaded successfully!")
}
```

### Creating Dataset Managers

```go
// From a profile manager
pm := profile.NewProfileManager()
dm, err := datasets.CreateDatasetManager(pm, "default")

// From a profile directly
zosmfProfile := &profile.ZOSMFProfile{...}
dm, err := datasets.NewDatasetManagerFromProfile(zosmfProfile)

// Direct connection
dm, err := datasets.CreateDatasetManagerDirect("mainframe.example.com", 443, "user", "pass")

// With additional options (BasePath defaults to /zosmf if omitted)
dm, err := datasets.CreateDatasetManagerDirectWithOptions("mainframe.example.com", 443, "user", "pass", false, "")
```

## API Reference

### Core Types

```go
// Dataset represents a z/OS dataset
type Dataset struct {
    Name         string      `json:"name"`
    Type         DatasetType `json:"type"`
    Volume       string      `json:"volume,omitempty"`
    Space        Space       `json:"space,omitempty"`
    RecordFormat RecordFormat `json:"recordFormat,omitempty"`
    RecordLength RecordLength `json:"recordLength,omitempty"`
    BlockSize    BlockSize   `json:"blockSize,omitempty"`
    Directory    int         `json:"directory,omitempty"`
    Created      time.Time   `json:"created,omitempty"`
    Modified     time.Time   `json:"modified,omitempty"`
    Size         int64       `json:"size,omitempty"`
    Used         int64       `json:"used,omitempty"`
    Extents      int         `json:"extents,omitempty"`
    Referenced   time.Time   `json:"referenced,omitempty"`
    Expiration   time.Time   `json:"expiration,omitempty"`
    Owner        string      `json:"owner,omitempty"`
    Security     string      `json:"security,omitempty"`
}

// DatasetMember represents a member in a partitioned dataset
type DatasetMember struct {
    Name      string    `json:"name"`
    Size      int64     `json:"size"`
    Created   time.Time `json:"created,omitempty"`
    Modified  time.Time `json:"modified,omitempty"`
    UserID    string    `json:"userid,omitempty"`
    Version   int       `json:"version,omitempty"`
    ModLevel  int       `json:"modLevel,omitempty"`
    ChangeDate time.Time `json:"changeDate,omitempty"`
}

// Space represents space allocation parameters
type Space struct {
    Primary   int       `json:"primary"`
    Secondary int       `json:"secondary"`
    Unit      SpaceUnit `json:"unit"`
    Directory int       `json:"directory,omitempty"`
}
```

### Dataset Types

```go
const (
    DatasetTypeSequential DatasetType = "SEQ"
    DatasetTypePartitioned DatasetType = "PO"
    DatasetTypePDSE       DatasetType = "PDSE"
    DatasetTypeVSAM       DatasetType = "VSAM"
)
```

### Space Units

```go
const (
    SpaceUnitTracks   SpaceUnit = "TRK"
    SpaceUnitCylinders SpaceUnit = "CYL"
    SpaceUnitKB       SpaceUnit = "KB"
    SpaceUnitMB       SpaceUnit = "MB"
    SpaceUnitGB       SpaceUnit = "GB"
)
```

### Record Formats

```go
const (
    RecordFormatFixed    RecordFormat = "F"
    RecordFormatVariable RecordFormat = "V"
    RecordFormatUndefined RecordFormat = "U"
)
```

## Usage Examples

### Creating Datasets

```go
// Create sequential dataset with defaults
err := dm.CreateSequentialDataset("TEST.SEQ")

// Create partitioned dataset with defaults
err := dm.CreatePartitionedDataset("TEST.PDS")

// Create dataset with custom options
space := datasets.Space{
    Primary:   100,
    Secondary: 50,
    Unit:      datasets.SpaceUnitCylinders,
    Directory: 20,
}
err := dm.CreateDatasetWithOptions("TEST.CUSTOM", datasets.DatasetTypeSequential, space,
    datasets.RecordFormatFixed, datasets.RecordLength132, datasets.BlockSize27920)

// Create dataset using request object
request := &datasets.CreateDatasetRequest{
    Name: "TEST.DATA",
    Type: datasets.DatasetTypeSequential,
    Space: datasets.Space{
        Primary:   10,
        Secondary: 5,
        Unit:      datasets.SpaceUnitTracks,
    },
    RecordFormat: datasets.RecordFormatFixed,
    RecordLength: datasets.RecordLength80,
    BlockSize:    datasets.BlockSize800,
}
err := dm.CreateDataset(request)
```

### Uploading Content

```go
// Upload text to sequential dataset
err := dm.UploadText("TEST.SEQ", "Hello, World!")

// Upload text to partitioned dataset member
err := dm.UploadTextToMember("TEST.PDS", "MEMBER1", "//TESTJOB JOB (ACCT),'USER'")

// Upload with custom options
request := &datasets.UploadRequest{
    DatasetName: "TEST.DATA",
    MemberName:  "MEMBER1", // Optional for partitioned datasets
    Content:     "Hello, World!",
    Encoding:    "UTF-8",
    Replace:     true,
}
err := dm.UploadContent(request)
```

### Downloading Content

```go
// Download text from sequential dataset
content, err := dm.DownloadText("TEST.SEQ")

// Download text from partitioned dataset member
content, err := dm.DownloadTextFromMember("TEST.PDS", "MEMBER1")

// Download with custom options
request := &datasets.DownloadRequest{
    DatasetName: "TEST.DATA",
    MemberName:  "MEMBER1", // Optional for partitioned datasets
    Encoding:    "UTF-8",
}
content, err := dm.DownloadContent(request)
```

### Listing and Filtering

```go
// List all datasets
datasetList, err := dm.ListDatasets(nil)

// List datasets with filter
filter := &datasets.DatasetFilter{
    Type:  "SEQ",
    Owner: "USER1",
    Limit: 10,
}
datasetList, err := dm.ListDatasets(filter)

// List members in partitioned dataset
memberList, err := dm.ListMembers("TEST.PDS")

// Get specific dataset information
dataset, err := dm.GetDataset("TEST.DATA")

// Get specific member information
member, err := dm.GetMember("TEST.PDS", "MEMBER1")
```

### Dataset Operations

```go
// Check if dataset exists
exists, err := dm.Exists("TEST.DATA")

// Copy dataset
err := dm.CopyDataset("SOURCE.DATA", "TARGET.DATA")

// Rename dataset
err := dm.RenameDataset("OLD.DATA", "NEW.DATA")

// Delete member
err := dm.DeleteMember("TEST.PDS", "MEMBER1")

// Delete dataset
err := dm.DeleteDataset("TEST.DATA")
```

### Validation

```go
// Validate dataset name
err := datasets.ValidateDatasetName("TEST.DATA")

// Validate member name
err := datasets.ValidateMemberName("MEMBER1")

// Validate create dataset request
request := &datasets.CreateDatasetRequest{...}
err := datasets.ValidateCreateDatasetRequest(request)

// Validate upload request
request := &datasets.UploadRequest{...}
err := datasets.ValidateUploadRequest(request)

// Validate download request
request := &datasets.DownloadRequest{...}
err := datasets.ValidateDownloadRequest(request)
```

### Space Allocation Helpers

```go
// Create default space allocation
space := datasets.CreateDefaultSpace(datasets.SpaceUnitTracks)
// Result: Primary: 10, Secondary: 5, Unit: TRK, Directory: 5

// Create large space allocation
space := datasets.CreateLargeSpace(datasets.SpaceUnitCylinders)
// Result: Primary: 100, Secondary: 50, Unit: CYL, Directory: 20

// Create small space allocation
space := datasets.CreateSmallSpace(datasets.SpaceUnitKB)
// Result: Primary: 5, Secondary: 2, Unit: KB, Directory: 2
```

## Error Handling

The SDK provides comprehensive error handling for all operations:

```go
// Check for specific error types
if err != nil {
    if strings.Contains(err.Error(), "API request failed with status 404") {
        // Dataset not found
        fmt.Println("Dataset does not exist")
    } else if strings.Contains(err.Error(), "API request failed with status 403") {
        // Permission denied
        fmt.Println("Insufficient permissions")
    } else if strings.Contains(err.Error(), "API request failed with status 409") {
        // Conflict (e.g., dataset already exists)
        fmt.Println("Dataset already exists")
    } else {
        // Other errors
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

## Resource Management

Always close dataset managers when done to prevent memory leaks:

```go
// Create dataset manager
dm, err := datasets.CreateDatasetManagerDirect("mainframe.example.com", 443, "user", "pass")
if err != nil {
    log.Fatal(err)
}
defer dm.CloseDatasetManager()

// Use the dataset manager
err = dm.CreateSequentialDataset("TEST.DATA")
if err != nil {
    log.Fatal(err)
}

// Dataset manager will be automatically closed when function exits
```

## Naming Conventions

### Dataset Names
- Maximum 44 characters
- Must start with A-Z, @, #, or $
- Can contain A-Z, 0-9, @, #, $, -, .
- Cannot contain consecutive periods (..)
- Cannot start or end with a period
- Cannot contain consecutive hyphens (--)

### Member Names
- Maximum 8 characters
- Must start with A-Z, @, #, or $
- Can contain A-Z, 0-9, @, #, $, -, .
- Cannot contain consecutive periods (..)
- Cannot start or end with a period

## Integration with Other SDK Components

The dataset management integrates seamlessly with other SDK components:

```go
// Use with profile management
pm := profile.NewProfileManager()
zosmfProfile, err := pm.GetZOSMFProfile("default")
if err != nil {
    log.Fatal(err)
}

// Create dataset manager from profile
dm, err := datasets.NewDatasetManagerFromProfile(zosmfProfile)
if err != nil {
    log.Fatal(err)
}

// Use with job management
jm, err := jobs.NewJobManagerFromProfile(zosmfProfile)
if err != nil {
    log.Fatal(err)
}

// Create dataset and submit job that uses it
err = dm.CreateSequentialDataset("JOB.INPUT")
if err != nil {
    log.Fatal(err)
}

err = dm.UploadText("JOB.INPUT", "//TESTJOB JOB (ACCT),'USER'")
if err != nil {
    log.Fatal(err)
}

response, err := jm.SubmitJobFromDataset("JOB.INPUT", "")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Job submitted: %s (%s)\n", response.JobName, response.JobID)
```

This integration allows for consistent configuration management and session handling across the SDK.
