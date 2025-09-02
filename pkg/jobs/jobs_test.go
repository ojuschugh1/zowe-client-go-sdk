package jobs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestNewJobManager(t *testing.T) {
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

	// Create job manager
	jm := NewJobManager(session)
	assert.NotNil(t, jm)
	assert.Equal(t, session, jm.session)
}

func TestNewJobManagerFromProfile(t *testing.T) {
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

	// Create job manager from profile
	jm, err := NewJobManagerFromProfile(profile)
	require.NoError(t, err)
	assert.NotNil(t, jm)
}

func TestCreateJobManager(t *testing.T) {
	// Create a test profile manager
	pm := &profile.ZOSMFProfileManager{}
	
	// This should fail since we don't have a real profile
	_, err := CreateJobManager(pm, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get ZOSMF profile")
}

func TestCreateJobManagerDirect(t *testing.T) {
	// Create job manager directly
	jm, err := CreateJobManagerDirect("localhost", 8080, "testuser", "testpass")
	require.NoError(t, err)
	assert.NotNil(t, jm)
}

func TestCreateJobManagerDirectWithOptions(t *testing.T) {
	// Create job manager with options
	jm, err := CreateJobManagerDirectWithOptions("localhost", 8080, "testuser", "testpass", false, "/api/v1")
	require.NoError(t, err)
	assert.NotNil(t, jm)
}

func TestListJobs(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		
		// Check query parameters
		assert.Equal(t, "testuser", r.URL.Query().Get("owner"))
		assert.Equal(t, "TEST", r.URL.Query().Get("prefix"))
		assert.Equal(t, "10", r.URL.Query().Get("max-jobs"))
		
		// Return mock response
		response := JobList{
			Jobs: []Job{
				{
					JobID:   "JOB001",
					JobName: "TESTJOB1",
					Owner:   "testuser",
					Status:  "OUTPUT",
				},
				{
					JobID:   "JOB002",
					JobName: "TESTJOB2",
					Owner:   "testuser",
					Status:  "ACTIVE",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager with test server
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test list jobs with filter
	filter := &JobFilter{
		Owner:   "testuser",
		Prefix:  "TEST",
		MaxJobs: 10,
	}

	jobList, err := jm.ListJobs(filter)
	require.NoError(t, err)
	assert.Len(t, jobList.Jobs, 2)
	assert.Equal(t, "JOB001", jobList.Jobs[0].JobID)
	assert.Equal(t, "TESTJOB1", jobList.Jobs[0].JobName)
	assert.Equal(t, "OUTPUT", jobList.Jobs[0].Status)
}

func TestGetJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/TESTJOB1/JOB001", r.URL.Path)
		
		// Return mock response
		job := Job{
			JobID:   "JOB001",
			JobName: "TESTJOB1",
			Owner:   "testuser",
			Status:  "OUTPUT",
			RetCode: "CC 0000",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get job
	job, err := jm.GetJob("TESTJOB1:JOB001")
	require.NoError(t, err)
	assert.Equal(t, "JOB001", job.JobID)
	assert.Equal(t, "TESTJOB1", job.JobName)
	assert.Equal(t, "testuser", job.Owner)
	assert.Equal(t, "OUTPUT", job.Status)
	assert.Equal(t, "CC 0000", job.RetCode)
}

func TestGetJobStatus(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/TESTJOB1/JOB001", r.URL.Path)
		
		// Return mock response
		job := Job{
			JobID:   "JOB001",
			JobName: "TESTJOB1",
			Owner:   "testuser",
			Status:  "OUTPUT",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get job status
	status, err := jm.GetJobStatus("TESTJOB1:JOB001")
	require.NoError(t, err)
	assert.Equal(t, "OUTPUT", status)
}

func TestSubmitJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		
		// Parse request body (should be plain text for JCL)
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A", string(body))
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		
		// Return mock response
		response := SubmitJobResponse{
			JobID:   "JOB001",
			JobName: "TESTJOB",
			Owner:   "testuser",
			Status:  "ACTIVE",
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test submit job
	request := &SubmitJobRequest{
		JobStatement: "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A",
	}

	response, err := jm.SubmitJob(request)
	require.NoError(t, err)
	assert.Equal(t, "JOB001", response.JobID)
	assert.Equal(t, "TESTJOB", response.JobName)
	assert.Equal(t, "testuser", response.Owner)
	assert.Equal(t, "ACTIVE", response.Status)
}

func TestSubmitJobStatement(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		
		// Return mock response
		response := SubmitJobResponse{
			JobID:   "JOB001",
			JobName: "TESTJOB",
			Owner:   "testuser",
			Status:  "ACTIVE",
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test submit job statement
	response, err := jm.SubmitJobStatement("//TESTJOB JOB (ACCT),'USER',MSGCLASS=A")
	require.NoError(t, err)
	assert.Equal(t, "JOB001", response.JobID)
}

func TestSubmitJobFromDataset(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		
		// Return mock response
		response := SubmitJobResponse{
			JobID:   "JOB001",
			JobName: "TESTJOB",
			Owner:   "testuser",
			Status:  "ACTIVE",
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test submit job from dataset
	response, err := jm.SubmitJobFromDataset("TEST.JCL", "")
	require.NoError(t, err)
	assert.Equal(t, "JOB001", response.JobID)
}

func TestCancelJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/JOB001/cancel", r.URL.Path)
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test cancel job
	err = jm.CancelJob("JOB001")
	require.NoError(t, err)
}

func TestDeleteJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/TESTJOB1/JOB001", r.URL.Path)
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test delete job
	err = jm.DeleteJob("TESTJOB1:JOB001")
	require.NoError(t, err)
}

func TestGetSpoolFiles(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/TESTJOB/JOB001/files", r.URL.Path)
		
		// Return mock response
		spoolFiles := []SpoolFile{
			{
				ID:      1,
				DDName:  "JESMSGLG",
				Records: 10,
				Bytes:   1000,
			},
			{
				ID:      2,
				DDName:  "JESJCL",
				Records: 5,
				Bytes:   500,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spoolFiles)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get spool files
	spoolFiles, err := jm.GetSpoolFiles("TESTJOB", "JOB001")
	require.NoError(t, err)
	assert.Len(t, spoolFiles, 2)
	assert.Equal(t, "JESMSGLG", spoolFiles[0].DDName)
	assert.Equal(t, 1, spoolFiles[0].ID)
	assert.Equal(t, "JESJCL", spoolFiles[1].DDName)
	assert.Equal(t, 2, spoolFiles[1].ID)
}

func TestGetSpoolFileContent(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/TESTJOB/JOB001/files/1/records", r.URL.Path)
		
		// Return mock content
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("JES2 JOB LOG OUTPUT"))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get spool file content
	content, err := jm.GetSpoolFileContent("TESTJOB", "JOB001", 1)
	require.NoError(t, err)
	assert.Equal(t, "JES2 JOB LOG OUTPUT", content)
}

func TestPurgeJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/JOB001/purge", r.URL.Path)
		
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test purge job
	err = jm.PurgeJob("JOB001")
	require.NoError(t, err)
}

func TestIsJobComplete(t *testing.T) {
	// Test completed statuses
	assert.True(t, isJobComplete("OUTPUT"))
	assert.True(t, isJobComplete("CC 0000"))
	assert.True(t, isJobComplete("CC 0001"))
	assert.True(t, isJobComplete("CC 0002"))
	assert.True(t, isJobComplete("CC 0003"))
	assert.True(t, isJobComplete("CC 0004"))
	assert.True(t, isJobComplete("ABEND"))

	// Test active statuses
	assert.False(t, isJobComplete("ACTIVE"))
	assert.False(t, isJobComplete("INPUT"))
	assert.False(t, isJobComplete("RUNNING"))
}

func TestValidateJobRequest(t *testing.T) {
	// Test valid request
	validRequest := &SubmitJobRequest{
		JobStatement: "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A",
	}
	err := ValidateJobRequest(validRequest)
	assert.NoError(t, err)

	// Test invalid request - no job source
	invalidRequest := &SubmitJobRequest{}
	err = ValidateJobRequest(invalidRequest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one job source must be specified")

	// Test invalid request - nil request
	err = ValidateJobRequest(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job request cannot be nil")

	// Test invalid request - job statement without JOB card
	invalidJobStatement := &SubmitJobRequest{
		JobStatement: "//STEP1 EXEC PGM=IEFBR14",
	}
	err = ValidateJobRequest(invalidJobStatement)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job statement must contain a JOB card")
}

func TestIsValidDatasetName(t *testing.T) {
	// Test valid dataset names
	assert.True(t, isValidDatasetName("TEST.DATASET"))
	assert.True(t, isValidDatasetName("USER.TEST.DATA"))
	assert.True(t, isValidDatasetName("A@B#C$D-E.F"))

	// Test invalid dataset names
	assert.False(t, isValidDatasetName("")) // Empty
	assert.False(t, isValidDatasetName("1TEST.DATASET")) // Starts with number
	assert.False(t, isValidDatasetName("TEST.DATASET.WITH.TOO.MANY.QUALIFIERS.THAT.EXCEEDS.THE.MAXIMUM.LENGTH")) // Too long
}

func TestCreateSimpleJobStatement(t *testing.T) {
	// Test with all parameters
	statement := CreateSimpleJobStatement("TESTJOB", "ACCT", "USER", "A", "(1,1)")
	expected := "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A,MSGLEVEL=(1,1)"
	assert.Equal(t, expected, statement)

	// Test with default values
	statement = CreateSimpleJobStatement("", "", "", "", "")
	expected = "//GOJOB JOB (ACCT),'USER',MSGCLASS=A,MSGLEVEL=(1,1)"
	assert.Equal(t, expected, statement)
}

func TestCreateJobWithStep(t *testing.T) {
	ddStatements := []string{
		"//DD1 DD DSN=TEST.DATA,DISP=SHR",
		"//DD2 DD SYSOUT=A",
	}

	job := CreateJobWithStep("TESTJOB", "ACCT", "USER", "A", "(1,1)", "STEP1", "IEFBR14", ddStatements)
	
	expected := "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A,MSGLEVEL=(1,1)\n//STEP1 EXEC PGM=IEFBR14\n//DD1 DD DSN=TEST.DATA,DISP=SHR\n//DD2 DD SYSOUT=A\n"
	assert.Equal(t, expected, job)
}

func TestGetJobsByOwner(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		assert.Equal(t, "testuser", r.URL.Query().Get("owner"))
		assert.Equal(t, "10", r.URL.Query().Get("max-jobs"))
		
		// Return mock response
		response := JobList{
			Jobs: []Job{
				{
					JobID:   "JOB001",
					JobName: "TESTJOB1",
					Owner:   "testuser",
					Status:  "OUTPUT",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get jobs by owner
	jobList, err := jm.GetJobsByOwner("testuser", 10)
	require.NoError(t, err)
	assert.Len(t, jobList.Jobs, 1)
	assert.Equal(t, "testuser", jobList.Jobs[0].Owner)
}

func TestGetJobsByPrefix(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		assert.Equal(t, "TEST", r.URL.Query().Get("prefix"))
		assert.Equal(t, "5", r.URL.Query().Get("max-jobs"))
		
		// Return mock response
		response := JobList{
			Jobs: []Job{
				{
					JobID:   "JOB001",
					JobName: "TESTJOB1",
					Owner:   "testuser",
					Status:  "OUTPUT",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get jobs by prefix
	jobList, err := jm.GetJobsByPrefix("TEST", 5)
	require.NoError(t, err)
	assert.Len(t, jobList.Jobs, 1)
	assert.Equal(t, "TESTJOB1", jobList.Jobs[0].JobName)
}

func TestGetJobsByStatus(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		assert.Equal(t, "OUTPUT", r.URL.Query().Get("status"))
		assert.Equal(t, "20", r.URL.Query().Get("max-jobs"))
		
		// Return mock response
		response := JobList{
			Jobs: []Job{
				{
					JobID:   "JOB001",
					JobName: "TESTJOB1",
					Owner:   "testuser",
					Status:  "OUTPUT",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)

	session, err := profile.NewSession()
	require.NoError(t, err)

	jm := NewJobManager(session)

	// Test get jobs by status
	jobList, err := jm.GetJobsByStatus("OUTPUT", 20)
	require.NoError(t, err)
	assert.Len(t, jobList.Jobs, 1)
	assert.Equal(t, "OUTPUT", jobList.Jobs[0].Status)
}

func TestCloseJobManager(t *testing.T) {
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

	// Create job manager
	jm := NewJobManager(session)
	assert.NotNil(t, jm)

	// Test closing the job manager
	err = jm.CloseJobManager()
	assert.NoError(t, err)
}

// Test error scenarios and edge cases
func TestSubmitJobErrors(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid job request"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test submit job with no source specified
	request := &SubmitJobRequest{}
	_, err = jm.SubmitJob(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no job source specified")

	// Test submit job with invalid request
	request = &SubmitJobRequest{
		JobStatement: "//TESTJOB JOB",
	}
	_, err = jm.SubmitJob(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}

func TestGetJobErrors(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Job not found"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test get non-existent job
	_, err = jm.GetJob("NONEXISTENT:JOB999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 404")
}

func TestCancelJobErrors(t *testing.T) {
	// Create test server that returns 409 (conflict)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Job cannot be cancelled"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test cancel job error
	err = jm.CancelJob("JOB001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 409")
}

func TestDeleteJobErrors(t *testing.T) {
	// Create test server that returns 403 (forbidden)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Permission denied"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test delete job error
	err = jm.DeleteJob("TESTJOB1:JOB001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 403")
}

func TestGetSpoolFilesErrors(t *testing.T) {
	// Create test server that returns 500 (internal server error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test get spool files error
	_, err = jm.GetSpoolFiles("INVALID", "JOB001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500")
}

func TestGetSpoolFileContentErrors(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Spool file not found"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test get spool file content error
	_, err = jm.GetSpoolFileContent("INVALID", "JOB001", 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 404")
}

func TestPurgeJobErrors(t *testing.T) {
	// Create test server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Cannot purge job"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test purge job error
	err = jm.PurgeJob("JOB001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}

func TestWaitForJobCompletionTimeout(t *testing.T) {
	// Create test server that always returns running status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Job{
			JobID:   "JOB001",
			JobName: "TESTJOB",
			Status:  "RUNNING",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test timeout scenario
	_, err = jm.WaitForJobCompletion("TESTJOB1:JOB001", 100*time.Millisecond, 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for job")
}

func TestWaitForJobCompletionError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Job not found"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test error during wait
	_, err = jm.WaitForJobCompletion("TESTJOB1:JOB001", 1*time.Second, 100*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get job status")
}

func TestSubmitJobWithAllSources(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		
		// Parse request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		response := SubmitJobResponse{
			JobID:   "JOB001",
			JobName: "TESTJOB",
			Owner:   "testuser",
			Status:  "INPUT",
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test submit job with dataset
	request := &SubmitJobRequest{
		JobDataSet: "TEST.JCL",
		Volume:     "VOL001",
	}
	response, err := jm.SubmitJob(request)
	require.NoError(t, err)
	assert.Equal(t, "JOB001", response.JobID)

	// Test submit job with local file
	request = &SubmitJobRequest{
		JobLocalFile: "test.jcl",
		Directory:    "/tmp",
		Extension:    "jcl",
	}
	response, err = jm.SubmitJob(request)
	require.NoError(t, err)
	assert.Equal(t, "JOB001", response.JobID)
}

func TestListJobsWithAllFilters(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs", r.URL.Path)
		
		// Check all query parameters
		assert.Equal(t, "testuser", r.URL.Query().Get("owner"))
		assert.Equal(t, "TEST", r.URL.Query().Get("prefix"))
		assert.Equal(t, "10", r.URL.Query().Get("max-jobs"))
		assert.Equal(t, "JOB001", r.URL.Query().Get("jobid"))
		assert.Equal(t, "TESTJOB", r.URL.Query().Get("jobname"))
		assert.Equal(t, "OUTPUT", r.URL.Query().Get("status"))
		assert.Equal(t, "CORRELATOR", r.URL.Query().Get("user-correlator"))
		
		response := JobList{
			Jobs: []Job{
				{
					JobID:   "JOB001",
					JobName: "TESTJOB",
					Owner:   "testuser",
					Status:  "OUTPUT",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test with all filters
	filter := &JobFilter{
		Owner:           "testuser",
		Prefix:          "TEST",
		MaxJobs:         10,
		JobID:           "JOB001",
		JobName:         "TESTJOB",
		Status:          "OUTPUT",
		UserCorrelator:  "CORRELATOR",
	}
	
	jobList, err := jm.ListJobs(filter)
	require.NoError(t, err)
	assert.Len(t, jobList.Jobs, 1)
	assert.Equal(t, "JOB001", jobList.Jobs[0].JobID)
}

func TestGetJobInfo(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/restjobs/jobs/TESTJOB1/JOB001/files", r.URL.Path)
		
		response := JobInfo{
			JobID:   "JOB001",
			JobName: "TESTJOB",
			Owner:   "testuser",
			Status:  "OUTPUT",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test get job info
	jobInfo, err := jm.GetJobInfo("TESTJOB1:JOB001")
	require.NoError(t, err)
	assert.Equal(t, "JOB001", jobInfo.JobID)
	assert.Equal(t, "TESTJOB", jobInfo.JobName)
}

func TestGetJobInfoError(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Job info not found"}`))
	}))
	defer server.Close()

	// Create job manager
	profile := createTestProfile(server.URL)
	session, err := profile.NewSession()
	require.NoError(t, err)
	jm := NewJobManager(session)

	// Test get job info error
	_, err = jm.GetJobInfo("TESTJOB1:JOB001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 404")
}
