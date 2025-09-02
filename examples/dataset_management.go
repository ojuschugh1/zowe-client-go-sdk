package main

import (
	"fmt"
	"log"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
	fmt.Println("Zowe Go SDK - Dataset Management Example")
	fmt.Println("=========================================")

	// Example 1: Create a dataset manager from a profile
	fmt.Println("\n1. Creating dataset manager from profile:")
	pm := profile.NewProfileManager()

	// Try to get default profile
	defaultProfile, err := pm.GetDefaultZOSMFProfile()
	if err != nil {
		fmt.Println("   No default profile found, creating dataset manager directly")
		// Create dataset manager directly
		dm, err := datasets.CreateDatasetManagerDirect("localhost", 8080, "testuser", "testpass")
		if err != nil {
			log.Fatal(err)
		}
		demonstrateDatasetManagement(dm)
	} else {
		fmt.Printf("   Using default profile: %s\n", defaultProfile.Name)
		dm, err := datasets.CreateDatasetManager(pm, "default")
		if err != nil {
			log.Fatal(err)
		}
		demonstrateDatasetManagement(dm)
	}
}

func demonstrateDatasetManagement(dm *datasets.ZOSMFDatasetManager) {
	// Example 2: List datasets
	fmt.Println("\n2. Listing datasets:")
	datasetList, err := dm.ListDatasets(nil)
	if err != nil {
		fmt.Printf("   Error listing datasets: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Found %d datasets\n", len(datasetList.Datasets))
		for _, dataset := range datasetList.Datasets {
					fmt.Printf("   - %s (%s): %s%% used, size: %s\n", 
			dataset.Name, dataset.Type, dataset.Used, dataset.SizeX)
		}
	}

	// Example 3: List datasets with filter
	fmt.Println("\n3. Listing datasets with filter:")
	filter := &datasets.DatasetFilter{
		Type:  "SEQ",
		Limit: 10,
	}
	datasetList, err = dm.ListDatasets(filter)
	if err != nil {
		fmt.Printf("   Error listing datasets with filter: %v\n", err)
	} else {
		fmt.Printf("   Found %d sequential datasets\n", len(datasetList.Datasets))
	}

	// Example 4: Create a sequential dataset
	fmt.Println("\n4. Creating a sequential dataset:")
	err = dm.CreateSequentialDataset("TEST.SEQ")
	if err != nil {
		fmt.Printf("   Error creating sequential dataset: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Sequential dataset created successfully!")
	}

	// Example 5: Create a partitioned dataset
	fmt.Println("\n5. Creating a partitioned dataset:")
	err = dm.CreatePartitionedDataset("TEST.PDS")
	if err != nil {
		fmt.Printf("   Error creating partitioned dataset: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Partitioned dataset created successfully!")
	}

	// Example 6: Create dataset with custom options
	fmt.Println("\n6. Creating dataset with custom options:")
	space := datasets.CreateLargeSpace(datasets.SpaceUnitCylinders)
	err = dm.CreateDatasetWithOptions("TEST.CUSTOM", datasets.DatasetTypeSequential, space, 
		datasets.RecordFormatFixed, datasets.RecordLength132, datasets.BlockSize27920)
	if err != nil {
		fmt.Printf("   Error creating custom dataset: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Custom dataset created successfully!")
	}

	// Example 7: Upload content to sequential dataset
	fmt.Println("\n7. Uploading content to sequential dataset:")
	content := "This is a test file content.\nIt contains multiple lines.\nEnd of file."
	err = dm.UploadText("TEST.SEQ", content)
	if err != nil {
		fmt.Printf("   Error uploading content: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Content uploaded successfully!")
	}

	// Example 8: Upload content to partitioned dataset member
	fmt.Println("\n8. Uploading content to partitioned dataset member:")
	memberContent := "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A\n//STEP1 EXEC PGM=IEFBR14\n//DD1 DD SYSOUT=A"
	err = dm.UploadTextToMember("TEST.PDS", "MEMBER1", memberContent)
	if err != nil {
		fmt.Printf("   Error uploading member content: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Member content uploaded successfully!")
	}

	// Example 9: Download content from dataset
	fmt.Println("\n9. Downloading content from dataset:")
	downloadedContent, err := dm.DownloadText("TEST.SEQ")
	if err != nil {
		fmt.Printf("   Error downloading content: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Downloaded content:\n%s\n", downloadedContent)
	}

	// Example 10: Download content from member
	fmt.Println("\n10. Downloading content from member:")
	memberContent, err = dm.DownloadTextFromMember("TEST.PDS", "MEMBER1")
	if err != nil {
		fmt.Printf("   Error downloading member content: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Downloaded member content:\n%s\n", memberContent)
	}

	// Example 11: List members in partitioned dataset
	fmt.Println("\n11. Listing members in partitioned dataset:")
	memberList, err := dm.ListMembers("TEST.PDS")
	if err != nil {
		fmt.Printf("   Error listing members: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Found %d members:\n", len(memberList.Members))
		for _, member := range memberList.Members {
			fmt.Printf("   - %s\n", member.Name)
		}
	}

	// Example 12: Get dataset information
	fmt.Println("\n12. Getting dataset information:")
	dataset, err := dm.GetDataset("TEST.SEQ")
	if err != nil {
		fmt.Printf("   Error getting dataset info: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Dataset: %s\n", dataset.Name)
		fmt.Printf("   Type: %s\n", dataset.Type)
		fmt.Printf("   Size: %s\n", dataset.SizeX)
		fmt.Printf("   Used: %s%%\n", dataset.Used)
	}

	// Example 13: Check if dataset exists
	fmt.Println("\n13. Checking if dataset exists:")
	exists, err := dm.Exists("TEST.SEQ")
	if err != nil {
		fmt.Printf("   Error checking existence: %v\n", err)
	} else {
		if exists {
			fmt.Println("   Dataset exists")
		} else {
			fmt.Println("   Dataset does not exist")
		}
	}

	// Example 14: Copy sequential dataset
	fmt.Println("\n14. Copying sequential dataset:")
	err = dm.CopySequentialDataset("TEST.SEQ", "TEST.COPY")
	if err != nil {
		fmt.Printf("   Error copying dataset: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Dataset copied successfully!")
	}

	// Example 15: Rename dataset
	fmt.Println("\n15. Renaming dataset:")
	err = dm.RenameDataset("TEST.COPY", "TEST.RENAMED")
	if err != nil {
		fmt.Printf("   Error renaming dataset: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Dataset renamed successfully!")
	}

	// Example 16: Delete member
	fmt.Println("\n16. Deleting member:")
	err = dm.DeleteMember("TEST.PDS", "MEMBER1")
	if err != nil {
		fmt.Printf("   Error deleting member: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Println("   Member deleted successfully!")
	}

	// Example 17: Delete datasets
	fmt.Println("\n17. Deleting datasets:")
	datasetsToDelete := []string{"TEST.SEQ", "TEST.PDS", "TEST.CUSTOM", "TEST.RENAMED"}
	for _, datasetName := range datasetsToDelete {
		err = dm.DeleteDataset(datasetName)
		if err != nil {
			fmt.Printf("   Error deleting %s: %v\n", datasetName, err)
		} else {
			fmt.Printf("   %s deleted successfully!\n", datasetName)
		}
	}

	// Example 18: Validation examples
	fmt.Println("\n18. Validation examples:")
	
	// Validate dataset names
	validNames := []string{"TEST.DATA", "USER.PROGRAM", "MY@DATA"}
	invalidNames := []string{"", "test.data", "123.DATA", "DATA..SET"}
	
	fmt.Println("   Valid dataset names:")
	for _, name := range validNames {
		err := datasets.ValidateDatasetName(name)
		if err != nil {
			fmt.Printf("   - %s: ERROR - %v\n", name, err)
		} else {
			fmt.Printf("   - %s: OK\n", name)
		}
	}
	
	fmt.Println("   Invalid dataset names:")
	for _, name := range invalidNames {
		err := datasets.ValidateDatasetName(name)
		if err != nil {
			fmt.Printf("   - %s: ERROR - %v\n", name, err)
		} else {
			fmt.Printf("   - %s: OK\n", name)
		}
	}

	// Validate member names
	validMembers := []string{"MEMBER1", "PROGRAM", "TEST@1"}
	invalidMembers := []string{"", "member1", "123MEMBER", "TOOLONG12"}
	
	fmt.Println("   Valid member names:")
	for _, name := range validMembers {
		err := datasets.ValidateMemberName(name)
		if err != nil {
			fmt.Printf("   - %s: ERROR - %v\n", name, err)
		} else {
			fmt.Printf("   - %s: OK\n", name)
		}
	}
	
	fmt.Println("   Invalid member names:")
	for _, name := range invalidMembers {
		err := datasets.ValidateMemberName(name)
		if err != nil {
			fmt.Printf("   - %s: ERROR - %v\n", name, err)
		} else {
			fmt.Printf("   - %s: OK\n", name)
		}
	}

	// Example 19: Space allocation examples
	fmt.Println("\n19. Space allocation examples:")
	
	defaultSpace := datasets.CreateDefaultSpace(datasets.SpaceUnitTracks)
	fmt.Printf("   Default space: %d/%d %s (dir: %d)\n", 
		defaultSpace.Primary, defaultSpace.Secondary, defaultSpace.Unit, defaultSpace.Directory)
	
	largeSpace := datasets.CreateLargeSpace(datasets.SpaceUnitCylinders)
	fmt.Printf("   Large space: %d/%d %s (dir: %d)\n", 
		largeSpace.Primary, largeSpace.Secondary, largeSpace.Unit, largeSpace.Directory)
	
	smallSpace := datasets.CreateSmallSpace(datasets.SpaceUnitKB)
	fmt.Printf("   Small space: %d/%d %s (dir: %d)\n", 
		smallSpace.Primary, smallSpace.Secondary, smallSpace.Unit, smallSpace.Directory)

	// Example 20: Close dataset manager
	fmt.Println("\n20. Closing dataset manager:")
	err = dm.CloseDatasetManager()
	if err != nil {
		fmt.Printf("   Error closing dataset manager: %v\n", err)
	} else {
		fmt.Println("   Dataset manager closed successfully")
		fmt.Println("   HTTP connections have been cleaned up")
	}

	fmt.Println("\nDataset Management Examples Complete!")
	fmt.Println("Note: Most operations require a connection to a real z/OS mainframe")
	fmt.Println("The examples above demonstrate the API structure and usage patterns.")
}
