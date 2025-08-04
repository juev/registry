#!/bin/bash
set -e

# Test script for Nexus Docker registry
REGISTRY_URL="localhost:8082"
TEST_IMAGE="hello-world:latest"
REGISTRY_IMAGE="$REGISTRY_URL/hello-world:test"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_status "Testing Nexus Docker registry..."

# Test 1: Check if Nexus is running
print_status "1. Checking if Nexus is accessible..."
if curl -f -s http://localhost:8081/service/rest/v1/status > /dev/null; then
    print_status "✓ Nexus is running"
else
    print_error "✗ Nexus is not accessible"
    exit 1
fi

# Test 2: Check if Docker registry port is accessible
print_status "2. Checking Docker registry endpoint..."
if curl -f -s http://$REGISTRY_URL/v2/ > /dev/null; then
    print_status "✓ Docker registry endpoint is accessible"
else
    print_warning "✗ Docker registry endpoint may not be ready yet"
fi

# Test 3: Check repository via API
print_status "3. Checking if docker-registry repository exists..."
if curl -f -s -u admin:admin123 http://localhost:8081/service/rest/v1/repositories/docker-registry > /dev/null; then
    print_status "✓ Docker registry repository exists"
else
    print_error "✗ Docker registry repository not found"
    exit 1
fi

# Test 4: Pull test image
print_status "4. Pulling test image..."
if docker pull $TEST_IMAGE > /dev/null 2>&1; then
    print_status "✓ Test image pulled successfully"
else
    print_error "✗ Failed to pull test image"
    exit 1
fi

# Test 5: Tag for registry
print_status "5. Tagging image for registry..."
if docker tag $TEST_IMAGE $REGISTRY_IMAGE; then
    print_status "✓ Image tagged successfully"
else
    print_error "✗ Failed to tag image"
    exit 1
fi

# Test 6: Push to registry
print_status "6. Pushing image to registry..."
if docker push $REGISTRY_IMAGE > /dev/null 2>&1; then
    print_status "✓ Image pushed successfully"
else
    print_warning "✗ Failed to push image (this may be due to Docker daemon config)"
    print_warning "Make sure to configure Docker daemon with insecure registries:"
    print_warning "Add the following to your Docker daemon.json:"
    cat docker-daemon-example.json
    echo ""
fi

# Test 7: Pull from registry
print_status "7. Testing pull from registry..."
# Remove local image first
docker rmi $REGISTRY_IMAGE > /dev/null 2>&1 || true

if docker pull $REGISTRY_IMAGE > /dev/null 2>&1; then
    print_status "✓ Image pulled from registry successfully"
else
    print_warning "✗ Failed to pull from registry"
fi

# Clean up
print_status "8. Cleaning up test images..."
docker rmi $TEST_IMAGE $REGISTRY_IMAGE > /dev/null 2>&1 || true

echo ""
print_status "=========================================="
print_status "Registry Test Complete!"
print_status "=========================================="
echo ""
echo "Registry URL: http://$REGISTRY_URL"
echo "Web UI: http://localhost:8081"
echo "Login: admin/admin123"