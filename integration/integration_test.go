package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/jobs"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
	"github.com/stretchr/testify/require"
)

// mockServer implements minimal endpoints needed by the SDK for an end-to-end smoke test
func newMockZosmfServer() *httptest.Server {
	// in-memory state
	type job struct{ ID, Name, Owner, Status string }
	jobsState := []job{}
	spool := map[string][]jobs.SpoolFile{}
	datasetContent := map[string]string{}

	mux := http.NewServeMux()

	// Jobs: list & submit & get & files & records & cancel/delete/purge
	mux.HandleFunc("/api/v1/restjobs/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// return all jobs
			out := jobs.JobList{Jobs: []jobs.Job{}}
			for _, j := range jobsState {
				out.Jobs = append(out.Jobs, jobs.Job{JobID: j.ID, JobName: j.Name, Owner: j.Owner, Status: j.Status})
			}
			_ = json.NewEncoder(w).Encode(out)
			return
		}
		if r.Method == http.MethodPut {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			id := fmt.Sprintf("JOB%08d", len(jobsState)+1)
			j := job{ID: id, Name: "INTEGJOB", Owner: "TESTUSR", Status: "ACTIVE"}
			jobsState = append(jobsState, j)
			spool[id] = []jobs.SpoolFile{{ID: 1, DDName: "JESMSGLG"}}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(jobs.SubmitJobResponse{JobID: id, JobName: j.Name})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/v1/restjobs/jobs/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/v1/restjobs/jobs/"):]
		// /{id}
		if r.Method == http.MethodGet && len(path) > 0 && !strings.Contains(path, "/") {
			// exact id match for get
			id := path
			for _, j := range jobsState {
				if j.ID == id {
					_ = json.NewEncoder(w).Encode(jobs.Job{JobID: j.ID, JobName: j.Name, Owner: j.Owner, Status: j.Status})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// /{id}/files or /{id}/files/{n}/records
		if r.Method == http.MethodGet && len(path) > 0 {
			// files
			if suffix := "/files"; len(path) > len(suffix) && path[len(path)-len(suffix):] == suffix {
				id := path[:len(path)-len("/files")]
				_ = json.NewEncoder(w).Encode(spool[id])
				return
			}
			// records
			if len(path) > 0 && path[len(path)-len("/records"):] == "/records" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("HELLO"))
				return
			}
		}
		// cancel/purge via PUT, delete via DELETE
		if r.Method == http.MethodPut && (len(path) > 0 && (len(path) >= len("/cancel") && path[len(path)-len("/cancel"): ] == "/cancel" || len(path) >= len("/purge") && path[len(path)-len("/purge"): ] == "/purge")) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Datasets minimal endpoints
	mux.HandleFunc("/api/v1/restfiles/ds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(datasets.DatasetList{Datasets: []datasets.Dataset{}, Returned: 0, Total: 0})
			return
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/v1/restfiles/ds/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/v1/restfiles/ds/"):]
		// content endpoints first to avoid matching generic dataset GET
		if len(path) > 0 && len(path) >= len("/content") && path[len(path)-len("/content"): ] == "/content" {
			if r.Method == http.MethodPost {
				ds := path[:len(path)-len("/content")]
				var body struct{ Content string `json:"content"` }
				_ = json.NewDecoder(r.Body).Decode(&body)
				datasetContent[ds] = body.Content
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				ds := path[:len(path)-len("/content")]
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(datasetContent[ds]))
				return
			}
		}
		if r.Method == http.MethodGet && len(path) > 0 {
			_ = json.NewEncoder(w).Encode(datasets.Dataset{Name: path, Type: datasets.DatasetTypeSequential})
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}

func TestIntegration_Smoke(t *testing.T) {
	server := newMockZosmfServer()
	defer server.Close()

	// (no profile needed; use direct session to mock server)

	// The httptest server gives full URL; override session creation by building direct session
	sess := &profile.Session{}
	// simulate http (no TLS) and base path
	sess.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	sess.Headers = map[string]string{"Content-Type": "application/json"}
	sess.BaseURL = server.URL + "/api/v1"

	// Jobs flow
	jm := jobs.NewJobManager(sess)
	resp, err := jm.SubmitJob(&jobs.SubmitJobRequest{JobStatement: "//JOB"})
	require.NoError(t, err)
	require.NotEmpty(t, resp.JobID)
	status, err := jm.GetJobStatus(resp.JobID)
	require.NoError(t, err)
	require.NotEmpty(t, status)
	files, err := jm.GetSpoolFiles(resp.JobID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(files), 1)
	content, err := jm.GetSpoolFileContent(resp.JobID, files[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, content)
	require.NoError(t, jm.CloseJobManager())

	// Datasets flow
	dm := datasets.NewDatasetManager(sess)
	require.NoError(t, dm.CreateDataset(&datasets.CreateDatasetRequest{Name: "TEST.DATA", Type: datasets.DatasetTypeSequential}))
	req := &datasets.UploadRequest{DatasetName: "TEST.DATA", Content: "hello"}
	require.NoError(t, dm.UploadContent(req))
	dl, err := dm.DownloadContent(&datasets.DownloadRequest{DatasetName: "TEST.DATA"})
	require.NoError(t, err)
	require.Equal(t, "hello", dl)
	require.NoError(t, dm.DeleteDataset("TEST.DATA"))
	require.NoError(t, dm.CloseDatasetManager())
}


