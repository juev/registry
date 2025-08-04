package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const (
	// Nexus registry configuration
	registryHost = "localhost:8082"
	imageName    = "nexus-test"
	imageTag     = "latest"
	username     = "admin"
	password     = "admin123"
)

// createAuthenticator creates basic authentication for Nexus registry
func createAuthenticator() authn.Authenticator {
	return &authn.Basic{
		Username: username,
		Password: password,
	}
}

// createRemoteOptions creates common remote options for both functions
func createRemoteOptions() []remote.Option {
	return []remote.Option{
		remote.WithAuth(createAuthenticator()),
		remote.WithTransport(&http.Transport{
			// Enable detailed logging of HTTP requests
			DisableKeepAlives:     false,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			DisableCompression:    false,
			ResponseHeaderTimeout: 30 * time.Second,
		}),
	}
}

// logHTTPRequest logs details about HTTP requests for debugging
func logHTTPRequest(method, url string) {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	log.Printf("Making %s request to: %s", method, url)
	log.Printf("Using Basic Auth: %s", auth[:10]+"...")
}

// fetchImageUsingRemoteImage demonstrates using remote.Image() to fetch image information
// This approach fetches the complete image manifest and provides full image details
func fetchImageUsingRemoteImage() error {
	log.Println("=== Starting fetchImageUsingRemoteImage ===")
	
	// Parse the image reference
	imageRef := fmt.Sprintf("%s/%s:%s", registryHost, imageName, imageTag)
	log.Printf("Parsing image reference: %s", imageRef)
	
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}
	log.Printf("Parsed reference: %s", ref.String())

	// Log the expected HTTP request
	logHTTPRequest("GET", fmt.Sprintf("http://%s/v2/%s/manifests/%s", registryHost, imageName, imageTag))

	// Create context with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch the image using remote.Image()
	log.Println("Calling remote.Image() - this will make manifest and config requests")
	img, err := remote.Image(ref, createRemoteOptions()...)
	if err != nil {
		return fmt.Errorf("failed to fetch image using remote.Image(): %w", err)
	}

	// Get image manifest
	log.Println("Retrieving image manifest...")
	manifest, err := img.Manifest()
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}
	log.Printf("Manifest media type: %s", manifest.MediaType)
	log.Printf("Manifest schema version: %d", manifest.SchemaVersion)
	log.Printf("Number of layers: %d", len(manifest.Layers))

	// Get image config
	log.Println("Retrieving image config...")
	config, err := img.ConfigFile()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	log.Printf("Image architecture: %s", config.Architecture)
	log.Printf("Image OS: %s", config.OS)
	if len(config.Config.Env) > 0 {
		log.Printf("Environment variables: %d entries", len(config.Config.Env))
	}

	// Get image size
	size, err := img.Size()
	if err != nil {
		log.Printf("Warning: failed to get image size: %v", err)
	} else {
		log.Printf("Image size: %d bytes", size)
	}

	log.Println("=== Completed fetchImageUsingRemoteImage ===")
	return nil
}

// fetchImageUsingRemoteGet demonstrates using remote.Get() to retrieve specific information
// This is a lower-level approach that gives more control over what is fetched
func fetchImageUsingRemoteGet() error {
	log.Println("=== Starting fetchImageUsingRemoteGet ===")

	// Parse the image reference
	imageRef := fmt.Sprintf("%s/%s:%s", registryHost, imageName, imageTag)
	log.Printf("Parsing image reference: %s", imageRef)
	
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}
	log.Printf("Parsed reference: %s", ref.String())

	// Log the expected HTTP request
	logHTTPRequest("GET", fmt.Sprintf("http://%s/v2/%s/manifests/%s", registryHost, imageName, imageTag))

	// Create context with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use remote.Get() to fetch the descriptor and manifest
	log.Println("Calling remote.Get() - this makes a targeted manifest request")
	descriptor, err := remote.Get(ref, createRemoteOptions()...)
	if err != nil {
		return fmt.Errorf("failed to fetch descriptor using remote.Get(): %w", err)
	}

	log.Printf("Descriptor media type: %s", descriptor.MediaType)
	log.Printf("Descriptor size: %d bytes", descriptor.Size)
	log.Printf("Descriptor digest: %s", descriptor.Digest)

	// Parse the manifest from the descriptor
	log.Println("Parsing manifest from descriptor...")
	manifest, err := descriptor.Image()
	if err != nil {
		return fmt.Errorf("failed to get image from descriptor: %w", err)
	}

	// Get manifest details
	manifestData, err := manifest.Manifest()
	if err != nil {
		return fmt.Errorf("failed to get manifest data: %w", err)
	}
	log.Printf("Manifest layers count: %d", len(manifestData.Layers))
	
	// Show the difference - remote.Get() gives us the raw descriptor
	log.Printf("Raw manifest size from descriptor: %d bytes", len(descriptor.Manifest))
	
	// We can also inspect individual layers if needed
	if len(manifestData.Layers) > 0 {
		log.Printf("First layer digest: %s", manifestData.Layers[0].Digest)
		log.Printf("First layer size: %d bytes", manifestData.Layers[0].Size)
	}

	log.Println("=== Completed fetchImageUsingRemoteGet ===")
	return nil
}

// demonstrateTransportLevelDebugging shows how to use transport-level debugging
func demonstrateTransportLevelDebugging() {
	log.Println("=== Transport-level debugging information ===")
	
	log.Println("To see detailed HTTP traffic, you can use tools like:")
	log.Println("- tcpdump: sudo tcpdump -i lo0 port 8082")
	log.Println("- wireshark on localhost interface")
	log.Println("- Check Nexus request logs in the Nexus admin interface")
	log.Println("- Enable Go HTTP client debugging with GODEBUG=http2debug=1")
}

func main() {
	log.Println("Starting go-containerregistry demonstration with Nexus registry")
	log.Printf("Target registry: %s", registryHost)
	log.Printf("Target image: %s/%s:%s", registryHost, imageName, imageTag)
	log.Println("Authentication: admin/admin123")
	log.Println()

	// Show transport debugging info
	demonstrateTransportLevelDebugging()
	log.Println()

	log.Println("This application demonstrates two different approaches:")
	log.Println("1. remote.Image() - High-level API that fetches complete image information")
	log.Println("2. remote.Get() - Lower-level API for targeted manifest requests")
	log.Println()

	// Test approach 1: remote.Image()
	log.Println("🔍 Testing Approach 1: remote.Image()")
	if err := fetchImageUsingRemoteImage(); err != nil {
		log.Printf("❌ Error with remote.Image(): %v", err)
	} else {
		log.Println("✅ remote.Image() completed successfully")
	}
	log.Println()

	// Small delay between requests for cleaner logs
	time.Sleep(2 * time.Second)

	// Test approach 2: remote.Get()
	log.Println("🔍 Testing Approach 2: remote.Get()")
	if err := fetchImageUsingRemoteGet(); err != nil {
		log.Printf("❌ Error with remote.Get(): %v", err)
	} else {
		log.Println("✅ remote.Get() completed successfully")
	}

	log.Println()
	log.Println("✨ Demonstration completed!")
	log.Println()
	log.Println("Key differences observed:")
	log.Println("• remote.Image() makes multiple requests (manifest + config + potentially layers)")
	log.Println("• remote.Get() makes targeted requests for specific information")
	log.Println("• Both support the same authentication and transport options")
	log.Println("• Check your Nexus request logs to see the actual HTTP requests made")
}