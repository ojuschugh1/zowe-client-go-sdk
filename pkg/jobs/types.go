package jobs

import (
	"time"
)

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
	Subsystem   string    `json:"subsystem,omitempty"`
	Type        string    `json:"type,omitempty"`
	Class       string    `json:"class,omitempty"`
	PhaseName   string    `json:"phase-name,omitempty"`
	PhaseNumber int       `json:"phase-number,omitempty"`
	RetCode     string    `json:"retcode,omitempty"`
	URL         string    `json:"url,omitempty"`
	FilesURL    string    `json:"files-url,omitempty"`
	JobCorrelator string  `json:"job-correlator,omitempty"`
	ExecutionClass string `json:"execution-class,omitempty"`
	ExecutionMode string  `json:"execution-mode,omitempty"`
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

// JobList represents a list of jobs
type JobList struct {
	Jobs []Job `json:"jobs"`
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

// SubmitJobResponse represents a job submission response
type SubmitJobResponse struct {
	JobID   string `json:"jobid"`
	JobName string `json:"jobname"`
	Owner   string `json:"owner"`
	Status  string `json:"status"`
	URL     string `json:"url,omitempty"`
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

// JobManager interface for job management operations
type JobManager interface {
	ListJobs(filter *JobFilter) (*JobList, error)
	GetJob(jobID string) (*Job, error)
	GetJobInfo(jobID string) (*JobInfo, error)
	GetJobStatus(jobID string) (string, error)
	GetJobByNameID(jobName, jobID string) (*Job, error)
	GetJobByCorrelator(correlator string) (*Job, error)
	SubmitJob(request *SubmitJobRequest) (*SubmitJobResponse, error)
	CancelJob(jobID string) error
	DeleteJob(jobID string) error
	GetSpoolFiles(jobID string) ([]SpoolFile, error)
	GetSpoolFileContent(jobID string, spoolID int) (string, error)
	PurgeJob(jobID string) error
	CloseJobManager() error
}

// ZOSMFJobManager implements JobManager for ZOSMF
type ZOSMFJobManager struct {
	session interface{} // Will be *profile.Session
}
