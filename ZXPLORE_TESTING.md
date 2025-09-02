# Testing Zowe Go SDK with zXplore Platform

This guide will help you test the Zowe Go SDK locally with the zXplore platform.

## Prerequisites

1. **zXplore Access**: You need access to the zXplore platform
2. **zXplore Credentials**: Your username and password for zXplore
3. **zXplore Connection Details**: Host, port, and protocol information

## Step 1: Get Your zXplore Connection Details

Contact your zXplore administrator or check your zXplore welcome email for:
- **Host**: The zXplore server hostname (e.g., `zxplore.ibm.com`)
- **Port**: Usually 443 for HTTPS
- **Protocol**: Usually HTTPS
- **Base Path**: Usually `/zosmf`
- **Username**: Your zXplore username
- **Password**: Your zXplore password

## Step 2: Configure Your Connection

### Option A: Using Environment Variables (Recommended for testing)

Set these environment variables in your terminal:

**Windows PowerShell:**
```powershell
$env:ZXPLORE_HOST = "your-zxplore-host.com"
$env:ZXPLORE_PORT = "443"
$env:ZXPLORE_USER = "your-username"
$env:ZXPLORE_PASSWORD = "your-password"
```

**Windows Command Prompt:**
```cmd
set ZXPLORE_HOST=your-zxplore-host.com
set ZXPLORE_PORT=443
set ZXPLORE_USER=your-username
set ZXPLORE_PASSWORD=your-password
```

**Linux/macOS:**
```bash
export ZXPLORE_HOST="your-zxplore-host.com"
export ZXPLORE_PORT="443"
export ZXPLORE_USER="your-username"
export ZXPLORE_PASSWORD="your-password"
```

### Option B: Using Zowe CLI Configuration File

1. Edit the `zowe.config.json` file in this directory
2. Replace the placeholder values with your actual zXplore credentials:

```json
{
  "profiles": {
    "zxplore": {
      "type": "zosmf",
      "properties": {
        "host": "your-actual-zxplore-host.com",
        "port": 443,
        "user": "your-actual-username",
        "password": "your-actual-password",
        "protocol": "https",
        "basePath": "/zosmf",
        "rejectUnauthorized": false,
        "responseTimeout": 30
      }
    }
  }
}
```

## Step 3: Run the Tests

### Basic Connectivity Test (Recommended first)

This test only checks if you can connect and list jobs/datasets:

```bash
go run test_with_zowe_config.go
```

### Full Feature Test

This test performs comprehensive testing including job submission and dataset creation:

```bash
go run test_zxplore.go
```

## Step 4: Understanding the Test Output

### Successful Output Example:
```
=== Zowe Go SDK - zXplore Local Testing ===

Connecting to zXplore at: https://your-zxplore-host.com:443/zosmf
User: your-username

1. Creating session...
✓ Session created successfully

2. Testing Profile Management...
✓ Profile validation passed
✓ Profile cloning works

3. Testing Jobs API...
  - Listing jobs...
✓ Found 5 jobs
  - Submitting test job...
✓ Job submitted successfully: JOB00000001 (TESTJOB)
  - Waiting for job to complete...
  - Getting job status...
✓ Job status: OUTPUT
  - Getting spool files...
✓ Found 3 spool files
  - Getting content of spool file 1 (JESMSGLG)...
✓ Spool file content length: 150 characters
  - Cleaning up test job...
✓ Test job deleted

4. Testing Datasets API...
  - Testing with dataset: YOURUSER.TEST.20231201123456
  - Listing datasets...
✓ Found 10 datasets
  - Creating test dataset...
✓ Test dataset created
  - Verifying dataset exists...
✓ Dataset exists
  - Uploading content to dataset...
✓ Content uploaded
  - Downloading content from dataset...
✓ Content downloaded (89 characters)
  - Getting dataset information...
✓ Dataset info: Name=YOURUSER.TEST.20231201123456, Type=SEQ, Size=89
  - Cleaning up test dataset...
✓ Test dataset deleted

=== All tests completed successfully! ===
```

### Common Error Messages and Solutions:

1. **"Failed to create session: x509: certificate signed by unknown authority"**
   - Solution: The `RejectUnauthorized: false` setting should handle this
   - If still failing, check if zXplore uses a different certificate setup

2. **"Failed to list jobs: API request failed with status 401"**
   - Solution: Check your username and password
   - Verify your zXplore account is active

3. **"Failed to list jobs: API request failed with status 403"**
   - Solution: Your user might not have permission to list jobs
   - Contact zXplore administrator for proper permissions

4. **"Failed to create dataset: API request failed with status 403"**
   - Solution: Your user might not have permission to create datasets
   - Check if you have proper dataset creation permissions

5. **"Failed to submit job: API request failed with status 400"**
   - Solution: The JCL might need adjustment for zXplore
   - Check zXplore-specific JCL requirements

## Step 5: Troubleshooting

### SSL/TLS Issues
If you encounter SSL certificate issues:
1. Make sure `RejectUnauthorized` is set to `false`
2. Check if zXplore uses a different port for HTTP (not HTTPS)
3. Verify the hostname is correct

### Authentication Issues
1. Double-check your username and password
2. Ensure your zXplore account is not locked
3. Try logging into zXplore web interface first to verify credentials

### Permission Issues
1. Contact zXplore administrator for proper permissions
2. Ask for permissions to:
   - List and submit jobs
   - Create and manage datasets
   - Access z/OSMF REST APIs

### Network Issues
1. Check if you can ping the zXplore host
2. Verify firewall settings
3. Try accessing zXplore web interface from the same machine

## Step 6: GitHub Actions Setup

Once local testing is successful, you can set up automated testing with GitHub Actions:

### 6.1: Create GitHub Secrets

1. Go to your repository Settings → Secrets and variables → Actions
2. Add the following secrets:
   - `ZXPLORE_HOST`: Your zXplore host IP (e.g., `204.90.115.200`)
   - `ZXPLORE_PORT`: Your zXplore port (usually `10443`)
   - `ZXPLORE_USER`: Your zXplore username (e.g., `Z74442`)
   - `ZXPLORE_PASSWORD`: Your zXplore password

### 6.2: Create GitHub Actions Workflow

Create `.github/workflows/zxplore-test.yml`:

```yaml
name: zXplore Integration Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Test zXplore Integration
      env:
        ZXPLORE_HOST: ${{ secrets.ZXPLORE_HOST }}
        ZXPLORE_PORT: ${{ secrets.ZXPLORE_PORT }}
        ZXPLORE_USER: ${{ secrets.ZXPLORE_USER }}
        ZXPLORE_PASSWORD: ${{ secrets.ZXPLORE_PASSWORD }}
      run: go run test_zxplore_github.go
```

### 6.3: Run the Tests

1. Commit and push your code to trigger the workflow
2. Check the Actions tab to see test results
3. Monitor the logs for any issues


## Security Notes

- Never commit credentials to version control
- Use environment variables or GitHub secrets for sensitive data
- The test files create temporary jobs and datasets that are cleaned up
- All test data uses your user ID to avoid conflicts

## Support

If you encounter issues: 
1. Check the zXplore documentation -: https://www.ibm.com/docs/en/zos
2. Contact zXplore support
3. Review the SDK documentation
4. Check the test output for specific error messages
