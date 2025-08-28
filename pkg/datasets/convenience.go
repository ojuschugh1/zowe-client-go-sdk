package datasets

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

// CreateDatasetManager creates a new dataset manager from a profile manager
func CreateDatasetManager(pm *profile.ZOSMFProfileManager, profileName string) (*ZOSMFDatasetManager, error) {
	zosmfProfile, err := pm.GetZOSMFProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ZOSMF profile '%s': %w", profileName, err)
	}

	session, err := zosmfProfile.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewDatasetManager(session), nil
}

// CreateDatasetManagerDirect creates a dataset manager directly with connection parameters
func CreateDatasetManagerDirect(host string, port int, user, password string) (*ZOSMFDatasetManager, error) {
	session, err := profile.CreateSessionDirect(host, port, user, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewDatasetManager(session), nil
}

// CreateDatasetManagerDirectWithOptions creates a dataset manager with additional options
func CreateDatasetManagerDirectWithOptions(host string, port int, user, password string, rejectUnauthorized bool, basePath string) (*ZOSMFDatasetManager, error) {
	session, err := profile.CreateSessionDirectWithOptions(host, port, user, password, rejectUnauthorized, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewDatasetManager(session), nil
}

// CreateSequentialDataset creates a sequential dataset with default parameters
func (dm *ZOSMFDatasetManager) CreateSequentialDataset(name string) error {
	request := &CreateDatasetRequest{
		Name: name,
		Type: DatasetTypeSequential,
		Space: Space{
			Primary:   10,
			Secondary: 5,
			Unit:      SpaceUnitTracks,
		},
		RecordFormat: RecordFormatVariable,
		RecordLength: RecordLength256,
		BlockSize:    BlockSize27920,
	}
	return dm.CreateDataset(request)
}

// CreatePartitionedDataset creates a partitioned dataset with default parameters
func (dm *ZOSMFDatasetManager) CreatePartitionedDataset(name string) error {
	request := &CreateDatasetRequest{
		Name: name,
		Type: DatasetTypePartitioned,
		Space: Space{
			Primary:   10,
			Secondary: 5,
			Unit:      SpaceUnitTracks,
			Directory: 5,
		},
		RecordFormat: RecordFormatVariable,
		RecordLength: RecordLength256,
		BlockSize:    BlockSize27920,
		Directory:    5,
	}
	return dm.CreateDataset(request)
}

// CreateDatasetWithOptions creates a dataset with custom parameters
func (dm *ZOSMFDatasetManager) CreateDatasetWithOptions(name string, datasetType DatasetType, space Space, recordFormat RecordFormat, recordLength RecordLength, blockSize BlockSize) error {
	request := &CreateDatasetRequest{
		Name:         name,
		Type:         datasetType,
		Space:        space,
		RecordFormat: recordFormat,
		RecordLength: recordLength,
		BlockSize:    blockSize,
	}
	if datasetType == DatasetTypePartitioned && space.Directory == 0 {
		request.Directory = 5 // Default directory blocks for partitioned datasets
	}
	return dm.CreateDataset(request)
}

// UploadText uploads text content to a dataset
func (dm *ZOSMFDatasetManager) UploadText(datasetName, content string) error {
	request := &UploadRequest{
		DatasetName: datasetName,
		Content:     content,
		Encoding:    "UTF-8",
		Replace:     true,
	}
	return dm.UploadContent(request)
}

// UploadTextToMember uploads text content to a member in a partitioned dataset
func (dm *ZOSMFDatasetManager) UploadTextToMember(datasetName, memberName, content string) error {
	request := &UploadRequest{
		DatasetName: datasetName,
		MemberName:  memberName,
		Content:     content,
		Encoding:    "UTF-8",
		Replace:     true,
	}
	return dm.UploadContent(request)
}

// DownloadText downloads text content from a dataset
func (dm *ZOSMFDatasetManager) DownloadText(datasetName string) (string, error) {
	request := &DownloadRequest{
		DatasetName: datasetName,
		Encoding:    "UTF-8",
	}
	return dm.DownloadContent(request)
}

// DownloadTextFromMember downloads text content from a member in a partitioned dataset
func (dm *ZOSMFDatasetManager) DownloadTextFromMember(datasetName, memberName string) (string, error) {
	request := &DownloadRequest{
		DatasetName: datasetName,
		MemberName:  memberName,
		Encoding:    "UTF-8",
	}
	return dm.DownloadContent(request)
}

// GetDatasetsByOwner gets datasets owned by a specific user
// Note: z/OSMF API doesn't support owner filtering directly, so we use name pattern
func (dm *ZOSMFDatasetManager) GetDatasetsByOwner(owner string, limit int) (*DatasetList, error) {
	// Use the owner as a high-level qualifier pattern (common convention)
	filter := &DatasetFilter{
		Name:  owner + ".*",
		Limit: limit,
	}
	return dm.ListDatasets(filter)
}

// GetDatasetsByType gets datasets of a specific type
func (dm *ZOSMFDatasetManager) GetDatasetsByType(datasetType string, limit int) (*DatasetList, error) {
	filter := &DatasetFilter{
		Type:  datasetType,
		Limit: limit,
	}
	return dm.ListDatasets(filter)
}

// GetDatasetsByName gets datasets matching a name pattern
func (dm *ZOSMFDatasetManager) GetDatasetsByName(namePattern string, limit int) (*DatasetList, error) {
	filter := &DatasetFilter{
		Name:  namePattern,
		Limit: limit,
	}
	return dm.ListDatasets(filter)
}

// ValidateDatasetName validates a dataset name according to z/OS naming conventions
func ValidateDatasetName(name string) error {
	if name == "" {
		return fmt.Errorf("dataset name cannot be empty")
	}

	// Check length (1-44 characters)
	if len(name) > 44 {
		return fmt.Errorf("dataset name cannot exceed 44 characters")
	}

	// Check for valid characters (A-Z, 0-9, @, #, $, -, .)
	validPattern := regexp.MustCompile(`^[A-Z@#$][A-Z0-9@#$.-]*$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("dataset name contains invalid characters")
	}

	// Check for consecutive periods
	if strings.Contains(name, "..") {
		return fmt.Errorf("dataset name cannot contain consecutive periods")
	}

	// Check for leading/trailing periods
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("dataset name cannot start or end with a period")
	}

	// Check for consecutive hyphens
	if strings.Contains(name, "--") {
		return fmt.Errorf("dataset name cannot contain consecutive hyphens")
	}

	return nil
}

// ValidateMemberName validates a member name according to z/OS naming conventions
func ValidateMemberName(name string) error {
	if name == "" {
		return fmt.Errorf("member name cannot be empty")
	}

	// Check length (1-8 characters)
	if len(name) > 8 {
		return fmt.Errorf("member name cannot exceed 8 characters")
	}

	// Check for valid characters (A-Z, 0-9, @, #, $, -, .)
	validPattern := regexp.MustCompile(`^[A-Z@#$][A-Z0-9@#$.-]*$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("member name contains invalid characters")
	}

	// Check for consecutive periods
	if strings.Contains(name, "..") {
		return fmt.Errorf("member name cannot contain consecutive periods")
	}

	// Check for leading/trailing periods
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("member name cannot start or end with a period")
	}

	return nil
}

// ValidateCreateDatasetRequest validates a create dataset request
func ValidateCreateDatasetRequest(request *CreateDatasetRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate dataset name
	if err := ValidateDatasetName(request.Name); err != nil {
		return fmt.Errorf("invalid dataset name: %w", err)
	}

	// Validate dataset type
	switch request.Type {
	case DatasetTypeSequential, DatasetTypePartitioned, DatasetTypePDSE, DatasetTypeVSAM:
		// Valid types
	default:
		return fmt.Errorf("invalid dataset type: %s", request.Type)
	}

	// Validate space allocation
	if request.Space.Primary <= 0 {
		return fmt.Errorf("primary space allocation must be greater than 0")
	}
	if request.Space.Secondary < 0 {
		return fmt.Errorf("secondary space allocation cannot be negative")
	}

	// Validate space unit
	switch request.Space.Unit {
	case SpaceUnitTracks, SpaceUnitCylinders, SpaceUnitKB, SpaceUnitMB, SpaceUnitGB:
		// Valid units
	default:
		return fmt.Errorf("invalid space unit: %s", request.Space.Unit)
	}

	// Validate record format
	if request.RecordFormat != "" {
		switch request.RecordFormat {
		case RecordFormatFixed, RecordFormatVariable, RecordFormatUndefined:
			// Valid formats
		default:
			return fmt.Errorf("invalid record format: %s", request.RecordFormat)
		}
	}

	// Validate record length
	if request.RecordLength > 0 {
		if request.RecordLength < 1 || request.RecordLength > 32760 {
			return fmt.Errorf("record length must be between 1 and 32760")
		}
	}

	// Validate block size
	if request.BlockSize > 0 {
		if request.BlockSize < 1 || request.BlockSize > 32760 {
			return fmt.Errorf("block size must be between 1 and 32760")
		}
	}

	// Validate directory blocks for partitioned datasets
	if request.Type == DatasetTypePartitioned && request.Directory > 0 {
		if request.Directory < 1 || request.Directory > 9999 {
			return fmt.Errorf("directory blocks must be between 1 and 9999")
		}
	}

	return nil
}

// ValidateUploadRequest validates an upload request
func ValidateUploadRequest(request *UploadRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate dataset name
	if err := ValidateDatasetName(request.DatasetName); err != nil {
		return fmt.Errorf("invalid dataset name: %w", err)
	}

	// Validate member name if provided
	if request.MemberName != "" {
		if err := ValidateMemberName(request.MemberName); err != nil {
			return fmt.Errorf("invalid member name: %w", err)
		}
	}

	// Validate content
	if request.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	return nil
}

// ValidateDownloadRequest validates a download request
func ValidateDownloadRequest(request *DownloadRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate dataset name
	if err := ValidateDatasetName(request.DatasetName); err != nil {
		return fmt.Errorf("invalid dataset name: %w", err)
	}

	// Validate member name if provided
	if request.MemberName != "" {
		if err := ValidateMemberName(request.MemberName); err != nil {
			return fmt.Errorf("invalid member name: %w", err)
		}
	}

	return nil
}

// CreateDefaultSpace creates a default space allocation
func CreateDefaultSpace(unit SpaceUnit) Space {
	return Space{
		Primary:   10,
		Secondary: 5,
		Unit:      unit,
		Directory: 5, // For partitioned datasets
	}
}

// CreateLargeSpace creates a large space allocation
func CreateLargeSpace(unit SpaceUnit) Space {
	return Space{
		Primary:   100,
		Secondary: 50,
		Unit:      unit,
		Directory: 20, // For partitioned datasets
	}
}

// CreateSmallSpace creates a small space allocation
func CreateSmallSpace(unit SpaceUnit) Space {
	return Space{
		Primary:   5,
		Secondary: 2,
		Unit:      unit,
		Directory: 2, // For partitioned datasets
	}
}

// CopyMemberToSameDataset copies a member within the same partitioned dataset
func (dm *ZOSMFDatasetManager) CopyMemberToSameDataset(datasetName, sourceMember, targetMember string) error {
	return dm.CopyMember(datasetName, sourceMember, datasetName, targetMember)
}

// CopyMemberWithSameName copies a member from one dataset to another with the same member name
func (dm *ZOSMFDatasetManager) CopyMemberWithSameName(sourceDataset, targetDataset, memberName string) error {
	return dm.CopyMember(sourceDataset, memberName, targetDataset, memberName)
}