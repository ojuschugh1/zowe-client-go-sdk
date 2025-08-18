package main

import (
	"fmt"
	"log"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/jobs"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
	fmt.Println("=== Zowe Go SDK - Testing with Zowe CLI Config ===")
	fmt.Println()

	// Create profile manager to read Zowe CLI config
	profileManager := &profile.ZOSMFProfileManager{}

	// Get the zxplore profile from Zowe CLI config
	config, err := profileManager.GetZOSMFProfile("zxplore")
	if err != nil {
		log.Fatalf("Failed to get zxplore profile: %v", err)
	}

	fmt.Printf("Using profile: %s\n", config.Name)
	fmt.Printf("Host: %s:%d\n", config.Host, config.Port)
	fmt.Printf("Protocol: %s\n", config.Protocol)
	fmt.Printf("Base Path: %s\n", config.BasePath)
	fmt.Printf("User: %s\n", config.User)
	fmt.Println()

	// Create session
	fmt.Println("1. Creating session...")
	session, err := config.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	fmt.Println("✓ Session created successfully")
	fmt.Println()

	// Test basic connectivity
	fmt.Println("2. Testing basic connectivity...")
	
	// Test Jobs API - just list jobs
	jobManager := jobs.NewJobManager(session)
	defer jobManager.CloseJobManager()

	jobList, err := jobManager.ListJobs(&jobs.JobFilter{
		MaxJobs: 5,
	})
	if err != nil {
		log.Printf("Failed to list jobs: %v", err)
	} else {
		fmt.Printf("✓ Successfully listed %d jobs\n", len(jobList.Jobs))
	}

	// Test Datasets API - just list datasets
	datasetManager := datasets.NewDatasetManager(session)
	defer datasetManager.CloseDatasetManager()

	datasetList, err := datasetManager.ListDatasets(&datasets.DatasetFilter{
		Limit: 5,
	})
	if err != nil {
		log.Printf("Failed to list datasets: %v", err)
	} else {
		fmt.Printf("✓ Successfully listed %d datasets\n", len(datasetList.Datasets))
	}

	fmt.Println()
	fmt.Println("=== Basic connectivity test completed! ===")
	fmt.Println("If you see the checkmarks above, your connection to zXplore is working!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Update the zowe.config.json file with your actual zXplore credentials")
	fmt.Println("2. Run: go run test_zxplore.go (for full testing)")
	fmt.Println("3. Or run: go run test_with_zowe_config.go (for basic connectivity)")
}
