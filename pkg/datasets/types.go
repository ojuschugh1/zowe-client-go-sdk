package datasets

import (
	"time"
)

// DatasetType represents the type of dataset
type DatasetType string

const (
	DatasetTypeSequential DatasetType = "SEQ"
	DatasetTypePartitioned DatasetType = "PO"
	DatasetTypePDSE       DatasetType = "PDSE"
	DatasetTypeVSAM       DatasetType = "VSAM"
)

// SpaceUnit represents the unit for space allocation
type SpaceUnit string

const (
	SpaceUnitTracks   SpaceUnit = "TRK"
	SpaceUnitCylinders SpaceUnit = "CYL"
	SpaceUnitKB       SpaceUnit = "KB"
	SpaceUnitMB       SpaceUnit = "MB"
	SpaceUnitGB       SpaceUnit = "GB"
)

// RecordFormat represents the record format
type RecordFormat string

const (
	RecordFormatFixed    RecordFormat = "F"
	RecordFormatVariable RecordFormat = "V"
	RecordFormatUndefined RecordFormat = "U"
)

// RecordLength represents the record length
type RecordLength int

const (
	RecordLength80  RecordLength = 80
	RecordLength132 RecordLength = 132
	RecordLength256 RecordLength = 256
	RecordLength512 RecordLength = 512
)

// BlockSize represents the block size
type BlockSize int

const (
	BlockSize80   BlockSize = 80
	BlockSize800  BlockSize = 800
	BlockSize27920 BlockSize = 27920
	BlockSize32760 BlockSize = 32760
)

// Dataset represents a z/OS dataset
type Dataset struct {
	Name         string      `json:"name"`
	Type         DatasetType `json:"type"`
	Volume       string      `json:"volume,omitempty"`
	Space        Space       `json:"space,omitempty"`
	RecordFormat RecordFormat `json:"recordFormat,omitempty"`
	RecordLength RecordLength `json:"recordLength,omitempty"`
	BlockSize    BlockSize   `json:"blockSize,omitempty"`
	Directory    int         `json:"directory,omitempty"` // For partitioned datasets
	Created      time.Time   `json:"created,omitempty"`
	Modified     time.Time   `json:"modified,omitempty"`
	Size         int64       `json:"size,omitempty"`
	Used         int64       `json:"used,omitempty"`
	Extents      int         `json:"extents,omitempty"`
	Referenced   time.Time   `json:"referenced,omitempty"`
	Expiration   time.Time   `json:"expiration,omitempty"`
	Owner        string      `json:"owner,omitempty"`
	Security     string      `json:"security,omitempty"`
}

// Space represents space allocation parameters
type Space struct {
	Primary   int       `json:"primary"`
	Secondary int       `json:"secondary"`
	Unit      SpaceUnit `json:"unit"`
	Directory int       `json:"directory,omitempty"` // For partitioned datasets
}

// DatasetMember represents a member in a partitioned dataset
type DatasetMember struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Created   time.Time `json:"created,omitempty"`
	Modified  time.Time `json:"modified,omitempty"`
	UserID    string    `json:"userid,omitempty"`
	Version   int       `json:"version,omitempty"`
	ModLevel  int       `json:"modLevel,omitempty"`
	ChangeDate time.Time `json:"changeDate,omitempty"`
}

// DatasetList represents a list of datasets
type DatasetList struct {
	Datasets []Dataset `json:"datasets"`
	Returned int       `json:"returned"`
	Total    int       `json:"total"`
}

// MemberList represents a list of members in a partitioned dataset
type MemberList struct {
	Members  []DatasetMember `json:"members"`
	Returned int             `json:"returned"`
	Total    int             `json:"total"`
}

// CreateDatasetRequest represents a request to create a dataset
type CreateDatasetRequest struct {
	Name         string      `json:"name"`
	Type         DatasetType `json:"type"`
	Volume       string      `json:"volume,omitempty"`
	Space        Space       `json:"space,omitempty"`
	RecordFormat RecordFormat `json:"recordFormat,omitempty"`
	RecordLength RecordLength `json:"recordLength,omitempty"`
	BlockSize    BlockSize   `json:"blockSize,omitempty"`
	Directory    int         `json:"directory,omitempty"`
}

// UploadRequest represents a request to upload content to a dataset
type UploadRequest struct {
	DatasetName string `json:"datasetName"`
	MemberName  string `json:"memberName,omitempty"` // For partitioned datasets
	Content     string `json:"content"`
	Encoding    string `json:"encoding,omitempty"`
	Replace     bool   `json:"replace,omitempty"`
}

// DownloadRequest represents a request to download content from a dataset
type DownloadRequest struct {
	DatasetName string `json:"datasetName"`
	MemberName  string `json:"memberName,omitempty"` // For partitioned datasets
	Encoding    string `json:"encoding,omitempty"`
}

// DatasetFilter represents filters for dataset queries
type DatasetFilter struct {
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Volume string `json:"volume,omitempty"`
	Owner  string `json:"owner,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// DatasetManager interface for dataset management operations
type DatasetManager interface {
	// Basic operations
	ListDatasets(filter *DatasetFilter) (*DatasetList, error)
	GetDataset(name string) (*Dataset, error)
	CreateDataset(request *CreateDatasetRequest) error
	DeleteDataset(name string) error
	
	// Content operations
	UploadContent(request *UploadRequest) error
	DownloadContent(request *DownloadRequest) (string, error)
	
	// Member operations (for partitioned datasets)
	ListMembers(datasetName string) (*MemberList, error)
	GetMember(datasetName, memberName string) (*DatasetMember, error)
	DeleteMember(datasetName, memberName string) error
	
	// Utility operations
	Exists(name string) (bool, error)
	CopyDataset(sourceName, targetName string) error
	RenameDataset(oldName, newName string) error
}

// ZOSMFDatasetManager implements DatasetManager for ZOSMF
type ZOSMFDatasetManager struct {
	session interface{} // Will be *profile.Session
}
