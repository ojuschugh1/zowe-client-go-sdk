package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

// z/OSMF job API endpoints
const (
	// Main jobs endpoint
	JobsEndpoint = "/restjobs/jobs"

	// Job by name and ID
	JobByNameIDEndpoint     = "/restjobs/jobs/%s/%s" // jobname/jobid
	JobByCorrelatorEndpoint = "/restjobs/jobs/%s"    // correlator

	// Job operations
	FilesEndpoint   = "/files"
	CancelEndpoint  = "/cancel"
	PurgeEndpoint   = "/purge"
	RecordsEndpoint = "/records"
	
	// File operations
	JobFilesEndpoint        = "/files"
	JobFilesByIDEndpoint    = "/files/%s/records"
	JobFilesJCLEndpoint     = "/files/JCL/records"
)

// NewJobManager creates a job manager with the given session
func NewJobManager(session *profile.Session) *ZOSMFJobManager {
	return &ZOSMFJobManager{
		session: session,
	}
}

// NewJobManagerFromProfile creates a job manager from a profile
func NewJobManagerFromProfile(profile *profile.ZOSMFProfile) (*ZOSMFJobManager, error) {
	session, err := profile.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return NewJobManager(session), nil
}

// ListJobs gets jobs matching the filter
func (jm *ZOSMFJobManager) ListJobs(filter *JobFilter) (*JobList, error) {
	session := jm.session.(*profile.Session)
	
	// Build query parameters
	params := url.Values{}
	if filter != nil {
		if filter.Owner != "" {
			params.Set("owner", filter.Owner)
		}
		if filter.Prefix != "" {
			params.Set("prefix", filter.Prefix)
		}
		if filter.MaxJobs > 0 {
			params.Set("max-jobs", strconv.Itoa(filter.MaxJobs))
		}
		if filter.JobID != "" {
			params.Set("jobid", filter.JobID)
		}
		if filter.JobName != "" {
			params.Set("jobname", filter.JobName)
		}
		if filter.Status != "" {
			params.Set("status", filter.Status)
		}
		if filter.UserCorrelator != "" {
			params.Set("user-correlator", filter.UserCorrelator)
		}
	}

	// Build URL
	apiURL := session.GetBaseURL() + JobsEndpoint
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response with fallback for array responses
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	// First try object with jobs field
	var jobList JobList
	if err := json.Unmarshal(bodyBytes, &jobList); err == nil && (len(jobList.Jobs) > 0 || string(bodyBytes) == "{}") {
		return &jobList, nil
	}
	// Fallback: direct array response
	var jobsArr []Job
	if err := json.Unmarshal(bodyBytes, &jobsArr); err == nil {
		return &JobList{Jobs: jobsArr}, nil
	}
	return nil, fmt.Errorf("failed to decode response: %s", string(bodyBytes))
}

// GetJob retrieves detailed information about a specific job by correlator or job ID
func (jm *ZOSMFJobManager) GetJob(correlator string) (*Job, error) {
	// Check if it's already in correlator format (jobname:jobid)
	if strings.Contains(correlator, ":") {
		// Parse correlator to get jobname and jobid
		jobName, jobID, err := parseCorrelator(correlator)
		if err != nil {
			return nil, fmt.Errorf("invalid correlator format: %w", err)
		}
		return jm.GetJobByNameID(jobName, jobID)
	}
	
	// If it's just a job ID, we need to find the job first
	// List jobs and find the one with this job ID
	jobFilter := &JobFilter{
		JobID: correlator,
		MaxJobs: 100, // Get more jobs to find the one we need
	}
	
	jobList, err := jm.ListJobs(jobFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to find job with ID %s: %w", correlator, err)
	}
	
	// Find the job with the specified job ID
	for _, job := range jobList.Jobs {
		if job.JobID == correlator {
			return jm.GetJobByNameID(job.JobName, job.JobID)
		}
	}
	
	return nil, fmt.Errorf("job with ID %s not found", correlator)
}

// GetJobInfo retrieves job information
func (jm *ZOSMFJobManager) GetJobInfo(correlator string) (*JobInfo, error) {
	// Parse correlator to get jobname and jobid
	jobName, jobID, err := parseCorrelator(correlator)
	if err != nil {
		return nil, fmt.Errorf("invalid correlator format: %w", err)
	}
	
	session := jm.session.(*profile.Session)
	
	// Build URL using jobname/jobid format
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByNameIDEndpoint, url.PathEscape(jobName), url.PathEscape(jobID)) + JobFilesEndpoint

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var jobInfo JobInfo
	if err := json.NewDecoder(resp.Body).Decode(&jobInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &jobInfo, nil
}

// GetJobStatus retrieves the status of a job
func (jm *ZOSMFJobManager) GetJobStatus(correlator string) (string, error) {
	job, err := jm.GetJob(correlator)
	if err != nil {
		return "", err
	}
	return job.Status, nil
}

// GetJobByNameID retrieves a job by job name and job id
func (jm *ZOSMFJobManager) GetJobByNameID(jobName, jobID string) (*Job, error) {
	session := jm.session.(*profile.Session)
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByNameIDEndpoint, url.PathEscape(jobName), url.PathEscape(jobID))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &job, nil
}

// GetJobByCorrelator retrieves a job by correlator
func (jm *ZOSMFJobManager) GetJobByCorrelator(correlator string) (*Job, error) {
	session := jm.session.(*profile.Session)
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(correlator))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &job, nil
}

// SubmitJob submits a new job
func (jm *ZOSMFJobManager) SubmitJob(request *SubmitJobRequest) (*SubmitJobResponse, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + JobsEndpoint

	// Prepare request body and content type based on submission type
	var requestBody []byte
	var contentType string
	var err error
	
	if request.JobStatement != "" {
		// Submit job statement as plain text (z/OSMF expects JCL as text/plain for direct submission)
		requestBody = []byte(request.JobStatement)
		contentType = "text/plain"
	} else if request.JobDataSet != "" {
		// Submit job from dataset using JSON format
		// z/OSMF expects the dataset name to be prefixed with "//" for absolute path
		datasetPath := request.JobDataSet
		if !strings.HasPrefix(datasetPath, "//") {
			datasetPath = "//" + datasetPath
		}
		
		body := map[string]interface{}{
			"file": datasetPath,
		}
		if request.Volume != "" {
			body["volume"] = request.Volume
		}
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dataset job request: %w", err)
		}
		contentType = "application/json"
	} else if request.JobLocalFile != "" {
		// Submit job from local file using JSON format
		body := map[string]interface{}{
			"file": request.JobLocalFile,
		}
		if request.Directory != "" {
			body["directory"] = request.Directory
		}
		if request.Extension != "" {
			body["extension"] = request.Extension
		}
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal local file job request: %w", err)
		}
		contentType = "application/json"
	} else {
		return nil, fmt.Errorf("no job source specified (jobStatement, jobDataSet, or jobLocalFile)")
	}

	// Create request (use PUT per z/OSMF documentation)
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", contentType)

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var submitResponse SubmitJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&submitResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &submitResponse, nil
}

// CancelJob cancels a running job
func (jm *ZOSMFJobManager) CancelJob(correlator string) error {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(correlator)) + CancelEndpoint

	// Create request
	req, err := http.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteJob deletes a job using correlator format (jobname:jobid)
func (jm *ZOSMFJobManager) DeleteJob(correlator string) error {
	// Parse correlator to get jobname and jobid
	jobName, jobID, err := parseCorrelator(correlator)
	if err != nil {
		return fmt.Errorf("invalid correlator format: %w", err)
	}
	
	return jm.DeleteJobByNameID(jobName, jobID)
}

// DeleteJobByNameID deletes a job using separate jobName and jobID
func (jm *ZOSMFJobManager) DeleteJobByNameID(jobName, jobID string) error {
	session := jm.session.(*profile.Session)
	
	// Build URL using jobName and jobID format
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByNameIDEndpoint, url.PathEscape(jobName), url.PathEscape(jobID))

	// Create request
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetSpoolFiles retrieves spool files for a job using jobname and jobid
func (jm *ZOSMFJobManager) GetSpoolFiles(jobName, jobID string) ([]SpoolFile, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL using the correct z/OSMF format: /restjobs/jobs/{jobname}/{jobid}/files
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByNameIDEndpoint, url.PathEscape(jobName), url.PathEscape(jobID)) + JobFilesEndpoint

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var spoolFiles []SpoolFile
	if err := json.NewDecoder(resp.Body).Decode(&spoolFiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return spoolFiles, nil
}

// GetSpoolFileContent retrieves the content of a specific spool file
func (jm *ZOSMFJobManager) GetSpoolFileContent(jobName, jobID string, spoolID int) (string, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL using the correct z/OSMF format: /restjobs/jobs/{jobname}/{jobid}/files/{id}/records
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByNameIDEndpoint, url.PathEscape(jobName), url.PathEscape(jobID)) + fmt.Sprintf(JobFilesByIDEndpoint, strconv.Itoa(spoolID))

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// GetSpoolFilesByCorrelator retrieves spool files for a job using correlator format (jobname:jobid)
// This is a convenience method that maintains backward compatibility
func (jm *ZOSMFJobManager) GetSpoolFilesByCorrelator(correlator string) ([]SpoolFile, error) {
	jobName, jobID, err := parseCorrelator(correlator)
	if err != nil {
		return nil, fmt.Errorf("invalid correlator format: %w", err)
	}
	return jm.GetSpoolFiles(jobName, jobID)
}

// GetSpoolFileContentByCorrelator retrieves the content of a specific spool file using correlator format
// This is a convenience method that maintains backward compatibility
func (jm *ZOSMFJobManager) GetSpoolFileContentByCorrelator(correlator string, spoolID int) (string, error) {
	jobName, jobID, err := parseCorrelator(correlator)
	if err != nil {
		return "", fmt.Errorf("invalid correlator format: %w", err)
	}
	return jm.GetSpoolFileContent(jobName, jobID, spoolID)
}



// PurgeJob purges a job (removes it from the system)
func (jm *ZOSMFJobManager) PurgeJob(correlator string) error {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(correlator)) + PurgeEndpoint

	// Create request
	req, err := http.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CloseJobManager closes the job manager and its underlying HTTP connections
func (jm *ZOSMFJobManager) CloseJobManager() error {
	session := jm.session.(*profile.Session)
	
	// Close idle connections in the HTTP client
	if client := session.GetHTTPClient(); client != nil {
		client.CloseIdleConnections()
	}
	
	return nil
}
