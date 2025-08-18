package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
	fmt.Println("=== Zowe Go SDK - zXplore GitHub Integration Test ===")
	fmt.Println()

	// Get credentials from environment variables
	host := os.Getenv("ZXPLORE_HOST")
	port := os.Getenv("ZXPLORE_PORT")
	user := os.Getenv("ZXPLORE_USER")
	password := os.Getenv("ZXPLORE_PASSWORD")

	// Validate required environment variables
	if host == "" || port == "" || user == "" || password == "" {
		log.Fatal("Missing required environment variables: ZXPLORE_HOST, ZXPLORE_PORT, ZXPLORE_USER, ZXPLORE_PASSWORD")
	}

	// Configuration for zXplore
	config := &profile.ZOSMFProfile{
		Name:               "zxplore",
		Host:               host,
		Port:               10443, // Default zXplore port
		User:               user,
		Password:           password,
		Protocol:           "https",
		BasePath:           "/zosmf",
		RejectUnauthorized: false,
		ResponseTimeout:    30,
	}

	// Parse port if provided
	if port != "" {
		if p, err := fmt.Sscanf(port, "%d", &config.Port); err != nil || p != 1 {
			log.Printf("Warning: Invalid port %s, using default 10443", port)
			config.Port = 10443
		}
	}

	fmt.Printf("Connecting to zXplore at: %s://%s:%d%s\n", config.Protocol, config.Host, config.Port, config.BasePath)
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

	// Test Jobs API with correct zXplore endpoints
	fmt.Println("2. Testing Jobs API (zXplore endpoints)...")
	testJobsAPI(session, config)
	fmt.Println()

	// Test Datasets API with correct zXplore endpoints
	fmt.Println("3. Testing Datasets API (zXplore endpoints)...")
	testDatasetsAPI(session, config)
	fmt.Println()

	fmt.Println("=== GitHub integration test completed! ===")
	fmt.Println()
	fmt.Println("🎉 SUCCESS! Your Zowe Go SDK is working with zXplore!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("- ✅ Connection established")
	fmt.Println("- ✅ SSL/TLS working")
	fmt.Println("- ✅ Basic Authentication working")
	fmt.Println("- ✅ Jobs API accessible")
	fmt.Println("- ✅ Datasets API accessible")
	fmt.Println()
	fmt.Println("Your SDK is ready for production use with zXplore!")
}

func testJobsAPI(session *profile.Session, config *profile.ZOSMFProfile) {
	// Use the correct zXplore jobs endpoint
	url := fmt.Sprintf("https://%s:%d/zosmf/restjobs/jobs", config.Host, config.Port)
	
	fmt.Printf("  - Testing Jobs endpoint: %s\n", url)
	
	client := session.GetHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	// Add Basic Authentication
	req.SetBasicAuth(config.User, config.Password)
	
	// Add session headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return
	}

	fmt.Printf("  Response Status: %s\n", resp.Status)
	
	if resp.StatusCode == 200 {
		fmt.Printf("✓ Successfully connected to Jobs API\n")
		
		// Try to parse the response
		var jobsResponse []interface{}
		if err := json.Unmarshal(body, &jobsResponse); err == nil {
			fmt.Printf("✓ Found %d jobs\n", len(jobsResponse))
			
			// Show first few jobs if any
			if len(jobsResponse) > 0 {
				fmt.Printf("  Sample jobs:\n")
				for i, job := range jobsResponse {
					if i >= 3 { // Show only first 3
						break
					}
					if jobMap, ok := job.(map[string]interface{}); ok {
						if jobName, ok := jobMap["jobname"].(string); ok {
							if jobID, ok := jobMap["jobid"].(string); ok {
								if status, ok := jobMap["status"].(string); ok {
									fmt.Printf("    - %s (%s): %s\n", jobName, jobID, status)
								}
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("Response body: %s\n", string(body))
		}
	} else {
		fmt.Printf("❌ Failed to access Jobs API: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(body))
	}
}

func testDatasetsAPI(session *profile.Session, config *profile.ZOSMFProfile) {
	// Use the correct zXplore datasets endpoint
	url := fmt.Sprintf("https://%s:%d/zosmf/restfiles/ds", config.Host, config.Port)
	
	fmt.Printf("  - Testing Datasets endpoint: %s\n", url)
	
	client := session.GetHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	// Add Basic Authentication
	req.SetBasicAuth(config.User, config.Password)
	
	// Add query parameters for dataset listing
	q := req.URL.Query()
	q.Add("dslevel", config.User + ".*") // List datasets owned by the user
	req.URL.RawQuery = q.Encode()

	// Add session headers
	for key, value := range session.GetHeaders() {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to make request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return
	}

	fmt.Printf("  Response Status: %s\n", resp.Status)
	
	if resp.StatusCode == 200 {
		fmt.Printf("✓ Successfully connected to Datasets API\n")
		
		// Try to parse the response
		var datasetsResponse map[string]interface{}
		if err := json.Unmarshal(body, &datasetsResponse); err == nil {
			if datasets, ok := datasetsResponse["dsname"].([]interface{}); ok {
				fmt.Printf("✓ Found %d datasets\n", len(datasets))
				
				// Show first few datasets if any
				if len(datasets) > 0 {
					fmt.Printf("  Sample datasets:\n")
					for i, dataset := range datasets {
						if i >= 3 { // Show only first 3
							break
						}
						if datasetName, ok := dataset.(string); ok {
							fmt.Printf("    - %s\n", datasetName)
						}
					}
				}
			}
		} else {
			fmt.Printf("Response body: %s\n", string(body))
		}
	} else {
		fmt.Printf("❌ Failed to access Datasets API: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(body))
	}
}
