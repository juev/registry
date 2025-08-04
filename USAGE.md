# Go Container Registry Demo

This application demonstrates two different approaches for interacting with a Nexus container registry using the `github.com/google/go-containerregistry` library.

## Prerequisites

1. **Nexus Registry Running**: Ensure your Nexus registry is running on `localhost:8082`
2. **Test Image Available**: Make sure the image `localhost:8082/nexus-test:latest` exists in your registry
3. **Authentication**: The application uses `admin/admin123` credentials

## Building and Running

### Build the application

```bash
go build
```

### Run the application

```bash
./registry
```

## What the Application Demonstrates

### 1. `fetchImageUsingRemoteImage()` Function

- Uses `remote.Image()` to fetch complete image information
- Makes multiple HTTP requests (manifest + config + potentially layers)
- Provides high-level access to image metadata
- Automatically handles image parsing and validation

**Key Features:**

- Retrieves complete image manifest
- Accesses image configuration details
- Gets layer information
- Calculates image size

### 2. `fetchImageUsingRemoteGet()` Function

- Uses `remote.Get()` for targeted manifest requests
- Lower-level API with more control over requests
- Makes fewer HTTP requests
- Ideal for specific metadata retrieval

**Key Features:**

- Fetches raw image descriptor
- Provides direct access to manifest data
- More efficient for targeted operations
- Lower overhead for specific use cases

## Monitoring Nexus Requests

To observe how different functions interact with Nexus, you can:

1. **Check Nexus Admin Logs**: Access the Nexus web interface and view request logs
2. **Use tcpdump**: `sudo tcpdump -i lo0 port 8082`
3. **Enable Go HTTP Debug**: Set `GODEBUG=http2debug=1` environment variable
4. **Use Wireshark**: Monitor localhost interface traffic

## Expected Output

The application will show:

- Detailed logging of each HTTP request being made
- Authentication information (masked for security)
- Image metadata retrieved by each approach
- Performance differences between the two methods

## Authentication

The application is configured with:

- **Registry**: `localhost:8082`
- **Username**: `admin`
- **Password**: `admin123`

Modify the constants in `main.go` if you need different credentials.

## Error Handling

The application includes comprehensive error handling for:

- Network connectivity issues
- Authentication failures
- Missing images or registries
- Timeout scenarios
- Invalid image references

## Customization

To test with different images or registries, modify these constants in `main.go`:

```go
const (
    registryHost = "localhost:8082"
    imageName    = "nexus-test"  
    imageTag     = "latest"
    username     = "admin"
    password     = "admin123"
)
```
