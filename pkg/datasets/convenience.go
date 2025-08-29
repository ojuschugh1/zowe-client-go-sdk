package datasets

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

// CreateDatasetManager creates a dataset manager from a profile manager
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

// CreateDatasetManagerDirect creates a dataset manager with connection details
func CreateDatasetManagerDirect(host string, port int, user, password string) (*ZOSMFDatasetManager, error) {
	session, err := profile.CreateSessionDirect(host, port, user, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewDatasetManager(session), nil
}

// CreateDatasetManagerDirectWithOptions creates a dataset manager with extra options
func CreateDatasetManagerDirectWithOptions(host string, port int, user, password string, rejectUnauthorized bool, basePath string) (*ZOSMFDatasetManager, error) {
	session, err := profile.CreateSessionDirectWithOptions(host, port, user, password, rejectUnauthorized, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return NewDatasetManager(session), nil
}

// CreateSequentialDataset creates a sequential dataset with defaults
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

// CreatePartitionedDataset creates a partitioned dataset with defaults
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

// CreateDatasetWithOptions creates a dataset with custom settings
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
		request.Directory = 5 // Default directory blocks for PDS
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
	// Basic validation
	if err := ValidateMemberName(memberName); err != nil {
		return fmt.Errorf("invalid member name: %w", err)
	}

	// Create the upload request
	request := &UploadRequest{
		DatasetName: datasetName,
		MemberName:  memberName,
		Content:     content,
		Encoding:    "UTF-8",
		Replace:     true,
	}

	// Try the upload with enhanced error handling
	err := dm.UploadContent(request)
	if err != nil {
		// Provide specific guidance for common PDS errors
		if strings.Contains(err.Error(), "ISRZ002") || strings.Contains(err.Error(), "I/O error") {
			return fmt.Errorf("PDS directory I/O error for member %s in %s: %w. This typically indicates:\n1. Directory corruption - use ISPF 3.1 or IEBCOPY to repair\n2. Insufficient directory space - reallocate PDS with more directory blocks\n3. Member name conflicts - check for duplicate or invalid names", memberName, datasetName, err)
		}
		if strings.Contains(err.Error(), "LMFIND error") {
			return fmt.Errorf("PDS directory search error for member %s in %s: %w. The PDS directory may need maintenance using ISPF utilities", memberName, datasetName, err)
		}
		return err
	}

	return nil
}

// UploadTextToMemberWithValidation uploads text content to a member with comprehensive validation and retry logic
func (dm *ZOSMFDatasetManager) UploadTextToMemberWithValidation(datasetName, memberName, content string) error {
	// First, validate the member name according to z/OS standards
	if err := ValidateMemberName(memberName); err != nil {
		return fmt.Errorf("invalid member name: %w", err)
	}

	// Check if the PDS exists and get its information
	exists, err := dm.Exists(datasetName)
	if err != nil {
		return fmt.Errorf("failed to check dataset existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("dataset %s does not exist", datasetName)
	}

	// Get dataset information to verify it's a PDS
	dsInfo, err := dm.GetDataset(datasetName)
	if err != nil {
		return fmt.Errorf("failed to get dataset information: %w", err)
	}
	
	// Verify it's a partitioned dataset
	if dsInfo.Type != "PO" && dsInfo.Type != "PO-E" {
		return fmt.Errorf("dataset %s is not a partitioned dataset (type: %s)", datasetName, dsInfo.Type)
	}

	// Try to list members first to ensure directory is accessible
	_, err = dm.ListMembers(datasetName)
	if err != nil {
		// If we can't list members, the directory might be corrupted
		return fmt.Errorf("PDS directory error for %s: %w. This may indicate directory corruption or insufficient directory space. Consider using IEBCOPY or ISPF to repair the PDS directory", datasetName, err)
	}

	// Check if member already exists - if so, we're replacing it
	memberExists := false
	members, err := dm.ListMembers(datasetName)
	if err == nil {
		for _, member := range members.Members {
			if member.Name == memberName {
				memberExists = true
				break
			}
		}
	}

	// Create the upload request with enhanced error handling
	request := &UploadRequest{
		DatasetName: datasetName,
		MemberName:  memberName,
		Content:     content,
		Encoding:    "UTF-8",
		Replace:     true,
	}

	// Attempt the upload with retry logic for directory issues
	err = dm.uploadWithRetry(request, memberExists)
	if err != nil {
		// Provide specific guidance for common PDS errors
		if strings.Contains(err.Error(), "ISRZ002") || strings.Contains(err.Error(), "I/O error") {
			return fmt.Errorf("PDS directory I/O error for member %s in %s: %w. This typically indicates:\n1. Directory corruption - use ISPF 3.1 or IEBCOPY to repair\n2. Insufficient directory space - reallocate PDS with more directory blocks\n3. Member name conflicts - check for duplicate or invalid names", memberName, datasetName, err)
		}
		if strings.Contains(err.Error(), "LMFIND error") {
			return fmt.Errorf("PDS directory search error for member %s in %s: %w. The PDS directory may need maintenance using ISPF utilities", memberName, datasetName, err)
		}
		return fmt.Errorf("failed to upload member %s to %s: %w", memberName, datasetName, err)
	}

	return nil
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

// uploadWithRetry attempts to upload content with retry logic for PDS directory issues
func (dm *ZOSMFDatasetManager) uploadWithRetry(request *UploadRequest, memberExists bool) error {
	const maxRetries = 3
	const retryDelay = time.Second * 2

	var lastError error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := dm.UploadContent(request)
		if err == nil {
			return nil // Success
		}
		
		lastError = err
		
		// Check if this is a retryable error
		errorStr := strings.ToLower(err.Error())
		isRetryable := strings.Contains(errorStr, "isrz002") ||
			strings.Contains(errorStr, "i/o error") ||
			strings.Contains(errorStr, "lmfind error") ||
			strings.Contains(errorStr, "directory") ||
			strings.Contains(errorStr, "timeout") ||
			strings.Contains(errorStr, "connection")
		
		if !isRetryable {
			// Non-retryable error, fail immediately
			return err
		}
		
		if attempt < maxRetries {
			// Wait before retry, with exponential backoff
			waitTime := time.Duration(attempt) * retryDelay
			time.Sleep(waitTime)
		}
	}
	
	return fmt.Errorf("upload failed after %d attempts: %w", maxRetries, lastError)
}

// CreatePDSWithDirectorySpace creates a PDS with adequate directory space to avoid I/O errors
func (dm *ZOSMFDatasetManager) CreatePDSWithDirectorySpace(name string, directoryBlocks int) error {
	if directoryBlocks < 5 {
		directoryBlocks = 10 // Minimum recommended directory blocks
	}
	
	request := &CreateDatasetRequest{
		Name: name,
		Type: DatasetTypePartitioned,
		Space: Space{
			Primary:   20, // Larger primary space
			Secondary: 10,
			Unit:      SpaceUnitTracks,
			Directory: directoryBlocks,
		},
		RecordFormat: RecordFormatVariable,
		RecordLength: RecordLength256,
		BlockSize:    BlockSize27920,
		Directory:    directoryBlocks,
	}
	return dm.CreateDataset(request)
}

// CheckPDSDirectoryHealth checks if a PDS directory is healthy and accessible
func (dm *ZOSMFDatasetManager) CheckPDSDirectoryHealth(datasetName string) error {
	// First check if dataset exists
	exists, err := dm.Exists(datasetName)
	if err != nil {
		return fmt.Errorf("failed to check dataset existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("dataset %s does not exist", datasetName)
	}

	// Get dataset information
	dsInfo, err := dm.GetDataset(datasetName)
	if err != nil {
		return fmt.Errorf("failed to get dataset information: %w", err)
	}
	
	// Verify it's a PDS
	if dsInfo.Type != "PO" && dsInfo.Type != "PO-E" {
		return fmt.Errorf("dataset %s is not a partitioned dataset (type: %s)", datasetName, dsInfo.Type)
	}

	// Try to list members to test directory accessibility
	_, err = dm.ListMembers(datasetName)
	if err != nil {
		return fmt.Errorf("PDS directory is not accessible: %w", err)
	}

	return nil
}