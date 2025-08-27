package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/datasets"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/jobs"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
	fmt.Println("=== Zowe Go SDK - zXplore GitHub Integration Test ===")
	fmt.Println()

	host := os.Getenv("ZXPLORE_HOST")
	port := os.Getenv("ZXPLORE_PORT")
	user := os.Getenv("ZXPLORE_USER")
	password := os.Getenv("ZXPLORE_PASSWORD")

	if host == "" || port == "" || user == "" || password == "" {
		log.Fatal("Missing required environment variables: ZXPLORE_HOST, ZXPLORE_PORT, ZXPLORE_USER, ZXPLORE_PASSWORD")
	}

	cfg := &profile.ZOSMFProfile{
		Name:               "zxplore",
		Host:               host,
		Port:               10443, // Default zXplore port
		Protocol:           "https",
		User:               user,
		Password:           password,
		RejectUnauthorized: false,
		// BasePath defaults to /zosmf via session if omitted
	}

	// Override port if environment variable is provided
	if port != "" {
		if _, err := fmt.Sscanf(port, "%d", &cfg.Port); err != nil {
			cfg.Port = 10443 // Fallback to default if parsing fails
		}
	}

	fmt.Printf("Connecting to %s://%s:%d%s as %s\n", cfg.Protocol, cfg.Host, cfg.Port, "/zosmf", cfg.User)

	// Create session
	sess, err := cfg.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// Use SDK managers only
	jm := jobs.NewJobManager(sess)
	dm := datasets.NewDatasetManager(sess)

	// Jobs check: list a few jobs (best-effort)
	fmt.Println("Listing jobs (best-effort)...")
	jl, err := jm.ListJobs(&jobs.JobFilter{MaxJobs: 5})
	if err != nil {
		fmt.Printf("Jobs list error: %v\n", err)
	} else {
		fmt.Printf("Jobs returned: %d\n", len(jl.Jobs))
	}

	// Datasets check: list datasets (best-effort)
	fmt.Println("Listing datasets (best-effort)...")
	dl, err := dm.ListDatasets(&datasets.DatasetFilter{Name: user + ".*", Limit: 10})
	if err != nil {
		fmt.Printf("Datasets list error: %v\n", err)
	} else {
		fmt.Printf("Datasets returned: %d\n", len(dl.Datasets))
		// Print first few dataset names for verification
		for i, ds := range dl.Datasets {
			if i >= 3 { // Show only first 3
				break
			}
			fmt.Printf("  Dataset %d: %s\n", i+1, ds.Name)
		}
	}

	// Close managers
	_ = jm.CloseJobManager()
	_ = dm.CloseDatasetManager()

	fmt.Println("=== GitHub integration test completed ===")
}
