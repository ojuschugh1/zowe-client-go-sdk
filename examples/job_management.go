package main

import (
	"fmt"
	"log"

	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/jobs"
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
	fmt.Println("Zowe Go SDK - Job Management Example")
	fmt.Println("====================================")

	// Example 1: Create a job manager from a profile
	fmt.Println("\n1. Creating job manager from profile:")
	pm := profile.NewProfileManager()

	// Try to get default profile
	defaultProfile, err := pm.GetDefaultZOSMFProfile()
	if err != nil {
		fmt.Println("   No default profile found, creating job manager directly")
		// Create job manager directly
		jm, err := jobs.CreateJobManagerDirect("localhost", 8080, "testuser", "testpass")
		if err != nil {
			log.Fatal(err)
		}
		demonstrateJobManagement(jm)
	} else {
		fmt.Printf("   Using default profile: %s\n", defaultProfile.Name)
		jm, err := jobs.CreateJobManager(pm, "default")
		if err != nil {
			log.Fatal(err)
		}
		demonstrateJobManagement(jm)
	}
}

func demonstrateJobManagement(jm *jobs.ZOSMFJobManager) {
	// Example 2: List jobs
	fmt.Println("\n2. Listing jobs:")
	jobList, err := jm.ListJobs(nil)
	if err != nil {
		fmt.Printf("   Error listing jobs: %v\n", err)
	} else {
		fmt.Printf("   Found %d jobs\n", len(jobList.Jobs))
		for _, job := range jobList.Jobs {
			fmt.Printf("   - %s (%s): %s\n", job.JobName, job.JobID, job.Status)
		}
	}

	// Example 3: List jobs with filter
	fmt.Println("\n3. Listing jobs with filter:")
	filter := &jobs.JobFilter{
		Owner:   "testuser",
		MaxJobs: 10,
	}
	jobList, err = jm.ListJobs(filter)
	if err != nil {
		fmt.Printf("   Error listing jobs with filter: %v\n", err)
	} else {
		fmt.Printf("   Found %d jobs for user 'testuser'\n", len(jobList.Jobs))
	}

	// Example 4: Submit a simple job
	fmt.Println("\n4. Submitting a simple job:")
	jobStatement := jobs.CreateSimpleJobStatement("TESTJOB", "ACCT", "USER", "A", "(1,1)")
	fmt.Printf("   Job statement: %s\n", jobStatement)

	// Note: This would actually submit the job if connected to a real mainframe
	// For demonstration, we'll just show the JCL
	jcl := jobStatement + "\n" +
		"//STEP1 EXEC PGM=IEFBR14\n" +
		"//DD1 DD DSN=TEST.DATA,DISP=SHR\n" +
		"//DD2 DD SYSOUT=A\n"

	fmt.Printf("   Complete JCL:\n%s\n", jcl)

	// Example 5: Submit job using convenience method
	fmt.Println("\n5. Submitting job using convenience method:")
	response, err := jm.SubmitJobStatement(jcl)
	if err != nil {
		fmt.Printf("   Error submitting job: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Job submitted successfully!\n")
		fmt.Printf("   Job ID: %s\n", response.JobID)
		fmt.Printf("   Job Name: %s\n", response.JobName)
		fmt.Printf("   Status: %s\n", response.Status)
	}

	// Example 6: Submit job from dataset
	fmt.Println("\n6. Submitting job from dataset:")
	response, err = jm.SubmitJobFromDataset("TEST.JCL", "")
	if err != nil {
		fmt.Printf("   Error submitting job from dataset: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Job submitted from dataset successfully!\n")
		fmt.Printf("   Job ID: %s\n", response.JobID)
	}

	// Example 7: Get job status
	fmt.Println("\n7. Getting job status:")
	// Use a sample job ID for demonstration
	jobID := "JOB001"
	status, err := jm.GetJobStatus(jobID)
	if err != nil {
		fmt.Printf("   Error getting job status: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Job %s status: %s\n", jobID, status)
	}

	// Example 8: Get job details
	fmt.Println("\n8. Getting job details:")
	job, err := jm.GetJob(jobID)
	if err != nil {
		fmt.Printf("   Error getting job details: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Job details:\n")
		fmt.Printf("   - Job ID: %s\n", job.JobID)
		fmt.Printf("   - Job Name: %s\n", job.JobName)
		fmt.Printf("   - Owner: %s\n", job.Owner)
		fmt.Printf("   - Status: %s\n", job.Status)
		if job.RetCode != "" {
			fmt.Printf("   - Return Code: %s\n", job.RetCode)
		}
	}

	// Example 9: Get spool files
	fmt.Println("\n9. Getting spool files:")
	// Use sample job name and ID for demonstration
	jobName := "TESTJOB"
	spoolFiles, err := jm.GetSpoolFiles(jobName, jobID)
	if err != nil {
		fmt.Printf("   Error getting spool files: %v\n", err)
		fmt.Println("   (This is expected if not connected to a real mainframe)")
	} else {
		fmt.Printf("   Found %d spool files:\n", len(spoolFiles))
		for _, spoolFile := range spoolFiles {
			fmt.Printf("   - %s (ID: %d, Records: %d, Bytes: %d)\n",
				spoolFile.DDName, spoolFile.ID, spoolFile.Records, spoolFile.Bytes)
		}
	}

	// Example 10: Get spool file content
	fmt.Println("\n10. Getting spool file content:")
	if len(spoolFiles) > 0 {
		content, err := jm.GetSpoolFileContent(jobName, jobID, spoolFiles[0].ID)
		if err != nil {
			fmt.Printf("   Error getting spool file content: %v\n", err)
		} else {
			fmt.Printf("   Content of %s:\n%s\n", spoolFiles[0].DDName, content)
		}
	}

	// Example 11: Get jobs by owner
	fmt.Println("\n11. Getting jobs by owner:")
	jobList, err = jm.GetJobsByOwner("testuser", 5)
	if err != nil {
		fmt.Printf("   Error getting jobs by owner: %v\n", err)
	} else {
		fmt.Printf("   Found %d jobs owned by 'testuser'\n", len(jobList.Jobs))
	}

	// Example 12: Get jobs by prefix
	fmt.Println("\n12. Getting jobs by prefix:")
	jobList, err = jm.GetJobsByPrefix("TEST", 5)
	if err != nil {
		fmt.Printf("   Error getting jobs by prefix: %v\n", err)
	} else {
		fmt.Printf("   Found %d jobs with prefix 'TEST'\n", len(jobList.Jobs))
	}

	// Example 13: Get jobs by status
	fmt.Println("\n13. Getting jobs by status:")
	jobList, err = jm.GetJobsByStatus("OUTPUT", 10)
	if err != nil {
		fmt.Printf("   Error getting jobs by status: %v\n", err)
	} else {
		fmt.Printf("   Found %d jobs with status 'OUTPUT'\n", len(jobList.Jobs))
	}

	// Example 14: Job validation
	fmt.Println("\n14. Job request validation:")
	validRequest := &jobs.SubmitJobRequest{
		JobStatement: "//TESTJOB JOB (ACCT),'USER',MSGCLASS=A",
	}
	err = jobs.ValidateJobRequest(validRequest)
	if err != nil {
		fmt.Printf("   Validation error: %v\n", err)
	} else {
		fmt.Println("   Job request is valid")
	}

	// Example 15: Create complex JCL
	fmt.Println("\n15. Creating complex JCL:")
	ddStatements := []string{
		"//DD1 DD DSN=TEST.DATA,DISP=SHR",
		"//DD2 DD SYSOUT=A",
		"//DD3 DD DSN=TEST.OUTPUT,DISP=(NEW,CATLG,DELETE)",
	}
	complexJCL := jobs.CreateJobWithStep("COMPLEX", "ACCT", "USER", "A", "(1,1)", "STEP1", "IEFBR14", ddStatements)
	fmt.Printf("   Complex JCL:\n%s\n", complexJCL)

	// Example 16: Wait for job completion (demonstration)
	fmt.Println("\n16. Waiting for job completion:")
	fmt.Println("   (This would wait for a job to complete if connected to a real mainframe)")
	fmt.Println("   For demonstration, showing the timeout and polling logic:")
	
	// This would be used in a real scenario:
	// status, err := jm.WaitForJobCompletion("JOB001", 5*time.Minute, 10*time.Second)
	// if err != nil {
	//     fmt.Printf("   Error waiting for job completion: %v\n", err)
	// } else {
	//     fmt.Printf("   Job completed with status: %s\n", status)
	// }

	// Example 17: Close job manager
	fmt.Println("\n17. Closing job manager:")
	err = jm.CloseJobManager()
	if err != nil {
		fmt.Printf("   Error closing job manager: %v\n", err)
	} else {
		fmt.Println("   Job manager closed successfully")
		fmt.Println("   HTTP connections have been cleaned up")
	}

	fmt.Println("\nJob Management Examples Complete!")
	fmt.Println("Note: Most operations require a connection to a real z/OS mainframe")
	fmt.Println("The examples above demonstrate the API structure and usage patterns.")
}
