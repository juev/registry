# LOG

```plain
2025/08/05 09:11:07 Starting go-containerregistry demonstration with HTTP monitoring
2025/08/05 09:11:07 Target registry: localhost:8082
2025/08/05 09:11:07 Target image: localhost:8082/nexus-test:latest
2025/08/05 09:11:07 Authentication: admin/admin123
2025/08/05 09:11:07
2025/08/05 09:11:07 This application demonstrates HTTP monitoring for two different approaches:
2025/08/05 09:11:07 1. remote.Image() - High-level API that fetches complete image information
2025/08/05 09:11:07 2. remote.Get() - Lower-level API for targeted manifest requests
2025/08/05 09:11:07
2025/08/05 09:11:07 🔍 Testing Approach 1: remote.Image() with HTTP monitoring
2025/08/05 09:11:07 === Starting fetchImageUsingRemoteImageWithMonitoring ===
2025/08/05 09:11:07 Parsing image reference: localhost:8082/nexus-test:latest
2025/08/05 09:11:07 Parsed reference: localhost:8082/nexus-test:latest
2025/08/05 09:11:07 Calling remote.Image() with HTTP monitoring - this will make manifest and config requests
2025/08/05 09:11:07 Retrieving image manifest...
2025/08/05 09:11:07 Manifest digest: sha256:d57353dc48511e072aaad614e6a7901ee31c6e139fac78bd5f5e70394fbca954
2025/08/05 09:11:07 Manifest media type: application/vnd.docker.distribution.manifest.v2+json
2025/08/05 09:11:07 Manifest schema version: 2
2025/08/05 09:11:07 Number of layers: 9
2025/08/05 09:11:07 Retrieving image config...
2025/08/05 09:11:07 Image architecture: arm64
2025/08/05 09:11:07 Image OS: linux
2025/08/05 09:11:07 Environment variables: 1 entries
2025/08/05 09:11:07 Image size: 2192 bytes
2025/08/05 09:11:07 === Completed fetchImageUsingRemoteImageWithMonitoring ===
2025/08/05 09:11:07 ✅ remote.Image() completed successfully
2025/08/05 09:11:07
2025/08/05 09:11:09 🔍 Testing Approach 2: remote.Get() with HTTP monitoring
2025/08/05 09:11:09 === Starting fetchImageUsingRemoteGetWithMonitoring ===
2025/08/05 09:11:09 Parsing image reference: localhost:8082/nexus-test:latest
2025/08/05 09:11:09 Parsed reference: localhost:8082/nexus-test:latest
2025/08/05 09:11:09 Calling remote.Get() with HTTP monitoring - this makes a targeted manifest request
2025/08/05 09:11:09 Descriptor digest: sha256:8cd22be2b6385ea81fe8a9802e0edd10dd87ab991b209e4a99f1e0d50ae18504
2025/08/05 09:11:09 Descriptor media type: application/vnd.docker.distribution.manifest.v2+json
2025/08/05 09:11:09 Descriptor size: 2192 bytes
2025/08/05 09:11:09 Descriptor digest: sha256:8cd22be2b6385ea81fe8a9802e0edd10dd87ab991b209e4a99f1e0d50ae18504
2025/08/05 09:11:09 Parsing manifest from descriptor...
2025/08/05 09:11:09 Manifest layers count: 9
2025/08/05 09:11:09 Raw manifest size from descriptor: 2192 bytes
2025/08/05 09:11:09 First layer digest: sha256:95459497489f07b9d71d294c852a09f9bbf1af51bb35db752a31f6f48935e293
2025/08/05 09:11:09 First layer size: 3342657 bytes
2025/08/05 09:11:09 === Completed fetchImageUsingRemoteGetWithMonitoring ===
2025/08/05 09:11:09 ✅ remote.Get() completed successfully
2025/08/05 09:11:09
2025/08/05 09:11:09 ✨ Both demonstrations completed!
2025/08/05 09:11:09
2025/08/05 09:11:09 🔍 HTTP MONITORING RESULTS:
2025/08/05 09:11:09 === HTTP MONITORING DETAILED REPORT ===
2025/08/05 09:11:09 Total requests: 7
2025/08/05 09:11:09 Requests by function: map[fetchImageUsingRemoteGet:3 fetchImageUsingRemoteImage:4]
2025/08/05 09:11:09 Requests by method: map[GET:7]
2025/08/05 09:11:09 Status codes: map[0:2 200:3 401:2]
2025/08/05 09:11:09 Total errors: 2
2025/08/05 09:11:09 Average duration: 12.97 ms
2025/08/05 09:11:09 Total request size: 0 bytes
2025/08/05 09:11:09 Total response size: 8816 bytes
2025/08/05 09:11:09
2025/08/05 09:11:09 Individual requests:
2025/08/05 09:11:09 Request #1 [fetchImageUsingRemoteImage]:
2025/08/05 09:11:09   GET https://localhost:8082/v2/
2025/08/05 09:11:09   Status: 0, Duration: 13.015334ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 0 bytes
2025/08/05 09:11:09   ERROR: tls: first record does not look like a TLS handshake
2025/08/05 09:11:09
2025/08/05 09:11:09 Request #2 [fetchImageUsingRemoteImage]:
2025/08/05 09:11:09   GET http://localhost:8082/v2/
2025/08/05 09:11:09   Status: 401, Duration: 13.208542ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 113 bytes
2025/08/05 09:11:09
2025/08/05 09:11:09 Request #3 [fetchImageUsingRemoteImage]:
2025/08/05 09:11:09   GET http://localhost:8082/v2/nexus-test/manifests/latest
2025/08/05 09:11:09   Status: 200, Duration: 15.182125ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 2192 bytes
2025/08/05 09:11:09   Accept: application/vnd.docker.distribution.manifest.v1+json,application/vnd.docker.distribution.manifest.v1+prettyjws,application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json
2025/08/05 09:11:09   User-Agent: go-containerregistry/v0.20.3
2025/08/05 09:11:09
2025/08/05 09:11:09 Request #4 [fetchImageUsingRemoteImage]:
2025/08/05 09:11:09   GET http://localhost:8082/v2/nexus-test/blobs/sha256:d57353dc48511e072aaad614e6a7901ee31c6e139fac78bd5f5e70394fbca954
2025/08/05 09:11:09   Status: 200, Duration: 4.647959ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 4206 bytes
2025/08/05 09:11:09   User-Agent: go-containerregistry/v0.20.3
2025/08/05 09:11:09
2025/08/05 09:11:09 Request #5 [fetchImageUsingRemoteGet]:
2025/08/05 09:11:09   GET https://localhost:8082/v2/
2025/08/05 09:11:09   Status: 0, Duration: 11.733333ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 0 bytes
2025/08/05 09:11:09   ERROR: tls: first record does not look like a TLS handshake
2025/08/05 09:11:09
2025/08/05 09:11:09 Request #6 [fetchImageUsingRemoteGet]:
2025/08/05 09:11:09   GET http://localhost:8082/v2/
2025/08/05 09:11:09   Status: 401, Duration: 13.66975ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 113 bytes
2025/08/05 09:11:09
2025/08/05 09:11:09 Request #7 [fetchImageUsingRemoteGet]:
2025/08/05 09:11:09   GET http://localhost:8082/v2/nexus-test/manifests/latest
2025/08/05 09:11:09   Status: 200, Duration: 19.352583ms
2025/08/05 09:11:09   Request size: 0 bytes, Response size: 2192 bytes
2025/08/05 09:11:09   Accept: application/vnd.docker.distribution.manifest.v1+json,application/vnd.docker.distribution.manifest.v1+prettyjws,application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json
2025/08/05 09:11:09   User-Agent: go-containerregistry/v0.20.3
2025/08/05 09:11:09
2025/08/05 09:11:09 === END OF HTTP MONITORING REPORT ===
2025/08/05 09:11:09 📊 FUNCTION COMPARISON:
2025/08/05 09:11:09 remote.Image() made 4 HTTP requests
2025/08/05 09:11:09 remote.Get() made 3 HTTP requests
2025/08/05 09:11:09 remote.Image() total time: 46.05396ms (avg: 11.51349ms per request)
2025/08/05 09:11:09 remote.Get() total time: 44.755666ms (avg: 14.918555ms per request)
2025/08/05 09:11:09
2025/08/05 09:11:09 Key insights from HTTP monitoring:
2025/08/05 09:11:09 • remote.Image() typically makes multiple requests (manifest + config + potentially layers)
2025/08/05 09:11:09 • remote.Get() makes targeted requests for specific information
2025/08/05 09:11:09 • All requests include proper authentication headers
2025/08/05 09:11:09 • Request/response sizes and timings are now visible
2025/08/05 09:11:09 • You can track exactly what URLs are being accessed
```
