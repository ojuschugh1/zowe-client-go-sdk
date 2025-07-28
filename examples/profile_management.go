package main

import (
	"fmt"
	"log"
	
	"github.com/ojuschugh1/zowe-client-go-sdk/pkg/profile"
)

func main() {
	fmt.Println("Zowe Go SDK - Profile Management Example")
	fmt.Println("========================================")

	// Example 1: Create a profile manager and load profiles from Zowe CLI config
	fmt.Println("\n1. Loading profiles from Zowe CLI configuration:")
	pm := profile.NewProfileManager()

	// List available profiles
	profiles, err := pm.ListZOSMFProfiles()
	if err != nil {
		log.Printf("Warning: Could not list profiles: %v", err)
		fmt.Println("   No profiles found or config file not accessible")
	} else {
		fmt.Printf("   Available profiles: %v\n", profiles)
	}

	// Try to get default profile
	defaultProfile, err := pm.GetDefaultZOSMFProfile()
	if err != nil {
		fmt.Println("   No default profile found")
	} else {
		fmt.Printf("   Default profile: %s (%s:%d)\n", defaultProfile.Name, defaultProfile.Host, defaultProfile.Port)
	}

	// Example 2: Create a profile programmatically
	fmt.Println("\n2. Creating a profile programmatically:")
	zosmfProfile := profile.CreateZOSMFProfile("example", "mainframe.example.com", 443, "myuser", "mypassword")
	fmt.Printf("   Created profile: %s\n", zosmfProfile.Name)
	fmt.Printf("   Host: %s:%d\n", zosmfProfile.Host, zosmfProfile.Port)
	fmt.Printf("   User: %s\n", zosmfProfile.User)

	// Example 3: Create a profile with custom options
	fmt.Println("\n3. Creating a profile with custom options:")
	customProfile := profile.CreateZOSMFProfileWithOptions(
		"custom",
		"dev-mainframe.example.com",
		8080,
		"devuser",
		"devpass",
		false, // Allow insecure connections
		"/api/v1",
	)
	fmt.Printf("   Created custom profile: %s\n", customProfile.Name)
	fmt.Printf("   Base URL: %s:%d%s\n", customProfile.Host, customProfile.Port, customProfile.BasePath)
	fmt.Printf("   Reject unauthorized: %t\n", customProfile.RejectUnauthorized)

	// Example 4: Create sessions from profiles
	fmt.Println("\n4. Creating sessions from profiles:")
	
	// Session from programmatic profile
	session1, err := zosmfProfile.NewSession()
	if err != nil {
		log.Printf("Error creating session: %v", err)
	} else {
		fmt.Printf("   Session 1 - Base URL: %s\n", session1.GetBaseURL())
		fmt.Printf("   Session 1 - Headers: %v\n", session1.GetHeaders())
	}

	// Session from custom profile
	session2, err := customProfile.NewSession()
	if err != nil {
		log.Printf("Error creating session: %v", err)
	} else {
		fmt.Printf("   Session 2 - Base URL: %s\n", session2.GetBaseURL())
		fmt.Printf("   Session 2 - Headers: %v\n", session2.GetHeaders())
	}

	// Example 5: Create session directly without profile
	fmt.Println("\n5. Creating session directly:")
	directSession, err := profile.CreateSessionDirect("direct.example.com", 443, "directuser", "directpass")
	if err != nil {
		log.Printf("Error creating direct session: %v", err)
	} else {
		fmt.Printf("   Direct session - Base URL: %s\n", directSession.GetBaseURL())
	}

	// Example 6: Session header management
	fmt.Println("\n6. Managing session headers:")
	if session1 != nil {
		session1.AddHeader("X-Custom-Header", "custom-value")
		session1.AddHeader("Authorization", "Bearer token123")
		fmt.Printf("   Added custom headers to session 1\n")
		fmt.Printf("   Headers: %v\n", session1.GetHeaders())

		session1.RemoveHeader("X-Custom-Header")
		fmt.Printf("   Removed X-Custom-Header\n")
		fmt.Printf("   Remaining headers: %v\n", session1.GetHeaders())
	}

	// Example 7: Profile validation
	fmt.Println("\n7. Profile validation:")
	invalidProfile := &profile.ZOSMFProfile{
		Name: "invalid",
		// Missing required fields
	}
	
	err = profile.ValidateProfile(invalidProfile)
	if err != nil {
		fmt.Printf("   Validation error: %v\n", err)
	}

	validProfile := &profile.ZOSMFProfile{
		Name:     "valid",
		Host:     "valid.example.com",
		Port:     443,
		User:     "validuser",
		Password: "validpass",
	}
	
	err = profile.ValidateProfile(validProfile)
	if err == nil {
		fmt.Printf("   Profile validation passed\n")
	}

	// Example 8: Profile cloning
	fmt.Println("\n8. Profile cloning:")
	original := profile.CreateZOSMFProfile("original", "orig.example.com", 443, "origuser", "origpass")
	cloned := profile.CloneProfile(original)
	cloned.Name = "cloned"
	cloned.Host = "cloned.example.com"
	
	fmt.Printf("   Original: %s (%s)\n", original.Name, original.Host)
	fmt.Printf("   Cloned: %s (%s)\n", cloned.Name, cloned.Host)

	fmt.Println("\nProfile management example completed!")
} 