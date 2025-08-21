# Job Management

The Zowe Go SDK provides comprehensive job management functionality for z/OS job operations. This allows you to submit, monitor, and manage jobs on mainframe systems through ZOSMF REST APIs.

## Overview

The job management system consists of the following key components:

- **Job**: Represents a z/OS job with its properties and status
- **JobInfo**: Detailed information about a job including creation and modification dates
- **SpoolFile**: Represents job output files (DD statements)
- **ZOSMFJobManager**: Manages job operations and provides REST API integration
- **JobManager**: Interface for job management operations

## Features

- ✅ Submit jobs using JCL statements (via PUT)
- ✅ Submit jobs from datasets (via PUT)
- ✅ Submit jobs from local files (via PUT)
- ✅ List jobs with filtering options
- ✅ Get detailed job information
- ✅ Monitor job status
- ✅ Cancel running jobs
- ✅ Delete completed jobs
- ✅ Retrieve spool files and their content
- ✅ Wait for job completion with timeout
- ✅ Job validation and JCL generation
- ✅ Comprehensive error handling

## Implementation Details

### HTTP Methods
- **Job Submission**: Uses PUT method as per IBM z/OSMF documentation
- **Job Cancellation**: Uses PUT method
- **Job Deletion**: Uses DELETE method
- **Job Retrieval**: Uses GET method

### Endpoint Templates
The SDK uses proper endpoint templates for consistent URL construction:
- Jobs collection: `/restjobs/jobs`
- Job by name and ID: `/restjobs/jobs/{jobname}/{jobid}`
- Job by correlator: `/restjobs/jobs/{correlator}`
- Spool files: `/restjobs/jobs/{correlator}/files`
- Spool file content: `/restjobs/jobs/{correlator}/files/{spoolid}/records`

### Parameter Naming
- **correlator**: Used for most job operations (recommended for API consistency)
- **jobName + jobID**: Used specifically for `GetJobByNameID` operations
- The `GetJob` function accepts a correlator parameter for consistency with IBM documentation

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/ojuschugh1/zowe-client-go-sdk/pkg/jobs"
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

### Creating Job Managers

```go
// From a profile manager
pm := profile.NewProfileManager()
jm, err := jobs.CreateJobManager(pm, "default")

// From a profile directly
zosmfProfile := &profile.ZOSMFProfile{...}
jm, err := jobs.NewJobManagerFromProfile(zosmfProfile)

// Direct connection
jm, err := jobs.CreateJobManagerDirect("mainframe.example.com", 443, "user", "pass")

// With additional options (BasePath defaults to /zosmf if omitted)
jm, err := jobs.CreateJobManagerDirectWithOptions("mainframe.example.com", 443, "user", "pass", false, "")
```

## API Reference

### Core Types

```go
// Job represents a z/OS job
type Job struct {
    JobID       string            `json:"jobid"`
    JobName     string            `json:"jobname"`
    Owner       string            `json:"owner"`
    Status      string            `json:"status"`
    Subsystem   string            `json:"subsystem,omitempty"`
    Type        string            `json:"type,omitempty"`
    Class       string            `json:"class,omitempty"`
    PhaseName   string            `json:"phase-name,omitempty"`
    PhaseNumber int               `json:"phase-number,omitempty"`
    RetCode     string            `json:"retcode,omitempty"`
    URL         string            `json:"url,omitempty"`
    FilesURL    string            `json:"files-url,omitempty"`
    JobCorrelator string          `json:"job-correlator,omitempty"`
    ExecutionClass string         `json:"execution-class,omitempty"`
    ExecutionMode string          `json:"execution-mode,omitempty"`
    JobInfo     *JobInfo          `json:"job-info,omitempty"`
    SpoolFiles  []SpoolFile       `json:"spool-files,omitempty"`
}

// JobInfo contains detailed information about a job
type JobInfo struct {
    JobID       string    `json:"jobid"`
    JobName     string    `json:"jobname"`
    Owner       string    `json:"owner"`
    Status      string    `json:"status"`
    // ... additional fields
    CreationDate time.Time `json:"creation-date,omitempty"`
    ModificationDate time.Time `json:"modification-date,omitempty"`
}

// SpoolFile represents a job output file
type SpoolFile struct {
    ID          int    `json:"id"`
    DDName      string `json:"ddname"`
    StepName    string `json:"stepname,omitempty"`
    ProcStep    string `json:"procstep,omitempty"`
    Class       string `json:"class,omitempty"`
    Records     int    `json:"records,omitempty"`
    Bytes       int    `json:"bytes,omitempty"`
    URL         string `json:"url,omitempty"`
    ContentURL  string `json:"content-url,omitempty"`
}

// SubmitJobRequest represents a job submission request
type SubmitJobRequest struct {
    JobDataSet string `json:"jobDataSet,omitempty"`
    JobLocalFile string `json:"jobLocalFile,omitempty"`
    JobStatement string `json:"jobStatement,omitempty"`
    Directory string `json:"directory,omitempty"`
    Extension string `json:"extension,omitempty"`
    Volume string `json:"volume,omitempty"`
}

// JobFilter represents filters for job queries
type JobFilter struct {
    Owner       string `json:"owner,omitempty"`
    Prefix      string `json:"prefix,omitempty"`
    MaxJobs     int    `json:"max-jobs,omitempty"`
    JobID       string `json:"jobid,omitempty"`
    JobName     string `json:"jobname,omitempty"`
    Status      string `json:"status,omitempty"`
    UserCorrelator string `json:"user-correlator,omitempty"`
}
```

### Key Functions

#### Job Manager Creation
- `NewJobManager(session *profile.Session) *ZOSMFJobManager`
- `NewJobManagerFromProfile(profile *profile.ZOSMFProfile) (*ZOSMFJobManager, error)`
- `CreateJobManager(pm *profile.ZOSMFProfileManager, profileName string) (*ZOSMFJobManager, error)`
- `CreateJobManagerDirect(host string, port int, user, password string) (*ZOSMFJobManager, error)`
- `CreateJobManagerDirectWithOptions(host string, port int, user, password string, rejectUnauthorized bool, basePath string) (*ZOSMFJobManager, error)`

#### Job Operations (z/OSMF /restjobs)
- `ListJobs(filter *JobFilter) (*JobList, error)`
- `GetJob(correlator string) (*Job, error)` - Get job by correlator (recommended)
- `GetJobInfo(correlator string) (*JobInfo, error)`
- `GetJobStatus(correlator string) (string, error)`
- `GetJobByNameID(jobName, jobID string) (*Job, error)` - Get job by name and ID
- `GetJobByCorrelator(correlator string) (*Job, error)` - Get job by correlator
- `SubmitJob(request *SubmitJobRequest) (*SubmitJobResponse, error)`
- `CancelJob(correlator string) error`
- `DeleteJob(correlator string) error`
- `PurgeJob(correlator string) error`

#### Spool File Operations
- `GetSpoolFiles(correlator string) ([]SpoolFile, error)`
- `GetSpoolFileContent(correlator string, spoolID int) (string, error)`

#### Convenience Functions
- `SubmitJobStatement(jclStatement string) (*SubmitJobResponse, error)`
- `SubmitJobFromDataset(dataset string, volume string) (*SubmitJobResponse, error)`
- `SubmitJobFromLocalFile(localFile, directory, extension string) (*SubmitJobResponse, error)`
- `WaitForJobCompletion(correlator string, timeout time.Duration, pollInterval time.Duration) (string, error)`
- `GetJobsByOwner(owner string, maxJobs int) (*JobList, error)`
- `GetJobsByPrefix(prefix string, maxJobs int) (*JobList, error)`
- `GetJobsByStatus(status string, maxJobs int) (*JobList, error)`
- `GetJobOutput(correlator string) (map[string]string, error)`
- `GetJobOutputByDDName(correlator, ddName string) (string, error)`

#### JCL Generation
- `CreateSimpleJobStatement(jobName, account, user, msgClass, msgLevel string) string`
- `CreateJobWithStep(jobName, account, user, msgClass, msgLevel, stepName, pgm string, ddStatements []string) string`

#### Validation
- `ValidateJobRequest(request *SubmitJobRequest) error`

## Usage Examples

### Submitting Jobs

```go
// Submit using JCL statement
jcl := "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A\n//STEP1 EXEC PGM=IEFBR14"
response, err := jm.SubmitJobStatement(jcl)

// Submit from dataset
response, err := jm.SubmitJobFromDataset("TEST.JCL", "")

// Submit using request object
request := &jobs.SubmitJobRequest{
    JobStatement: "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A",
}
response, err := jm.SubmitJob(request)
```

### Listing and Filtering Jobs

```go
// List all jobs
jobList, err := jm.ListJobs(nil)

// List jobs with filter
filter := &jobs.JobFilter{
    Owner:   "myuser",
    Prefix:  "TEST",
    MaxJobs: 10,
    Status:  "OUTPUT",
}
jobList, err := jm.ListJobs(filter)

// Convenience methods
jobList, err := jm.GetJobsByOwner("myuser", 10)
jobList, err := jm.GetJobsByPrefix("TEST", 5)
jobList, err := jm.GetJobsByStatus("OUTPUT", 20)
```

### Monitoring Jobs

```go
// Get job status
status, err := jm.GetJobStatus("<correlator>")

// Get detailed job information by correlator
job, err := jm.GetJob("<correlator>")

// Or get by job name and id
job, err := jm.GetJobByNameID("JOBNAME", "JOB001")

// Wait for job completion
status, err := jm.WaitForJobCompletion("JOB001", 5*time.Minute, 10*time.Second)
```

### Working with Spool Files

```go
// Get all spool files for a job
spoolFiles, err := jm.GetSpoolFiles("JOB001")

// Get content of a specific spool file
content, err := jm.GetSpoolFileContent("JOB001", 1)

// Get all job output
output, err := jm.GetJobOutput("JOB001")

// Get output for specific DD name
content, err := jm.GetJobOutputByDDName("JOB001", "SYSOUT")
```

### Job Management

```go
// Cancel a running job
err := jm.CancelJob("JOB001")

// Delete a completed job
err := jm.DeleteJob("JOB001")

// Purge a job (remove from system)
err := jm.PurgeJob("JOB001")

// Close job manager and clean up connections
err := jm.CloseJobManager()
```

### Resource Management

```go
// Always close job managers when done to prevent memory leaks
jm, err := jobs.CreateJobManagerDirect("mainframe.example.com", 443, "user", "pass")
if err != nil {
    log.Fatal(err)
}
defer jm.CloseJobManager()

// Use the job manager
response, err := jm.SubmitJobStatement("//TESTJOB JOB (ACCT),'USER',MSGCLASS=A")
if err != nil {
    log.Fatal(err)
}

// Job manager will be automatically closed when function exits
```

### JCL Generation

```go
// Create simple job statement
jobStatement := jobs.CreateSimpleJobStatement("TESTJOB", "ACCT", "USER", "A", "(1,1)")

// Create complete job with step
ddStatements := []string{
    "//DD1 DD DSN=TEST.DATA,DISP=SHR",
    "//DD2 DD SYSOUT=A",
}
jcl := jobs.CreateJobWithStep("TESTJOB", "ACCT", "USER", "A", "(1,1)", "STEP1", "IEFBR14", ddStatements)
```

### Validation

```go
// Validate job request
request := &jobs.SubmitJobRequest{
    JobStatement: "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A",
}
err := jobs.ValidateJobRequest(request)
if err != nil {
    log.Printf("Invalid job request: %v", err)
}
```

## Job Status Values

The SDK recognizes the following job status values:

- **ACTIVE**: Job is currently running
- **INPUT**: Job is waiting for input
- **OUTPUT**: Job has completed and is in output queue
- **CC 0000**: Job completed with condition code 0000 (success)
- **CC 0001**: Job completed with condition code 0001
- **CC 0002**: Job completed with condition code 0002
- **CC 0003**: Job completed with condition code 0003
- **CC 0004**: Job completed with condition code 0004
- **ABEND**: Job abended

## Error Handling

The SDK provides comprehensive error handling:

```go
// Check for specific error types
response, err := jm.SubmitJob(request)
if err != nil {
    if strings.Contains(err.Error(), "API request failed with status 401") {
        // Authentication error
        log.Println("Authentication failed")
    } else if strings.Contains(err.Error(), "API request failed with status 404") {
        // Job not found
        log.Println("Job not found")
    } else {
        // Other errors
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Best Practices

1. **Always validate job requests** before submission
2. **Use appropriate timeouts** when waiting for job completion
3. **Handle errors gracefully** and provide meaningful error messages
4. **Clean up completed jobs** to avoid accumulation
5. **Use filters** when listing jobs to improve performance
6. **Monitor job status** before accessing spool files
7. **Use convenience methods** for common operations

## Configuration

The job management system uses the same configuration as the profile management system:

- **Zowe CLI Configuration**: Reads from `~/.zowe/zowe.config.json` (Unix/Linux/macOS) or `%USERPROFILE%\.zowe\zowe.config.json` (Windows)
- **Direct Connection**: Can be configured with connection parameters directly
- **TLS Configuration**: Supports secure and insecure connections
- **Timeout Configuration**: Configurable timeouts for operations

## Security Considerations

- **Credentials**: Passwords are handled securely through the profile system
- **TLS**: Supports TLS encryption for secure communication
- **Authentication**: Uses ZOSMF authentication mechanisms
- **Authorization**: Respects z/OS security controls and RACF profiles

## Integration with Profile Management

The job management system integrates seamlessly with the profile management system:

```go
// Create profile manager
pm := profile.NewProfileManager()

// Get profile and create job manager
zosmfProfile, err := pm.GetZOSMFProfile("default")
if err != nil {
    log.Fatal(err)
}

// Create job manager from profile
jm, err := jobs.NewJobManagerFromProfile(zosmfProfile)
if err != nil {
    log.Fatal(err)
}

// Use job manager
response, err := jm.SubmitJobStatement("//TESTJOB JOB (ACCT),'USER',MSGCLASS=A")
```

This integration allows for consistent configuration management and session handling across the SDK.
