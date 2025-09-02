package datasets

// DatasetType represents the type of dataset
type DatasetType string

const (
	DatasetTypeSequential DatasetType = "PS"   // Sequential
	DatasetTypePartitioned DatasetType = "PO"  // Partitioned (PDS)
	DatasetTypePDSE       DatasetType = "PDSE" // PDSE
	DatasetTypeVSAM       DatasetType = "VSAM" // VSAM
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
	Name         string `json:"dsname"`           // Dataset name
	Type         string `json:"dsorg"`            // Organization (PS, PO, etc.)
	Volume       string `json:"vol,omitempty"`    // Volume serial
	BlockSize    string `json:"blksz,omitempty"`  // Block size
	RecordLength string `json:"lrecl,omitempty"`  // Record length  
	RecordFormat string `json:"recfm,omitempty"`  // Record format
	Catalog      string `json:"catnm,omitempty"`  // Catalog name
	CreatedDate  string `json:"cdate,omitempty"`  // Creation date
	Device       string `json:"dev,omitempty"`    // Device type
	DatasetType  string `json:"dsntp,omitempty"`  // Dataset type
	ExpiryDate   string `json:"edate,omitempty"`  // Expiry date
	Extents      string `json:"extx,omitempty"`   // Number of extents
	Migrated     string `json:"migr,omitempty"`   // Migration status
	MultiVolume  string `json:"mvol,omitempty"`   // Multi-volume
	Overflow     string `json:"ovf,omitempty"`    // Overflow
	RefDate      string `json:"rdate,omitempty"`  // Referenced date
	SizeX        string `json:"sizex,omitempty"`  // Size
	SpaceUnit    string `json:"spacu,omitempty"`  // Space unit
	Used         string `json:"used,omitempty"`   // Used percentage
	VolumeList   string `json:"vols,omitempty"`   // Volume list
}

// Space represents space allocation parameters
type Space struct {
	Primary   int       `json:"primary"`
	Secondary int       `json:"secondary"`
	Unit      SpaceUnit `json:"unit"`
	Directory int       `json:"directory,omitempty"` // For PDS
}

// DatasetMember represents a member in a partitioned dataset
type DatasetMember struct {
	Name string `json:"member"` // Member name
}

// DatasetList represents a list of datasets
type DatasetList struct {
	Datasets     []Dataset `json:"items"`           // Dataset array
	ReturnedRows int       `json:"returnedRows"`    // Rows returned
	MoreRows     bool      `json:"moreRows"`        // More data available
	JSONVersion  int       `json:"JSONversion"`     // API version
}

// MemberList represents a list of members in a PDS
type MemberList struct {
	Members      []DatasetMember `json:"items"`           // Member array
	ReturnedRows int             `json:"returnedRows"`    // Rows returned
	JSONVersion  int             `json:"JSONversion"`     // API version
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

// UploadRequest represents a request to upload content
type UploadRequest struct {
	DatasetName string `json:"datasetName"`
	MemberName  string `json:"memberName,omitempty"` // For PDS members
	Content     string `json:"content"`
	Encoding    string `json:"encoding,omitempty"`
	Replace     bool   `json:"replace,omitempty"`
}

// DownloadRequest represents a request to download content
type DownloadRequest struct {
	DatasetName string `json:"datasetName"`
	MemberName  string `json:"memberName,omitempty"` // For PDS members
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

// DatasetManager interface for dataset operations
type DatasetManager interface {
	// Basic operations
	ListDatasets(filter *DatasetFilter) (*DatasetList, error)
	GetDataset(name string) (*Dataset, error)
	GetDatasetInfo(name string) (*Dataset, error)
	CreateDataset(request *CreateDatasetRequest) error
	DeleteDataset(name string) error
	
	// Content operations
	UploadContent(request *UploadRequest) error
	DownloadContent(request *DownloadRequest) (string, error)
	
	// Member operations (for PDS)
	ListMembers(datasetName string) (*MemberList, error)
	GetMember(datasetName, memberName string) (*DatasetMember, error)
	DeleteMember(datasetName, memberName string) error
	
	// Utility operations
	Exists(name string) (bool, error)
	CopySequentialDataset(sourceName, targetName string) error
	CopyMember(sourceName, sourceMember, targetName, targetMember string) error
	RenameDataset(oldName, newName string) error
}

// ZOSMFDatasetManager implements DatasetManager for ZOSMF
type ZOSMFDatasetManager struct {
	session interface{} // Will be *profile.Session
}
