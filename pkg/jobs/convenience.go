package jobs

import (
	"fmt"
	"strings"
	"time"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

// parseCorrelator parses "jobname:jobid" into separate parts
func parseCorrelator(correlator string) (jobName, jobID string, err error) {
	parts := strings.Split(correlator, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("correlator must be in format 'jobname:jobid', got: %s", correlator)
	}
	return parts[0], parts[1], nil
}

// CreateJobManager creates a job manager from a profile manager
func CreateJobManager(pm *profile.ZOSMFProfileManager, profileName string) (*ZOSMFJobManager, error) {
	zosmfProfile, err := pm.GetZOSMFProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ZOSMF profile '%s': %w", profileName, err)
	}

	session, err := zosmfProfile.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewJobManager(session), nil
}

// CreateJobManagerDirect creates a job manager with connection details
func CreateJobManagerDirect(host string, port int, user, password string) (*ZOSMFJobManager, error) {
	session, err := profile.CreateSessionDirect(host, port, user, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewJobManager(session), nil
}

// CreateJobManagerDirectWithOptions creates a job manager with extra options
func CreateJobManagerDirectWithOptions(host string, port int, user, password string, rejectUnauthorized bool, basePath string) (*ZOSMFJobManager, error) {
	session, err := profile.CreateSessionDirectWithOptions(host, port, user, password, rejectUnauthorized, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewJobManager(session), nil
}

// SubmitJobStatement submits a job using a JCL statement
func (jm *ZOSMFJobManager) SubmitJobStatement(jclStatement string) (*SubmitJobResponse, error) {
	request := &SubmitJobRequest{
		JobStatement: jclStatement,
	}
	return jm.SubmitJob(request)
}

// SubmitJobFromDataset submits a job from a dataset
func (jm *ZOSMFJobManager) SubmitJobFromDataset(dataset string, volume string) (*SubmitJobResponse, error) {
	// Ensure dataset name is properly formatted for z/OSMF
	// Remove any existing "//" prefix to ensure consistency
	dataset = strings.TrimPrefix(dataset, "//")
	
	// z/OSMF automatically adds the user prefix, so we should use relative dataset names
	// If the dataset name starts with the user ID, remove it to avoid duplication
	// This is a common pattern in z/OSMF APIs
	session := jm.session.(*profile.Session)
	userID := session.User
	if strings.HasPrefix(dataset, userID+".") {
		dataset = strings.TrimPrefix(dataset, userID+".")
	}
	
	request := &SubmitJobRequest{
		JobDataSet: dataset,
		Volume:     volume,
	}
	return jm.SubmitJob(request)
}

// SubmitJobFromLocalFile submits a job from a local file
func (jm *ZOSMFJobManager) SubmitJobFromLocalFile(localFile, directory, extension string) (*SubmitJobResponse, error) {
	request := &SubmitJobRequest{
		JobLocalFile: localFile,
		Directory:    directory,
		Extension:    extension,
	}
	return jm.SubmitJob(request)
}

// WaitForJobCompletion waits for a job to complete and returns the final status
func (jm *ZOSMFJobManager) WaitForJobCompletion(correlator string, timeout time.Duration, pollInterval time.Duration) (string, error) {
	startTime := time.Now()
	
	for {
		// Check if timeout exceeded
		if time.Since(startTime) > timeout {
			return "", fmt.Errorf("timeout waiting for job %s to complete", correlator)
		}

		// Get job status
		status, err := jm.GetJobStatus(correlator)
		if err != nil {
			return "", fmt.Errorf("failed to get job status: %w", err)
		}

		// Check if job is complete
		if isJobComplete(status) {
			return status, nil
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}

// isJobComplete checks if a job status indicates completion
func isJobComplete(status string) bool {
	completedStatuses := []string{"OUTPUT", "CC 0000", "CC 0001", "CC 0002", "CC 0003", "CC 0004", "ABEND"}
	status = strings.ToUpper(status)
	
	for _, completedStatus := range completedStatuses {
		if strings.Contains(status, completedStatus) {
			return true
		}
	}
	return false
}

// GetJobsByOwner retrieves jobs owned by a specific user
func (jm *ZOSMFJobManager) GetJobsByOwner(owner string, maxJobs int) (*JobList, error) {
	filter := &JobFilter{
		Owner:   owner,
		MaxJobs: maxJobs,
	}
	return jm.ListJobs(filter)
}

// GetJobsByPrefix retrieves jobs with a specific name prefix
func (jm *ZOSMFJobManager) GetJobsByPrefix(prefix string, maxJobs int) (*JobList, error) {
	filter := &JobFilter{
		Prefix:  prefix,
		MaxJobs: maxJobs,
	}
	return jm.ListJobs(filter)
}

// GetJobsByStatus retrieves jobs with a specific status
func (jm *ZOSMFJobManager) GetJobsByStatus(status string, maxJobs int) (*JobList, error) {
	filter := &JobFilter{
		Status:  status,
		MaxJobs: maxJobs,
	}
	return jm.ListJobs(filter)
}

// GetJobOutput retrieves the output of a completed job
func (jm *ZOSMFJobManager) GetJobOutput(correlator string) (map[string]string, error) {
	var jobName, jobID string
	var err error
	
	// Check if it's already in correlator format (jobname:jobid)
	if strings.Contains(correlator, ":") {
		// Parse correlator to get jobname and jobid
		jobName, jobID, err = parseCorrelator(correlator)
		if err != nil {
			return nil, fmt.Errorf("invalid correlator format: %w", err)
		}
	} else {
		// If it's just a job ID, we need to find the job first
		jobFilter := &JobFilter{
			JobID: correlator,
			MaxJobs: 100,
		}
		
		jobList, err := jm.ListJobs(jobFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to find job with ID %s: %w", correlator, err)
		}
		
		// Find the job with the specified job ID
		found := false
		for _, job := range jobList.Jobs {
			if job.JobID == correlator {
				jobName = job.JobName
				jobID = job.JobID
				found = true
				break
			}
		}
		
		if !found {
			return nil, fmt.Errorf("job with ID %s not found", correlator)
		}
	}

	// Get spool files
	spoolFiles, err := jm.GetSpoolFiles(jobName, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get spool files: %w", err)
	}

	// Get content for each spool file
	output := make(map[string]string)
	for _, spoolFile := range spoolFiles {
		content, err := jm.GetSpoolFileContent(jobName, jobID, spoolFile.ID)
		if err != nil {
			// Log error but continue with other files
			continue
		}
		output[spoolFile.DDName] = content
	}

	return output, nil
}

// GetJobOutputByDDName retrieves the output of a specific DD name for a job
func (jm *ZOSMFJobManager) GetJobOutputByDDName(correlator, ddName string) (string, error) {
	var jobName, jobID string
	var err error
	
	// Check if it's already in correlator format (jobname:jobid)
	if strings.Contains(correlator, ":") {
		// Parse correlator to get jobname and jobid
		jobName, jobID, err = parseCorrelator(correlator)
		if err != nil {
			return "", fmt.Errorf("invalid correlator format: %w", err)
		}
	} else {
		// If it's just a job ID, we need to find the job first
		jobFilter := &JobFilter{
			JobID: correlator,
			MaxJobs: 100,
		}
		
		jobList, err := jm.ListJobs(jobFilter)
		if err != nil {
			return "", fmt.Errorf("failed to find job with ID %s: %w", correlator, err)
		}
		
		// Find the job with the specified job ID
		found := false
		for _, job := range jobList.Jobs {
			if job.JobID == correlator {
				jobName = job.JobName
				jobID = job.JobID
				found = true
				break
			}
		}
		
		if !found {
			return "", fmt.Errorf("job with ID %s not found", correlator)
		}
	}

	// Get spool files
	spoolFiles, err := jm.GetSpoolFiles(jobName, jobID)
	if err != nil {
		return "", fmt.Errorf("failed to get spool files: %w", err)
	}

	// Find the spool file with the specified DD name
	for _, spoolFile := range spoolFiles {
		if spoolFile.DDName == ddName {
			content, err := jm.GetSpoolFileContent(jobName, jobID, spoolFile.ID)
			if err != nil {
				return "", fmt.Errorf("failed to get content for DD %s: %w", ddName, err)
			}
			return content, nil
		}
	}

	return "", fmt.Errorf("DD name %s not found for job %s", ddName, correlator)
}

// ValidateJobRequest validates a job submission request
func ValidateJobRequest(request *SubmitJobRequest) error {
	if request == nil {
		return fmt.Errorf("job request cannot be nil")
	}

	// Check that at least one job source is specified
	if request.JobStatement == "" && request.JobDataSet == "" && request.JobLocalFile == "" {
		return fmt.Errorf("at least one job source must be specified (jobStatement, jobDataSet, or jobLocalFile)")
	}

	// Validate job statement
	if request.JobStatement != "" {
		if !strings.Contains(strings.ToUpper(request.JobStatement), "JOB") {
			return fmt.Errorf("job statement must contain a JOB card")
		}
	}

	// Validate dataset name
	if request.JobDataSet != "" {
		if !isValidDatasetName(request.JobDataSet) {
			return fmt.Errorf("invalid dataset name: %s", request.JobDataSet)
		}
	}

	return nil
}

// isValidDatasetName validates a z/OS dataset name
func isValidDatasetName(dataset string) bool {
	// Basic validation for z/OS dataset names
	// Dataset names should be 1-44 characters, alphanumeric, @, #, $, -, .
	// Cannot start with a number
	if len(dataset) == 0 || len(dataset) > 44 {
		return false
	}

	// Check first character
	if dataset[0] >= '0' && dataset[0] <= '9' {
		return false
	}

	// Check all characters
	for _, char := range dataset {
		if !isValidDatasetChar(char) {
			return false
		}
	}

	return true
}

// isValidDatasetChar checks if a character is valid in a z/OS dataset name
func isValidDatasetChar(char rune) bool {
	return (char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '@' || char == '#' || char == '$' || char == '-' || char == '.'
}

// CreateSimpleJobStatement creates a simple JCL job statement
func CreateSimpleJobStatement(jobName, account, user, msgClass, msgLevel string) string {
	if jobName == "" {
		jobName = "GOJOB"
	}
	if account == "" {
		account = "ACCT"
	}
	if user == "" {
		user = "USER"
	}
	if msgClass == "" {
		msgClass = "A"
	}
	if msgLevel == "" {
		msgLevel = "(1,1)"
	}

	return fmt.Sprintf("//%s JOB (%s),'%s',MSGCLASS=%s,MSGLEVEL=%s", 
		jobName, account, user, msgClass, msgLevel)
}

// CreateJobWithStep creates a complete JCL job with a step
func CreateJobWithStep(jobName, account, user, msgClass, msgLevel, stepName, pgm string, ddStatements []string) string {
	jobStatement := CreateSimpleJobStatement(jobName, account, user, msgClass, msgLevel)
	
	jcl := jobStatement + "\n"
	
	if stepName == "" {
		stepName = "STEP1"
	}
	
	jcl += fmt.Sprintf("//%s EXEC PGM=%s\n", stepName, pgm)
	
	for _, ddStatement := range ddStatements {
		jcl += ddStatement + "\n"
	}
	
	return jcl
}
