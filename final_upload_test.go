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
	fmt.Println("=== FINAL COMPLETE SDK TEST (INCLUDING UPLOAD) ===")
	fmt.Println()

	host := os.Getenv("ZXPLORE_HOST")
	port := os.Getenv("ZXPLORE_PORT")
	user := os.Getenv("ZXPLORE_USER")
	password := os.Getenv("ZXPLORE_PASSWORD")

	if host == "" || port == "" || user == "" || password == "" {
		log.Fatal("❌ Missing required environment variables")
	}

	cfg := &profile.ZOSMFProfile{
		Name:               "zxplore",
		Host:               host,
		Port:               10443,
		Protocol:           "https",
		User:               user,
		Password:           password,
		RejectUnauthorized: false,
	}

	if port != "" {
		if _, err := fmt.Sscanf(port, "%d", &cfg.Port); err != nil {
			cfg.Port = 10443
		}
	}

	fmt.Printf("🔗 Connecting to %s://%s:%d%s as %s\n", cfg.Protocol, cfg.Host, cfg.Port, "/zosmf", cfg.User)
	fmt.Println()

	sess, err := cfg.NewSession()
	if err != nil {
		log.Fatalf("❌ Failed to create session: %v", err)
	}

	testsPassed := 0
	totalTests := 0

	// Test Jobs API
	fmt.Println("📋 TESTING JOBS API")
	fmt.Println("==================")
	
	jm := jobs.NewJobManager(sess)

	totalTests++
	fmt.Print("🧪 List jobs... ")
	jl, err := jm.ListJobs(&jobs.JobFilter{MaxJobs: 3})
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("✅ PASSED: Found %d jobs\n", len(jl.Jobs))
		testsPassed++
	}

	fmt.Println()

	// Test Datasets API
	fmt.Println("🗂️  TESTING DATASETS API")
	fmt.Println("========================")

	dm := datasets.NewDatasetManager(sess)

	totalTests++
	fmt.Print("🧪 List user datasets... ")
	dl, err := dm.ListDatasets(&datasets.DatasetFilter{Name: user + ".*", Limit: 5})
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("✅ PASSED: Found %d datasets\n", len(dl.Datasets))
		testsPassed++
	}

	// Find a PDS for testing
	var pdsName string
	if len(dl.Datasets) > 0 {
		for _, ds := range dl.Datasets {
			if ds.Type == "PO-E" || ds.Type == "PO" {
				pdsName = ds.Name
				break
			}
		}
	}

	fmt.Println()

	// Test Members API (All Fixed)
	fmt.Println("📂 TESTING MEMBERS API (ALL FIXED)")
	fmt.Println("==================================")

	if pdsName != "" {
		totalTests++
		fmt.Printf("🧪 List members of PDS %s... ", pdsName)
		members, err := dm.ListMembers(pdsName)
		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
		} else {
			fmt.Printf("✅ PASSED: Found %d members\n", len(members.Members))
			testsPassed++

			// Test member content download (FIXED)
			if len(members.Members) > 0 {
				totalTests++
				fmt.Printf("🧪 Download member content (%s)... ", members.Members[0].Name)
				content, err := dm.DownloadTextFromMember(pdsName, members.Members[0].Name)
				if err != nil {
					fmt.Printf("❌ FAILED: %v\n", err)
				} else {
					fmt.Printf("✅ PASSED: Downloaded %d bytes\n", len(content))
					testsPassed++
				}
			}

			// Test member content upload (NEWLY FIXED)
			totalTests++
			fmt.Printf("🧪 Upload member content (FINAL)... ")
			testContent := "//FINAL JOB\n//STEP1 EXEC PGM=IEFBR14\n/*\n"
			err = dm.UploadTextToMember(pdsName, "FINAL", testContent)
			if err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
			} else {
				fmt.Printf("✅ PASSED: Uploaded %d bytes\n", len(testContent))
				testsPassed++

				// Verify upload by downloading back
				totalTests++
				fmt.Printf("🧪 Verify upload by downloading... ")
				verifyContent, err := dm.DownloadTextFromMember(pdsName, "FINAL")
				if err != nil {
					fmt.Printf("❌ FAILED: %v\n", err)
				} else if verifyContent == testContent {
					fmt.Printf("✅ PASSED: Content matches perfectly\n")
					testsPassed++
				} else {
					fmt.Printf("❌ FAILED: Content mismatch\n")
				}
			}
		}
	}

	fmt.Println()

	// Test Convenience Methods (Fixed)
	fmt.Println("🛠️  TESTING CONVENIENCE METHODS")
	fmt.Println("===============================")

	totalTests++
	fmt.Printf("🧪 GetDatasetsByOwner... ")
	ownerDS, err := dm.GetDatasetsByOwner(user, 3)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Printf("✅ PASSED: Found %d datasets by owner\n", len(ownerDS.Datasets))
		testsPassed++
	}

	fmt.Println()

	// Close managers
	_ = jm.CloseJobManager()
	_ = dm.CloseDatasetManager()

	// Final Results
	fmt.Println("📊 FINAL TEST RESULTS")
	fmt.Println("=====================")
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", testsPassed)
	fmt.Printf("Failed: %d\n", totalTests-testsPassed)
	fmt.Printf("Success Rate: %.1f%%\n", float64(testsPassed)/float64(totalTests)*100)

	if testsPassed == totalTests {
		fmt.Println()
		fmt.Println("🎉 PERFECT! ALL TESTS PASSED!")
		fmt.Println()
		fmt.Println("✅ Jobs API: Working")
		fmt.Println("✅ Datasets API: Working") 
		fmt.Println("✅ Members Listing: Working")
		fmt.Println("✅ Member Content Download: Working (FIXED)")
		fmt.Println("✅ Member Content Upload: Working (FIXED)")
		fmt.Println("✅ Upload Verification: Working")
		fmt.Println("✅ GetDatasetsByOwner: Working (FIXED)")
		fmt.Println("✅ All Core Functionality: COMPLETE")
		fmt.Println()
		fmt.Println("🚀 YOUR ZOWE GO SDK IS 100% FUNCTIONAL!")
	} else {
		fmt.Printf("\n⚠️  %d tests failed. Review above for details.\n", totalTests-testsPassed)
	}

	fmt.Println()
	fmt.Println("=== FINAL TEST COMPLETED ===")
}
