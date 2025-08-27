package main

import (
	"fmt"
	"log"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
)

func main() {
	fmt.Println("=== Dataset Information Example ===")
	
	// Create a profile for your z/OS system
	zosmfProfile := &profile.ZOSMFProfile{
		Host:     "your-zos-host",
		Port:     443,
		User:     "your-user",
		Password: "your-password",
		RejectUnauthorized: false,
	}

	// Create dataset manager
	dm, err := datasets.NewDatasetManagerFromProfile(zosmfProfile)
	if err != nil {
		log.Fatalf("Failed to create dataset manager: %v", err)
	}
	defer dm.CloseDatasetManager()

	// Example 1: Get information about a specific dataset
	datasetName := "YOUR.DATASET.NAME"
	fmt.Printf("Getting information for dataset: %s\n", datasetName)
	
	info, err := dm.GetDatasetInfo(datasetName)
	if err != nil {
		log.Printf("Failed to get dataset info: %v", err)
	} else {
		fmt.Printf("Dataset Information:\n")
		fmt.Printf("  Name: %s\n", info.Name)
		fmt.Printf("  Type: %s\n", info.Type)
		fmt.Printf("  Record Format: %s\n", info.RecordFormat)
		fmt.Printf("  Record Length: %s\n", info.RecordLength)
		fmt.Printf("  Block Size: %s\n", info.BlockSize)
		fmt.Printf("  Volume: %s\n", info.Volume)
		fmt.Printf("  Created Date: %s\n", info.CreatedDate)
		fmt.Printf("  Expiry Date: %s\n", info.ExpiryDate)
		fmt.Printf("  Used: %s%%\n", info.Used)
	}

	// Example 2: Compare with existing GetDataset method
	fmt.Printf("\nComparing with GetDataset method:\n")
	existingInfo, err := dm.GetDataset(datasetName)
	if err != nil {
		log.Printf("Failed to get dataset using GetDataset: %v", err)
	} else {
		fmt.Printf("GetDataset result:\n")
		fmt.Printf("  Name: %s\n", existingInfo.Name)
		fmt.Printf("  Type: %s\n", existingInfo.Type)
		fmt.Printf("  Record Format: %s\n", existingInfo.RecordFormat)
		fmt.Printf("  Record Length: %s\n", existingInfo.RecordLength)
		fmt.Printf("  Block Size: %s\n", existingInfo.BlockSize)
		fmt.Printf("  Volume: %s\n", existingInfo.Volume)
	}

	// Example 3: Error handling for non-existent dataset
	fmt.Printf("\nTesting error handling:\n")
	nonExistentDataset := "NONEXISTENT.DATASET"
	_, err = dm.GetDatasetInfo(nonExistentDataset)
	if err != nil {
		fmt.Printf("Correctly handled non-existent dataset: %v\n", err)
	} else {
		fmt.Printf("Unexpected success for non-existent dataset\n")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("The GetDatasetInfo method provides:")
	fmt.Println("- Direct API access for dataset metadata")
	fmt.Println("- Fallback to list API if direct access fails")
	fmt.Println("- Comprehensive error handling")
	fmt.Println("- Compatibility with different z/OSMF versions")
}
