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

// z/OSMF dataset API endpoints
const (
	// Main datasets endpoint
	DatasetsEndpoint = "/restfiles/ds"
	
	// Dataset by name
	DatasetByNameEndpoint = "/restfiles/ds/%s"
	
	// Member endpoints
	MembersEndpoint  = "/member"
	ContentEndpoint  = "/content"
	
	// Member by name
	MemberByNameEndpoint = "/member/%s"
	
	// Content endpoints
	DatasetContentEndpoint = "/content"
	MemberContentEndpoint  = "/content/%s"
)

// NewDatasetManager creates a dataset manager with the given session
func NewDatasetManager(session *profile.Session) *ZOSMFDatasetManager {
	return &ZOSMFDatasetManager{
		session: session,
	}
}

// NewDatasetManagerFromProfile creates a dataset manager from a profile
func NewDatasetManagerFromProfile(profile *profile.ZOSMFProfile) (*ZOSMFDatasetManager, error) {
	session, err := profile.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return NewDatasetManager(session), nil
}

// ListDatasets gets datasets matching the filter
func (dm *ZOSMFDatasetManager) ListDatasets(filter *DatasetFilter) (*DatasetList, error) {
	session := dm.session.(*profile.Session)
	
	// Build query parameters
	params := url.Values{}
	
	// Need either dslevel or volser parameter
	hasRequiredParam := false
	
	if filter != nil {
		if filter.Name != "" {
			// Dataset name pattern (wildcards supported)
			params.Set("dslevel", filter.Name)
			hasRequiredParam = true
		}
		if filter.Volume != "" {
			// Volume serial number
			params.Set("volser", filter.Volume)
			hasRequiredParam = true
		}
		if filter.Owner != "" {
			// Starting dataset name for pagination
			params.Set("start", filter.Owner)
		}
		// Limit is handled via header, not query param
	}
	
	// Default to user's datasets if no filter specified
	if !hasRequiredParam {
		// Use user ID to avoid listing everything
		params.Set("dslevel", session.User+".*")
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
	
	// Set result limit
	if filter != nil && filter.Limit > 0 {
		req.Header.Set("X-IBM-Max-Items", strconv.Itoa(filter.Limit))
	} else {
		req.Header.Set("X-IBM-Max-Items", "0") // 0 = no limit
	}
	
	// Get basic attributes only
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

// GetDataset gets info for a specific dataset
func (dm *ZOSMFDatasetManager) GetDataset(name string) (*Dataset, error) {
	// Use list API to get metadata
	dl, err := dm.ListDatasets(&DatasetFilter{Name: name})
	if err != nil {
		return nil, err
	}
	
	// Find the dataset in results
	for _, ds := range dl.Datasets {
		if ds.Name == name {
			return &ds, nil
		}
	}
	
	return nil, fmt.Errorf("dataset not found: %s", name)
}

// GetDatasetInfo gets detailed dataset info, trying direct API first
func (dm *ZOSMFDatasetManager) GetDatasetInfo(name string) (*Dataset, error) {
	// Try direct API first
	dataset, err := dm.getDatasetInfoDirect(name)
	if err == nil {
		return dataset, nil
	}
	
	// Fall back to list API
	return dm.GetDataset(name)
}

// getDatasetInfoDirect tries to get dataset info via direct API
func (dm *ZOSMFDatasetManager) getDatasetInfoDirect(name string) (*Dataset, error) {
	session := dm.session.(*profile.Session)
	
	// Build URL for direct dataset access
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(name))
	
	// Request metadata, not content
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
	
	// Build URL using correct z/OSMF format: /zosmf/restfiles/ds/<dataset-name>(<member-name>)
	apiURL := session.GetBaseURL() + fmt.Sprintf("/restfiles/ds/%s(%s)", url.PathEscape(datasetName), url.PathEscape(memberName))

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

	// For member access, z/OSMF returns the member content as text, not JSON
	// We'll create a DatasetMember with the member name since we can't get metadata
	// from this endpoint. The member exists if we get a successful response.
	member := &DatasetMember{
		Name: memberName,
	}

	return member, nil
}

// DeleteMember deletes a member from a partitioned dataset
func (dm *ZOSMFDatasetManager) DeleteMember(datasetName, memberName string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL using correct z/OSMF format: /zosmf/restfiles/ds/<dataset-name>(<member-name>)
	apiURL := session.GetBaseURL() + fmt.Sprintf("/restfiles/ds/%s(%s)", url.PathEscape(datasetName), url.PathEscape(memberName))

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

// CopySequentialDataset copies a sequential dataset using the z/OSMF REST API
// This function handles copying entire datasets (not members)
func (dm *ZOSMFDatasetManager) CopySequentialDataset(sourceName, targetName string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL to the target dataset (z/OSMF format: PUT to target with source in body)
	apiURL := session.GetBaseURL() + fmt.Sprintf(DatasetByNameEndpoint, url.PathEscape(targetName))

	// Prepare request body according to z/OSMF API specification for dataset copy
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

// CopyMember copies a member from one partitioned dataset to another using the z/OSMF REST API
// sourceName should be in format "DATASET.NAME" and sourceMember is the member name
// targetName should be in format "DATASET.NAME" and targetMember is the member name
func (dm *ZOSMFDatasetManager) CopyMember(sourceName, sourceMember, targetName, targetMember string) error {
	session := dm.session.(*profile.Session)
	
	// Build URL to the target member using correct z/OSMF format: /zosmf/restfiles/ds/<target-dataset>(<target-member>)
	apiURL := session.GetBaseURL() + fmt.Sprintf("/restfiles/ds/%s(%s)", url.PathEscape(targetName), url.PathEscape(targetMember))

	// Prepare request body according to z/OSMF API specification for member copy
	requestBody := map[string]interface{}{
		"request": "copy",
		"from-dataset": map[string]string{
			"dsn":    sourceName,
			"member": sourceMember,
		},
	}

	// Serialize request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request (PUT to target member)
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
