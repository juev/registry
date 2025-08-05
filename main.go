package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
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

// HTTPRequestMetrics holds detailed information about each HTTP request
type HTTPRequestMetrics struct {
	Method       string
	URL          string
	Headers      map[string][]string
	RequestSize  int64
	ResponseSize int64
	StatusCode   int
	Duration     time.Duration
	Timestamp    time.Time
	Error        error
	FunctionName string
}

// HTTPMonitor collects and manages HTTP request metrics
type HTTPMonitor struct {
	requests []HTTPRequestMetrics
	mu       sync.RWMutex
}

// NewHTTPMonitor creates a new HTTP monitor instance
func NewHTTPMonitor() *HTTPMonitor {
	return &HTTPMonitor{
		requests: make([]HTTPRequestMetrics, 0),
	}
}

// MonitoringRoundTripper wraps http.RoundTripper to capture HTTP metrics
type MonitoringRoundTripper struct {
	transport    http.RoundTripper
	monitor      *HTTPMonitor
	functionName string
}

// RoundTrip implements http.RoundTripper interface with monitoring capabilities
func (m *MonitoringRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Copy headers for monitoring (avoid modifying original request)
	headers := make(map[string][]string)
	for key, values := range req.Header {
		headers[key] = make([]string, len(values))
		copy(headers[key], values)
	}

	// Calculate request size
	var requestSize int64
	if req.Body != nil {
		if req.ContentLength > 0 {
			requestSize = req.ContentLength
		}
	}

	// Execute the request
	resp, err := m.transport.RoundTrip(req)

	duration := time.Since(startTime)

	// Create metrics record
	metrics := HTTPRequestMetrics{
		Method:       req.Method,
		URL:          req.URL.String(),
		Headers:      headers,
		RequestSize:  requestSize,
		Duration:     duration,
		Timestamp:    startTime,
		Error:        err,
		FunctionName: m.functionName,
	}

	// Fill response metrics if successful
	if resp != nil {
		metrics.StatusCode = resp.StatusCode
		metrics.ResponseSize = resp.ContentLength
		if metrics.ResponseSize == -1 {
			// Content-Length not set, estimate from headers
			if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
				fmt.Sscanf(contentLength, "%d", &metrics.ResponseSize)
			}
		}
	}

	// Record the metrics
	m.monitor.AddRequest(metrics)

	return resp, err
}

// AddRequest safely adds a request metric to the monitor
func (h *HTTPMonitor) AddRequest(metrics HTTPRequestMetrics) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requests = append(h.requests, metrics)
}

// GetRequests returns a copy of all recorded requests
func (h *HTTPMonitor) GetRequests() []HTTPRequestMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	requests := make([]HTTPRequestMetrics, len(h.requests))
	copy(requests, h.requests)
	return requests
}

// GetRequestsByFunction returns requests filtered by function name
func (h *HTTPMonitor) GetRequestsByFunction(functionName string) []HTTPRequestMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var filtered []HTTPRequestMetrics
	for _, req := range h.requests {
		if req.FunctionName == functionName {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// GetSummary returns summary statistics of HTTP requests
func (h *HTTPMonitor) GetSummary() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	summary := make(map[string]any)

	if len(h.requests) == 0 {
		summary["total_requests"] = 0
		return summary
	}

	// Basic statistics
	summary["total_requests"] = len(h.requests)

	// Group by function
	byFunction := make(map[string]int)
	byMethod := make(map[string]int)
	var totalDuration time.Duration
	var totalRequestSize, totalResponseSize int64
	statusCodes := make(map[int]int)
	errorCount := 0

	for _, req := range h.requests {
		byFunction[req.FunctionName]++
		byMethod[req.Method]++
		totalDuration += req.Duration
		totalRequestSize += req.RequestSize
		totalResponseSize += req.ResponseSize
		statusCodes[req.StatusCode]++
		if req.Error != nil {
			errorCount++
		}
	}

	summary["requests_by_function"] = byFunction
	summary["requests_by_method"] = byMethod
	summary["status_codes"] = statusCodes
	summary["total_errors"] = errorCount
	summary["average_duration_ms"] = float64(totalDuration.Nanoseconds()) / float64(len(h.requests)) / 1_000_000
	summary["total_request_size_bytes"] = totalRequestSize
	summary["total_response_size_bytes"] = totalResponseSize

	return summary
}

// PrintDetailedReport prints a comprehensive report of all HTTP requests
func (h *HTTPMonitor) PrintDetailedReport() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	log.Println("=== HTTP MONITORING DETAILED REPORT ===")

	if len(h.requests) == 0 {
		log.Println("No HTTP requests recorded")
		return
	}

	// Sort requests by timestamp
	sortedRequests := make([]HTTPRequestMetrics, len(h.requests))
	copy(sortedRequests, h.requests)
	sort.Slice(sortedRequests, func(i, j int) bool {
		return sortedRequests[i].Timestamp.Before(sortedRequests[j].Timestamp)
	})

	// Print summary
	summary := h.GetSummary()
	log.Printf("Total requests: %v", summary["total_requests"])
	log.Printf("Requests by function: %v", summary["requests_by_function"])
	log.Printf("Requests by method: %v", summary["requests_by_method"])
	log.Printf("Status codes: %v", summary["status_codes"])
	log.Printf("Total errors: %v", summary["total_errors"])
	log.Printf("Average duration: %.2f ms", summary["average_duration_ms"])
	log.Printf("Total request size: %v bytes", summary["total_request_size_bytes"])
	log.Printf("Total response size: %v bytes", summary["total_response_size_bytes"])
	log.Println()

	// Print individual requests
	log.Println("Individual requests:")
	for i, req := range sortedRequests {
		log.Printf("Request #%d [%s]:", i+1, req.FunctionName)
		log.Printf("  %s %s", req.Method, req.URL)
		log.Printf("  Status: %d, Duration: %v", req.StatusCode, req.Duration)
		log.Printf("  Request size: %d bytes, Response size: %d bytes", req.RequestSize, req.ResponseSize)

		// Print key headers (excluding sensitive ones)
		keyHeaders := []string{"Content-Type", "Accept", "User-Agent", "Content-Length"}
		for _, header := range keyHeaders {
			if values, exists := req.Headers[header]; exists {
				log.Printf("  %s: %s", header, strings.Join(values, ", "))
			}
		}

		if req.Error != nil {
			log.Printf("  ERROR: %v", req.Error)
		}
		log.Println()
	}

	log.Println("=== END OF HTTP MONITORING REPORT ===")
}

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

// createRemoteOptionsWithMonitoring creates remote options with HTTP monitoring
func createRemoteOptionsWithMonitoring(monitor *HTTPMonitor, functionName string) []remote.Option {
	baseTransport := &http.Transport{
		DisableKeepAlives:     false,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    false,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	monitoringTransport := &MonitoringRoundTripper{
		transport:    baseTransport,
		monitor:      monitor,
		functionName: functionName,
	}

	return []remote.Option{
		remote.WithAuth(createAuthenticator()),
		remote.WithTransport(monitoringTransport),
	}
}

// logHTTPRequest logs details about HTTP requests for debugging
func logHTTPRequest(method, url string) {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	log.Printf("Making %s request to: %s", method, url)
	log.Printf("Using Basic Auth: %s", auth[:10]+"...")
}

// fetchImageUsingRemoteImageWithMonitoring demonstrates using remote.Image() with HTTP monitoring
func fetchImageUsingRemoteImageWithMonitoring(monitor *HTTPMonitor) error {
	log.Println("=== Starting fetchImageUsingRemoteImageWithMonitoring ===")

	// Parse the image reference
	imageRef := fmt.Sprintf("%s/%s:%s", registryHost, imageName, imageTag)
	log.Printf("Parsing image reference: %s", imageRef)

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}
	log.Printf("Parsed reference: %s", ref.String())

	// Create context with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch the image using remote.Image() with monitoring
	log.Println("Calling remote.Image() with HTTP monitoring - this will make manifest and config requests")
	img, err := remote.Image(ref, createRemoteOptionsWithMonitoring(monitor, "fetchImageUsingRemoteImage")...)
	if err != nil {
		return fmt.Errorf("failed to fetch image using remote.Image(): %w", err)
	}

	// Get image manifest
	log.Println("Retrieving image manifest...")
	manifest, err := img.Manifest()
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}
	log.Printf("Manifest digest: %s", manifest.Config.Digest)
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

	log.Println("=== Completed fetchImageUsingRemoteImageWithMonitoring ===")
	return nil
}

// fetchImageUsingRemoteGetWithMonitoring demonstrates using remote.Get() with HTTP monitoring
func fetchImageUsingRemoteGetWithMonitoring(monitor *HTTPMonitor) error {
	log.Println("=== Starting fetchImageUsingRemoteGetWithMonitoring ===")

	// Parse the image reference
	imageRef := fmt.Sprintf("%s/%s:%s", registryHost, imageName, imageTag)
	log.Printf("Parsing image reference: %s", imageRef)

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}
	log.Printf("Parsed reference: %s", ref.String())

	// Create context with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use remote.Get() to fetch the descriptor and manifest with monitoring
	log.Println("Calling remote.Get() with HTTP monitoring - this makes a targeted manifest request")
	descriptor, err := remote.Get(ref, createRemoteOptionsWithMonitoring(monitor, "fetchImageUsingRemoteGet")...)
	if err != nil {
		return fmt.Errorf("failed to fetch descriptor using remote.Get(): %w", err)
	}

	log.Printf("Descriptor digest: %s", descriptor.Digest)
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

	log.Println("=== Completed fetchImageUsingRemoteGetWithMonitoring ===")
	return nil
}

func main() {
	log.Println("Starting go-containerregistry demonstration with HTTP monitoring")
	log.Printf("Target registry: %s", registryHost)
	log.Printf("Target image: %s/%s:%s", registryHost, imageName, imageTag)
	log.Println("Authentication: admin/admin123")
	log.Println()

	// Create HTTP monitor
	monitor := NewHTTPMonitor()

	log.Println("This application demonstrates HTTP monitoring for two different approaches:")
	log.Println("1. remote.Image() - High-level API that fetches complete image information")
	log.Println("2. remote.Get() - Lower-level API for targeted manifest requests")
	log.Println()

	// Test approach 1: remote.Image() with monitoring
	log.Println("🔍 Testing Approach 1: remote.Image() with HTTP monitoring")
	if err := fetchImageUsingRemoteImageWithMonitoring(monitor); err != nil {
		log.Printf("❌ Error with remote.Image(): %v", err)
	} else {
		log.Println("✅ remote.Image() completed successfully")
	}
	log.Println()

	// Small delay between requests for cleaner logs
	time.Sleep(2 * time.Second)

	// Test approach 2: remote.Get() with monitoring
	log.Println("🔍 Testing Approach 2: remote.Get() with HTTP monitoring")
	if err := fetchImageUsingRemoteGetWithMonitoring(monitor); err != nil {
		log.Printf("❌ Error with remote.Get(): %v", err)
	} else {
		log.Println("✅ remote.Get() completed successfully")
	}

	log.Println()
	log.Println("✨ Both demonstrations completed!")
	log.Println()

	// Print monitoring results
	log.Println("🔍 HTTP MONITORING RESULTS:")
	monitor.PrintDetailedReport()

	// Show summary comparison
	log.Println("📊 FUNCTION COMPARISON:")
	remoteImageRequests := monitor.GetRequestsByFunction("fetchImageUsingRemoteImage")
	remoteGetRequests := monitor.GetRequestsByFunction("fetchImageUsingRemoteGet")

	log.Printf("remote.Image() made %d HTTP requests", len(remoteImageRequests))
	log.Printf("remote.Get() made %d HTTP requests", len(remoteGetRequests))

	if len(remoteImageRequests) > 0 {
		var totalDuration time.Duration
		for _, req := range remoteImageRequests {
			totalDuration += req.Duration
		}
		log.Printf("remote.Image() total time: %v (avg: %v per request)",
			totalDuration, totalDuration/time.Duration(len(remoteImageRequests)))
	}

	if len(remoteGetRequests) > 0 {
		var totalDuration time.Duration
		for _, req := range remoteGetRequests {
			totalDuration += req.Duration
		}
		log.Printf("remote.Get() total time: %v (avg: %v per request)",
			totalDuration, totalDuration/time.Duration(len(remoteGetRequests)))
	}

	log.Println()
	log.Println("Key insights from HTTP monitoring:")
	log.Println("• remote.Image() typically makes multiple requests (manifest + config + potentially layers)")
	log.Println("• remote.Get() makes targeted requests for specific information")
	log.Println("• All requests include proper authentication headers")
	log.Println("• Request/response sizes and timings are now visible")
	log.Println("• You can track exactly what URLs are being accessed")
}
