package datasets

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestProfile creates a profile for testing with the given server URL
func createTestProfile(serverURL string) *profile.ZOSMFProfile {
	// Extract host and port from server URL
	host := strings.TrimPrefix(serverURL, "http://")
	host = strings.TrimPrefix(host, "https://")
	
	return &profile.ZOSMFProfile{
		Name:               "test",
		Host:               host,
		Port:               0, // Let the session determine the port from the URL
		User:               "testuser",
		Password:           "testpass",
		RejectUnauthorized: false,
		BasePath:           "/api/v1",
		Protocol:           "http", // Force HTTP for test server
	}
}

func TestNewDatasetManager(t *testing.T) {
	// Create a test session
	profile := &profile.ZOSMFProfile{
		Name:               "test",
		Host:               "localhost",
		Port:               8080,
		User:               "testuser",
		Password:           "testpass",
		RejectUnauthorized: false,
		BasePath:           "/api/v1",
	}

	session, err := profile.NewSession()
	require.NoError(t, err)

	// Create dataset manager
	dm := NewDatasetManager(session)
	assert.NotNil(t, dm)
	assert.Equal(t, session, dm.session)
}

func TestNewDatasetManagerFromProfile(t *testing.T) {
	// Create a test profile
	profile := &profile.ZOSMFProfile{
		Name:               "test",
		Host:               "localhost",
		Port:               8080,
		User:               "testuser",
		Password:           "testpass",
		RejectUnauthorized: false,
		BasePath:           "/api/v1",
	}

	// Create dataset manager from profile
	dm, err := NewDatasetManagerFromProfile(profile)
	require.NoError(t, err)
	assert.NotNil(t, dm)
}

func TestCreateDatasetManager(t *testing.T) {
	// Create a test profile manager
	pm := &profile.ZOSMFProfileManager{}
	
	// This should fail since we don't have a real profile
	_, err := CreateDatasetManager(pm, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get ZOSMF profile")
}

func TestCreateDatasetManagerDirect(t *testing.T) {
	// Create dataset manager directly
	dm, err := CreateDatasetManagerDirect("localhost", 8080, "testuser", "testpass")
	require.NoError(t, err)
	assert.NotNil(t, dm)
}

func TestCreateDatasetManagerDirectWithOptions(t *testing.T) {
	// Create dataset manager with options
	dm, err := CreateDatasetManagerDirectWithOptions("localhost", 8080, "testuser", "testpass", false, "/api/v1")
	require.NoError(t, err)
	assert.NotNil(t, dm)
}

func TestListDatasets(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds", r.URL.Path)
		
		// Return mock response matching z/OSMF API format
		response := DatasetList{
			Datasets: []Dataset{
				{
					Name: "TEST.DATA",
					Type: "PS",
					Used: "512",
				},
			},
			ReturnedRows: 1,
			JSONVersion:  1,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test list datasets
	datasetList, err := dm.ListDatasets(nil)
	require.NoError(t, err)
	assert.Len(t, datasetList.Datasets, 1)
	assert.Equal(t, "TEST.DATA", datasetList.Datasets[0].Name)
}

func TestGetDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds", r.URL.Path)
		// GetDataset now uses the list API with a filter
		
		// Return mock response matching z/OSMF API format
		response := DatasetList{
			Datasets: []Dataset{
				{
					Name: "TEST.DATA",
					Type: "PS",
					Used: "512",
				},
			},
			ReturnedRows: 1,
			JSONVersion:  1,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test get dataset
	dataset, err := dm.GetDataset("TEST.DATA")
	require.NoError(t, err)
	assert.Equal(t, "TEST.DATA", dataset.Name)
	assert.Equal(t, "PS", dataset.Type)
}

func TestCreateDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.DATA", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		// Verify request body
		assert.Equal(t, "TEST.DATA", requestBody["dsname"])
		assert.Equal(t, "PS", requestBody["dsorg"])
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test create dataset
	request := &CreateDatasetRequest{
		Name: "TEST.DATA",
		Type: DatasetTypeSequential,
		Space: Space{
			Primary:   10,
			Secondary: 5,
			Unit:      SpaceUnitTracks,
		},
		RecordFormat: RecordFormatFixed,
		RecordLength: RecordLength80,
		BlockSize:    BlockSize800,
	}
	
	err = dm.CreateDataset(request)
	assert.NoError(t, err)
}

func TestDeleteDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.DATA", r.URL.Path)
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test delete dataset
	err = dm.DeleteDataset("TEST.DATA")
	assert.NoError(t, err)
}

func TestUploadContent(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.DATA", r.URL.Path)
		
		// Verify content type
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(body))
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test upload content
	request := &UploadRequest{
		DatasetName: "TEST.DATA",
		Content:     "Hello, World!",
		Encoding:    "UTF-8",
		Replace:     true,
	}
	
	err = dm.UploadContent(request)
	assert.NoError(t, err)
}

func TestDownloadContent(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.DATA", r.URL.Path)
		
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test download content
	request := &DownloadRequest{
		DatasetName: "TEST.DATA",
		Encoding:    "UTF-8",
	}
	
	content, err := dm.DownloadContent(request)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", content)
}

func TestListMembers(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.PDS/member", r.URL.Path)
		
		// Return mock response matching z/OSMF API format
		response := MemberList{
			Members: []DatasetMember{
				{
					Name: "MEMBER1",
				},
			},
			ReturnedRows: 1,
			JSONVersion:  1,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test list members
	memberList, err := dm.ListMembers("TEST.PDS")
	require.NoError(t, err)
	assert.Len(t, memberList.Members, 1)
	assert.Equal(t, "MEMBER1", memberList.Members[0].Name)
}

func TestGetMember(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.PDS(MEMBER1)", r.URL.Path)
		
		// Return mock member content (z/OSMF returns member content as text)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Member content here"))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test get member
	member, err := dm.GetMember("TEST.PDS", "MEMBER1")
	require.NoError(t, err)
	assert.Equal(t, "MEMBER1", member.Name)
}

func TestDeleteMember(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.PDS(MEMBER1)", r.URL.Path)
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test delete member
	err = dm.DeleteMember("TEST.PDS", "MEMBER1")
	assert.NoError(t, err)
}

func TestExists(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds", r.URL.Path)
		
		// Return mock response matching z/OSMF API format
		response := DatasetList{
			Datasets: []Dataset{
				{
					Name: "TEST.DATA",
					Type: "PS",
				},
			},
			ReturnedRows: 1,
			JSONVersion:  1,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test exists
	exists, err := dm.Exists("TEST.DATA")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCopySequentialDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TARGET.DATA", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		// Verify request body structure
		assert.Equal(t, "copy", requestBody["request"])
		fromDataset := requestBody["from-dataset"].(map[string]interface{})
		assert.Equal(t, "SOURCE.DATA", fromDataset["dsn"])
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test copy dataset
	err = dm.CopySequentialDataset("SOURCE.DATA", "TARGET.DATA")
	assert.NoError(t, err)
}

func TestCopyMember(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TARGET.PDS(MEMBER2)", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		// Verify request body structure
		assert.Equal(t, "copy", requestBody["request"])
		fromDataset := requestBody["from-dataset"].(map[string]interface{})
		assert.Equal(t, "SOURCE.PDS", fromDataset["dsn"])
		assert.Equal(t, "MEMBER1", fromDataset["member"])
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test copy member
	err = dm.CopyMember("SOURCE.PDS", "MEMBER1", "TARGET.PDS", "MEMBER2")
	require.NoError(t, err)
}

func TestRenameDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/NEW.DATA", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		// Verify request body structure
		assert.Equal(t, "rename", requestBody["request"])
		fromDataset := requestBody["from-dataset"].(map[string]interface{})
		assert.Equal(t, "OLD.DATA", fromDataset["dsn"])
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test rename dataset
	err = dm.RenameDataset("OLD.DATA", "NEW.DATA")
	assert.NoError(t, err)
}

func TestCloseDatasetManager(t *testing.T) {
	// Create a test session
	profile := &profile.ZOSMFProfile{
		Name:               "test",
		Host:               "localhost",
		Port:               8080,
		User:               "testuser",
		Password:           "testpass",
		RejectUnauthorized: false,
		BasePath:           "/api/v1",
	}

	session, err := profile.NewSession()
	require.NoError(t, err)

	// Create dataset manager
	dm := NewDatasetManager(session)
	assert.NotNil(t, dm)

	// Test closing the dataset manager
	err = dm.CloseDatasetManager()
	assert.NoError(t, err)
}

// Test convenience functions
func TestCreateSequentialDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.SEQ", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		// Verify request body
		assert.Equal(t, "TEST.SEQ", requestBody["dsname"])
		assert.Equal(t, "PS", requestBody["dsorg"])
		assert.Equal(t, "TRK", requestBody["alcunit"])
		assert.Equal(t, float64(10), requestBody["primary"])
		assert.Equal(t, float64(5), requestBody["secondary"])
		assert.Equal(t, "V", requestBody["recfm"])
		assert.Equal(t, float64(256), requestBody["lrecl"])
		assert.Equal(t, float64(27920), requestBody["blksize"])
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test create sequential dataset
	err = dm.CreateSequentialDataset("TEST.SEQ")
	assert.NoError(t, err)
}

func TestCreatePartitionedDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.PDS", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		// Verify request body
		assert.Equal(t, "TEST.PDS", requestBody["dsname"])
		assert.Equal(t, "PO", requestBody["dsorg"])
		assert.Equal(t, "TRK", requestBody["alcunit"])
		assert.Equal(t, float64(10), requestBody["primary"])
		assert.Equal(t, float64(5), requestBody["secondary"])
		assert.Equal(t, float64(5), requestBody["dirblk"])
		assert.Equal(t, "V", requestBody["recfm"])
		assert.Equal(t, float64(256), requestBody["lrecl"])
		assert.Equal(t, float64(27920), requestBody["blksize"])
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test create partitioned dataset
	err = dm.CreatePartitionedDataset("TEST.PDS")
	assert.NoError(t, err)
}

func TestUploadText(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.DATA", r.URL.Path)
		
		// Verify content type
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(body))
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test upload text
	err = dm.UploadText("TEST.DATA", "Hello, World!")
	assert.NoError(t, err)
}

func TestUploadTextToMember(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)  // Changed from POST to PUT for members
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.PDS(MEMBER1)", r.URL.Path)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))  // Changed from JSON to plain text
		
		// Read request body as plain text
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		
		// Verify request body is plain text
		assert.Equal(t, "Hello, World!", string(body))
		
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test upload text to member
	err = dm.UploadTextToMember("TEST.PDS", "MEMBER1", "Hello, World!")
	assert.NoError(t, err)
}

func TestDownloadText(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.DATA", r.URL.Path)
		assert.Equal(t, "UTF-8", r.URL.Query().Get("encoding"))
		
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test download text
	content, err := dm.DownloadText("TEST.DATA")
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", content)
}

func TestDownloadTextFromMember(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restfiles/ds/TEST.PDS(MEMBER1)", r.URL.Path)
		assert.Equal(t, "UTF-8", r.URL.Query().Get("encoding"))
		
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test download text from member
	content, err := dm.DownloadTextFromMember("TEST.PDS", "MEMBER1")
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", content)
}

// Test validation functions
func TestValidateDatasetName(t *testing.T) {
	// Test valid names
	validNames := []string{
		"TEST.DATA",
		"USER.PROGRAM",
		"SYSTEM.LIBRARY",
		"MY@DATA",
		"TEST#FILE",
		"DATA$SET",
		"A.B.C",
	}

	for _, name := range validNames {
		err := ValidateDatasetName(name)
		assert.NoError(t, err, "Dataset name '%s' should be valid", name)
	}

	// Test invalid names
	invalidNames := []string{
		"",                    // Empty
		"toolongdatasetnamewithwaytoomanycharacters", // Too long
		"test.data",           // Lowercase
		"123.DATA",            // Starts with number
		".DATA",               // Starts with period
		"DATA.",               // Ends with period
		"DATA..SET",           // Consecutive periods
		"DATA--SET",           // Consecutive hyphens
		"DATA SET",            // Contains space
	}

	for _, name := range invalidNames {
		err := ValidateDatasetName(name)
		assert.Error(t, err, "Dataset name '%s' should be invalid", name)
	}
}

func TestValidateMemberName(t *testing.T) {
	// Test valid names
	validNames := []string{
		"MEMBER1",
		"PROGRAM",
		"TEST@1",
		"FILE#1",
		"DATA$1",
		"A.B",
	}

	for _, name := range validNames {
		err := ValidateMemberName(name)
		assert.NoError(t, err, "Member name '%s' should be valid", name)
	}

	// Test invalid names
	invalidNames := []string{
		"",                    // Empty
		"TOOLONG12",           // Too long (10 characters)
		"member1",             // Lowercase
		"123MEMBER",           // Starts with number
		".MEMBER",             // Starts with period
		"MEMBER.",             // Ends with period
		"MEM..BER",            // Consecutive periods
		"MEM BER",             // Contains space
	}

	for _, name := range invalidNames {
		err := ValidateMemberName(name)
		assert.Error(t, err, "Member name '%s' should be invalid", name)
	}
}

func TestValidateCreateDatasetRequest(t *testing.T) {
	// Test valid request
	validRequest := &CreateDatasetRequest{
		Name: "TEST.DATA",
		Type: DatasetTypeSequential,
		Space: Space{
			Primary:   10,
			Secondary: 5,
			Unit:      SpaceUnitTracks,
		},
		RecordFormat: RecordFormatFixed,
		RecordLength: RecordLength80,
		BlockSize:    BlockSize800,
	}

	err := ValidateCreateDatasetRequest(validRequest)
	assert.NoError(t, err)

	// Test invalid requests
	invalidRequests := []*CreateDatasetRequest{
		nil, // Nil request
		{
			Name: "", // Invalid name
			Type: DatasetTypeSequential,
		},
		{
			Name: "TEST.DATA",
			Type: "INVALID", // Invalid type
		},
		{
			Name: "TEST.DATA",
			Type: DatasetTypeSequential,
			Space: Space{
				Primary: 0, // Invalid primary space
			},
		},
	}

	for _, request := range invalidRequests {
		err := ValidateCreateDatasetRequest(request)
		assert.Error(t, err)
	}
}

func TestValidateUploadRequest(t *testing.T) {
	// Test valid request
	validRequest := &UploadRequest{
		DatasetName: "TEST.DATA",
		Content:     "Hello, World!",
	}

	err := ValidateUploadRequest(validRequest)
	assert.NoError(t, err)

	// Test invalid requests
	invalidRequests := []*UploadRequest{
		nil, // Nil request
		{
			DatasetName: "", // Invalid name
			Content:     "Hello, World!",
		},
		{
			DatasetName: "TEST.DATA",
			Content:     "", // Empty content
		},
	}

	for _, request := range invalidRequests {
		err := ValidateUploadRequest(request)
		assert.Error(t, err)
	}
}

func TestValidateDownloadRequest(t *testing.T) {
	// Test valid request
	validRequest := &DownloadRequest{
		DatasetName: "TEST.DATA",
	}

	err := ValidateDownloadRequest(validRequest)
	assert.NoError(t, err)

	// Test invalid requests
	invalidRequests := []*DownloadRequest{
		nil, // Nil request
		{
			DatasetName: "", // Invalid name
		},
	}

	for _, request := range invalidRequests {
		err := ValidateDownloadRequest(request)
		assert.Error(t, err)
	}
}

// Test space creation functions
func TestCreateDefaultSpace(t *testing.T) {
	space := CreateDefaultSpace(SpaceUnitTracks)
	assert.Equal(t, 10, space.Primary)
	assert.Equal(t, 5, space.Secondary)
	assert.Equal(t, SpaceUnitTracks, space.Unit)
	assert.Equal(t, 5, space.Directory)
}

func TestCreateLargeSpace(t *testing.T) {
	space := CreateLargeSpace(SpaceUnitCylinders)
	assert.Equal(t, 100, space.Primary)
	assert.Equal(t, 50, space.Secondary)
	assert.Equal(t, SpaceUnitCylinders, space.Unit)
	assert.Equal(t, 20, space.Directory)
}

func TestCreateSmallSpace(t *testing.T) {
	space := CreateSmallSpace(SpaceUnitKB)
	assert.Equal(t, 5, space.Primary)
	assert.Equal(t, 2, space.Secondary)
	assert.Equal(t, SpaceUnitKB, space.Unit)
	assert.Equal(t, 2, space.Directory)
}

// Test error scenarios
func TestListDatasetsError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test list datasets error
	_, err = dm.ListDatasets(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500")
}

func TestGetDatasetError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Dataset not found"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test get dataset error
	_, err = dm.GetDataset("NONEXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 404")
}

func TestCreateDatasetError(t *testing.T) {
	// Create test server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid dataset request"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test create dataset error
	request := &CreateDatasetRequest{
		Name: "TEST.DATA",
		Type: DatasetTypeSequential,
	}
	
	err = dm.CreateDataset(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}

func TestDeleteDatasetError(t *testing.T) {
	// Create test server that returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test delete dataset error
	err = dm.DeleteDataset("TEST.DATA")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 403")
}

func TestUploadContentError(t *testing.T) {
	// Create test server that returns 409
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Dataset in use"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test upload content error
	request := &UploadRequest{
		DatasetName: "TEST.DATA",
		Content:     "Hello, World!",
	}
	
	err = dm.UploadContent(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 409")
}

func TestDownloadContentError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Dataset not found"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test download content error
	request := &DownloadRequest{
		DatasetName: "NONEXISTENT",
	}
	
	_, err = dm.DownloadContent(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 404")
}

func TestListMembersError(t *testing.T) {
	// Create test server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Not a partitioned dataset"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test list members error
	_, err = dm.ListMembers("TEST.SEQ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}

func TestGetMemberError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Member not found"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test get member error
	_, err = dm.GetMember("TEST.PDS", "NONEXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 404")
}

func TestDeleteMemberError(t *testing.T) {
	// Create test server that returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test delete member error
	err = dm.DeleteMember("TEST.PDS", "MEMBER1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 403")
}

func TestCopySequentialDatasetError(t *testing.T) {
	// Create test server that returns 409
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Target dataset exists"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test copy dataset error
	err = dm.CopySequentialDataset("SOURCE.DATA", "TARGET.DATA")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 409")
}

func TestRenameDatasetError(t *testing.T) {
	// Create test server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid new name"}`))
	}))
	defer server.Close()

	// Create dataset manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	dm := NewDatasetManager(session)

	// Test rename dataset error
	err = dm.RenameDataset("OLD.DATA", "NEW.DATA")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}
