package datasets

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

// API endpoint constants and templates aligned to z/OSMF dataset APIs
const (
	// Collection endpoints
	DatasetsEndpoint = "/restfiles/ds"
	
	// Resource templates
	DatasetByNameEndpoint = "/restfiles/ds/%s"
	
	// Sub-resources
	MembersEndpoint  = "/member"  // Fixed: was "/members", should be "/member" per z/OSMF API
	ContentEndpoint  = "/content"
	
	// Member-specific endpoints
	MemberByNameEndpoint = "/member/%s"  // Fixed: was "/members/%s", should be "/member/%s" per z/OSMF API
	
	// Content endpoints
	DatasetContentEndpoint = "/content"
	MemberContentEndpoint  = "/content/%s"
)

// NewDatasetManager creates a new dataset manager using a session
func NewDatasetManager(session *profile.Session) *ZOSMFDatasetManager {
	return &ZOSMFDatasetManager{
		session: session,
	}
}

// NewDatasetManagerFromProfile creates a new dataset manager from a profile
func NewDatasetManagerFromProfile(profile *profile.ZOSMFProfile) (*ZOSMFDatasetManager, error) {
	session, err := profile.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return NewDatasetManager(session), nil
}

// ListDatasets retrieves a list of datasets based on the provided filter
func (dm *ZOSMFDatasetManager) ListDatasets(filter *DatasetFilter) (*DatasetList, error) {
	session := dm.session.(*profile.Session)
	
	// Build query parameters according to z/OSMF API documentation
	params := url.Values{}
	
	// z/OSMF requires either dslevel or volser parameter - provide default if none specified
	hasRequiredParam := false
	
	if filter != nil {
		if filter.Name != "" {
			// Use dslevel parameter for dataset name pattern (supports wildcards)
			params.Set("dslevel", filter.Name)
			hasRequiredParam = true
		}
		if filter.Volume != "" {
			// Use volser parameter for volume serial
			params.Set("volser", filter.Volume)
			hasRequiredParam = true
		}
		if filter.Owner != "" {
			// Use start parameter for pagination (dataset name to start from)
			params.Set("start", filter.Owner)
		}
		// Note: Limit is handled via X-IBM-Max-Items header, not query parameter
		// Note: Type/dsorg filtering is not supported in z/OSMF list datasets API
	}
	
	// If no required parameter (dslevel or volser) is provided, use a user-specific pattern
	if !hasRequiredParam {
		// Use the user ID from the session as default pattern to avoid overly broad searches
		params.Set("dslevel", session.User+".*") // List datasets starting with user ID
	}

	// Build URL
	apiURL := session.GetBaseURL() + DatasetsEndpoint
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
	
	// Set X-IBM-Max-Items header for limiting results (instead of query parameter)
	if filter != nil && filter.Limit > 0 {
		req.Header.Set("X-IBM-Max-Items", strconv.Itoa(filter.Limit))
	} else {
		// Set to 0 to return all items by default
		req.Header.Set("X-IBM-Max-Items", "0")
	}
	
	// Set X-IBM-Attributes header to specify what attributes to return
	req.Header.Set("X-IBM-Attributes", "base")

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
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var datasetList DatasetList
	if err := json.Unmarshal(bodyBytes, &datasetList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &datasetList, nil
}

// GetDataset retrieves detailed information about a specific dataset
// Note: The individual dataset API returns binary content, not JSON metadata
// Use ListDatasets with a specific filter to get dataset metadata instead
func (dm *ZOSMFDatasetManager) GetDataset(name string) (*Dataset, error) {
	// Use the list API to get dataset metadata
	dl, err := dm.ListDatasets(&DatasetFilter{Name: name})
	if err != nil {
		return nil, err
	}
	
	// Find the specific dataset in the results
	for _, ds := range dl.Datasets {
		if ds.Name == name {
			return &ds, nil
		}
	}
	
	return nil, fmt.Errorf("dataset not found: %s", name)
}

// GetDatasetInfo retrieves detailed information about a specific dataset
// This method first tries the direct API call, and if that fails, falls back to the list API approach
func (dm *ZOSMFDatasetManager) GetDatasetInfo(name string) (*Dataset, error) {
	// First try the direct API approach
	dataset, err := dm.getDatasetInfoDirect(name)
	if err == nil {
		return dataset, nil
	}
	
	// If direct API fails, fall back to the existing GetDataset method
	// which uses the list API with a filter
	return dm.GetDataset(name)
}

// getDatasetInfoDirect attempts to retrieve dataset info using direct API call
// This is a private method that may not work in all z/OSMF environments
func (dm *ZOSMFDatasetManager) getDatasetInfoDirect(name string) (*Dataset, error) {
	session := dm.session.(*profile.Session)
	
	// Build URL for direct dataset info retrieval
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(name))
	
	// Add query parameter to request metadata instead of content
	params := url.Values{}
	params.Set("metadata", "true")
	apiURL += "?" + params.Encode()

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("dataset not found: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Try to parse response body as JSON
	var dataset Dataset
	if err := json.NewDecoder(resp.Body).Decode(&dataset); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dataset, nil
}

// CreateDataset creates a new dataset using the correct z/OSMF REST API format
// Based on IBM documentation: POST /zosmf/restfiles/ds/<data-set-name>
func (dm *ZOSMFDatasetManager) CreateDataset(request *CreateDatasetRequest) error {
	session := dm.session.(*profile.Session)
	
	// Build URL using the correct format from IBM documentation
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(request.Name))

	// Prepare request body
	requestBody := map[string]interface{}{
		"dsname": request.Name,
		"dsorg":  string(request.Type),
	}

	// Add optional parameters
	if request.Volume != "" {
		requestBody["vol"] = request.Volume
	}
	if request.Space.Primary > 0 {
		requestBody["alcunit"] = string(request.Space.Unit)
		requestBody["primary"] = request.Space.Primary
		requestBody["secondary"] = request.Space.Secondary
		if request.Space.Directory > 0 {
			requestBody["dirblk"] = request.Space.Directory
		}
	}
	if request.RecordFormat != "" {
		requestBody["recfm"] = string(request.RecordFormat)
	}
	if request.RecordLength > 0 {
		requestBody["lrecl"] = int(request.RecordLength)
	}
	if request.BlockSize > 0 {
		requestBody["blksize"] = int(request.BlockSize)
	}
	if request.Directory > 0 {
		requestBody["dirblk"] = request.Directory
	}

	// Serialize request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteDataset deletes a dataset
func (dm *ZOSMFDatasetManager) DeleteDataset(name string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL using template
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(name))

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

// UploadContent uploads content to a dataset
func (dm *ZOSMFDatasetManager) UploadContent(request *UploadRequest) error {
	session := dm.session.(*profile.Session)
	
	// Build URL using correct z/OSMF format
	var apiURL string
	if request.MemberName != "" {
		// For members, use dataset(member) format
		apiURL = session.GetBaseURL() + fmt.Sprintf("/restfiles/ds/%s(%s)", url.PathEscape(request.DatasetName), url.PathEscape(request.MemberName))
	} else {
		// For datasets, use the dataset endpoint directly (no /content suffix)
		apiURL = session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(request.DatasetName))
	}

	var req *http.Request
	var err error

	if request.MemberName != "" {
		// For members, use PUT with plain text content
		req, err = http.NewRequest("PUT", apiURL, bytes.NewBufferString(request.Content))
	} else {
		// For datasets, use PUT with plain text content (per z/OSMF API specification)
		req, err = http.NewRequest("PUT", apiURL, bytes.NewBufferString(request.Content))
	}
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	
	// For both datasets and members, use plain text content type (per z/OSMF API specification)
	req.Header.Set("Content-Type", "text/plain")

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadContent downloads content from a dataset
func (dm *ZOSMFDatasetManager) DownloadContent(request *DownloadRequest) (string, error) {
	session := dm.session.(*profile.Session)
	
	// Build URL using correct z/OSMF format
	var apiURL string
	if request.MemberName != "" {
		// For members, use dataset(member) format
		apiURL = session.GetBaseURL() + fmt.Sprintf("/restfiles/ds/%s(%s)", url.PathEscape(request.DatasetName), url.PathEscape(request.MemberName))
	} else {
		// For datasets, use the dataset endpoint directly (no /content suffix)
		apiURL = session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(request.DatasetName))
	}

	// Add query parameters
	params := url.Values{}
	if request.Encoding != "" {
		params.Set("encoding", request.Encoding)
	}
	if len(params) > 0 {
		apiURL += "?" + params.Encode()
	}

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

// ListMembers retrieves a list of members in a partitioned dataset
func (dm *ZOSMFDatasetManager) ListMembers(datasetName string) (*MemberList, error) {
	session := dm.session.(*profile.Session)
	
	// Build URL using template
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(datasetName)) + MembersEndpoint

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
	var memberList MemberList
	if err := json.NewDecoder(resp.Body).Decode(&memberList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &memberList, nil
}

// GetMember retrieves information about a specific member
func (dm *ZOSMFDatasetManager) GetMember(datasetName, memberName string) (*DatasetMember, error) {
	session := dm.session.(*profile.Session)
	
	// Build URL using template
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(datasetName)) + fmt.Sprintf(MemberByNameEndpoint, url.PathEscape(memberName))

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
	var member DatasetMember
	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &member, nil
}

// DeleteMember deletes a member from a partitioned dataset
func (dm *ZOSMFDatasetManager) DeleteMember(datasetName, memberName string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL using template
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(datasetName)) + fmt.Sprintf(MemberByNameEndpoint, url.PathEscape(memberName))

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

// Exists checks if a dataset exists using the list API
func (dm *ZOSMFDatasetManager) Exists(name string) (bool, error) {
	// Use the list API with the exact dataset name to check existence
	dl, err := dm.ListDatasets(&DatasetFilter{Name: name})
	if err != nil {
		return false, err
	}
	
	// Check if the dataset was found in the results
	for _, ds := range dl.Datasets {
		if ds.Name == name {
			return true, nil
		}
	}
	
	return false, nil
}

// CopyDataset copies a dataset using the z/OSMF REST API
func (dm *ZOSMFDatasetManager) CopyDataset(sourceName, targetName string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL to the target dataset (z/OSMF format: PUT to target with source in body)
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(targetName))

	// Prepare request body according to z/OSMF API specification
	requestBody := map[string]interface{}{
		"request": "copy",
		"from-dataset": map[string]string{
			"dsn": sourceName,
		},
	}

	// Serialize request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request (PUT to target dataset, not POST to source/copy)
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RenameDataset renames a dataset using the z/OSMF REST API
func (dm *ZOSMFDatasetManager) RenameDataset(oldName, newName string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL to the new dataset name (z/OSMF format: PUT to target with source in body)
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(newName))

	// Prepare request body according to z/OSMF API specification
	requestBody := map[string]interface{}{
		"request": "rename",
		"from-dataset": map[string]string{
			"dsn": oldName,
		},
	}

	// Serialize request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request (PUT to target dataset, not PUT to source/rename)
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := session.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CloseDatasetManager closes the dataset manager and its underlying HTTP connections
func (dm *ZOSMFDatasetManager) CloseDatasetManager() error {
	session := dm.session.(*profile.Session)
	
	// Close idle connections in the HTTP client
	if client := session.GetHTTPClient(); client != nil {
		client.CloseIdleConnections()
	}
	
	return nil
}
