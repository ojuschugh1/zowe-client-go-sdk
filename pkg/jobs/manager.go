package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

// API endpoint constants and templates based on z/OSMF docs
const (
	// Collection endpoints
	JobsEndpoint = "/restjobs/jobs"

	// Resource templates
	JobByNameIDEndpoint     = "/restjobs/jobs/%s/%s" // jobname/jobid
	JobByCorrelatorEndpoint = "/restjobs/jobs/%s"    // correlator

	// Sub-resources
	FilesEndpoint   = "/files"
	CancelEndpoint  = "/cancel"
	PurgeEndpoint   = "/purge"
	RecordsEndpoint = "/records"
)

// NewJobManager creates a new job manager using a session
func NewJobManager(session *profile.Session) *ZOSMFJobManager {
	return &ZOSMFJobManager{
		session: session,
	}
}

// NewJobManagerFromProfile creates a new job manager from a profile
func NewJobManagerFromProfile(profile *profile.ZOSMFProfile) (*ZOSMFJobManager, error) {
	session, err := profile.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return NewJobManager(session), nil
}

// ListJobs retrieves a list of jobs based on the provided filter
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

// GetJob retrieves detailed information about a specific job
func (jm *ZOSMFJobManager) GetJob(jobID string) (*Job, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL (treat provided id as correlator for back-compat)
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID))

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
	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &job, nil
}

// GetJobInfo retrieves job information
func (jm *ZOSMFJobManager) GetJobInfo(jobID string) (*JobInfo, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID)) + FilesEndpoint

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
func (jm *ZOSMFJobManager) GetJobStatus(jobID string) (string, error) {
	job, err := jm.GetJob(jobID)
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

	// Prepare request body
	var requestBody interface{}
	if request.JobStatement != "" {
		// Submit job statement
		requestBody = map[string]string{
			"jobStatement": request.JobStatement,
		}
	} else if request.JobDataSet != "" {
		// Submit job from dataset
		requestBody = map[string]string{
			"jobDataSet": request.JobDataSet,
		}
		if request.Volume != "" {
			requestBody.(map[string]string)["volume"] = request.Volume
		}
	} else if request.JobLocalFile != "" {
		// Submit job from local file
		requestBody = map[string]string{
			"jobLocalFile": request.JobLocalFile,
		}
		if request.Directory != "" {
			requestBody.(map[string]string)["directory"] = request.Directory
		}
		if request.Extension != "" {
			requestBody.(map[string]string)["extension"] = request.Extension
		}
	} else {
		return nil, fmt.Errorf("no job source specified (jobStatement, jobDataSet, or jobLocalFile)")
	}

	// Serialize request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request (use PUT per z/OSMF documentation)
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(jsonBody))
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
func (jm *ZOSMFJobManager) CancelJob(jobID string) error {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID)) + CancelEndpoint

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

// DeleteJob deletes a job
func (jm *ZOSMFJobManager) DeleteJob(jobID string) error {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID))

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

// GetSpoolFiles retrieves spool files for a job
func (jm *ZOSMFJobManager) GetSpoolFiles(jobID string) ([]SpoolFile, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID)) + FilesEndpoint

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
func (jm *ZOSMFJobManager) GetSpoolFileContent(jobID string, spoolID int) (string, error) {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID)) + "/files/" + strconv.Itoa(spoolID) + RecordsEndpoint

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

// PurgeJob purges a job (removes it from the system)
func (jm *ZOSMFJobManager) PurgeJob(jobID string) error {
	session := jm.session.(*profile.Session)
	
	// Build URL
	apiURL := session.GetBaseURL() + fmt.Sprintf(JobByCorrelatorEndpoint, url.PathEscape(jobID)) + PurgeEndpoint

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
