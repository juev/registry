#!/bin/bash
set -e

# Nexus configuration script
# Configures Nexus for Docker registry functionality via REST API

NEXUS_URL="http://localhost:8081"
NEXUS_USER="admin"
NEXUS_PASS="admin123"
DOCKER_PORT="8082"
REPO_NAME="docker-registry"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Nexus is ready
wait_for_nexus() {
    print_status "Waiting for Nexus to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$NEXUS_URL/service/rest/v1/status" > /dev/null 2>&1; then
            print_status "Nexus is ready!"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - Nexus not ready yet, waiting 10 seconds..."
        sleep 10
        ((attempt++))
    done
    
    print_error "Nexus failed to start within expected time"
    exit 1
}

# Function to make authenticated API calls
nexus_api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local content_type="${4:-application/json}"
    
    local curl_cmd="curl -s -w '%{http_code}' -u $NEXUS_USER:$NEXUS_PASS"
    
    if [ "$method" = "POST" ] || [ "$method" = "PUT" ]; then
        curl_cmd="$curl_cmd -X $method -H 'Content-Type: $content_type'"
        if [ -n "$data" ]; then
            curl_cmd="$curl_cmd -d '$data'"
        fi
    elif [ "$method" = "DELETE" ]; then
        curl_cmd="$curl_cmd -X DELETE"
    fi
    
    curl_cmd="$curl_cmd $NEXUS_URL$endpoint"
    
    # Execute the curl command and capture both body and status code
    local response=$(eval $curl_cmd)
    local http_code="${response: -3}"
    local body="${response%???}"
    
    echo "$http_code:$body"
}

# Function to check if repository exists
check_repository_exists() {
    local repo_name="$1"
    print_status "Checking if repository '$repo_name' exists..."
    
    local response=$(nexus_api_call "GET" "/service/rest/v1/repositories")
    local http_code=$(echo "$response" | cut -d':' -f1)
    local body=$(echo "$response" | cut -d':' -f2-)
    
    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q "\"name\":\"$repo_name\""; then
            return 0  # Repository exists
        fi
    fi
    return 1  # Repository doesn't exist
}

# Function to create Docker hosted repository
create_docker_repository() {
    print_status "Creating Docker hosted repository '$REPO_NAME'..."
    
    local repo_config='{
        "name": "'$REPO_NAME'",
        "online": true,
        "storage": {
            "blobStoreName": "default",
            "strictContentTypeValidation": true,
            "writePolicy": "ALLOW"
        },
        "docker": {
            "v1Enabled": true,
            "forceBasicAuth": true,
            "httpPort": '$DOCKER_PORT',
            "httpsPort": null,
            "subdomain": null
        },
        "component": {
            "proprietaryComponents": true
        }
    }'
    
    local response=$(nexus_api_call "POST" "/service/rest/v1/repositories/docker/hosted" "$repo_config")
    local http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "201" ]; then
        print_status "Docker repository '$REPO_NAME' created successfully!"
        return 0
    elif [ "$http_code" = "400" ]; then
        print_warning "Repository '$REPO_NAME' may already exist, attempting to update..."
        update_docker_repository
        return $?
    else
        print_error "Failed to create repository. HTTP code: $http_code"
        echo "Response: $(echo "$response" | cut -d':' -f2-)"
        return 1
    fi
}

# Function to update existing Docker repository with correct settings
update_docker_repository() {
    print_status "Updating Docker repository '$REPO_NAME' configuration..."
    
    local repo_config='{
        "name": "'$REPO_NAME'",
        "online": true,
        "storage": {
            "blobStoreName": "default",
            "strictContentTypeValidation": true,
            "writePolicy": "ALLOW"
        },
        "docker": {
            "v1Enabled": true,
            "forceBasicAuth": true,
            "httpPort": '$DOCKER_PORT',
            "httpsPort": null,
            "subdomain": null
        },
        "component": {
            "proprietaryComponents": true
        }
    }'
    
    local response=$(nexus_api_call "PUT" "/service/rest/v1/repositories/docker/hosted/$REPO_NAME" "$repo_config")
    local http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "204" ]; then
        print_status "Docker repository '$REPO_NAME' updated successfully!"
        return 0
    else
        print_error "Failed to update repository. HTTP code: $http_code"
        echo "Response: $(echo "$response" | cut -d':' -f2-)"
        return 1
    fi
}

# Function to configure security realms
configure_security_realms() {
    print_status "Configuring security realms..."
    
    # Get current security configuration
    local response=$(nexus_api_call "GET" "/service/rest/v1/security/realms/available")
    local http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" != "200" ]; then
        print_error "Failed to get available security realms"
        return 1
    fi
    
    # Configure active realms (enable NexusAuthenticatingRealm and DockerToken if needed)
    local realm_config='["NexusAuthenticatingRealm", "NexusAuthorizingRealm"]'
    
    response=$(nexus_api_call "PUT" "/service/rest/v1/security/realms/active" "$realm_config")
    http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "204" ]; then
        print_status "Security realms configured successfully!"
        return 0
    else
        print_error "Failed to configure security realms. HTTP code: $http_code"
        return 1
    fi
}

# Function to configure anonymous access and fix authorization issues
configure_anonymous_access() {
    print_status "Configuring anonymous access for Docker pulls..."
    
    # Step 1: Enable anonymous access
    local response=$(nexus_api_call "GET" "/service/rest/v1/security/anonymous")
    local http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "200" ]; then
        local body=$(echo "$response" | cut -d':' -f2-)
        if echo "$body" | grep -q '"enabled":true'; then
            print_status "Anonymous access is already enabled"
        else
            print_status "Enabling anonymous access..."
            local anonymous_config='{
                "enabled": true,
                "userId": "anonymous",
                "realmName": "NexusAuthorizingRealm"
            }'
            
            response=$(nexus_api_call "PUT" "/service/rest/v1/security/anonymous" "$anonymous_config")
            http_code=$(echo "$response" | cut -d':' -f1)
            
            if [ "$http_code" = "200" ]; then
                print_status "Anonymous access enabled successfully!"
            else
                print_warning "Failed to enable anonymous access. HTTP code: $http_code"
            fi
        fi
    fi
    
    # Step 2: Fix anonymous user by assigning proper roles
    print_status "Configuring anonymous user with proper roles..."
    local anonymous_user_config='{
        "userId": "anonymous",
        "firstName": "Anonymous",
        "lastName": "User", 
        "emailAddress": "anonymous@example.org",
        "source": "default",
        "status": "active",
        "readOnly": false,
        "roles": ["nx-anonymous"],
        "externalRoles": []
    }'
    
    response=$(nexus_api_call "PUT" "/service/rest/v1/security/users/anonymous" "$anonymous_user_config")
    http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "200" ]; then
        print_status "Anonymous user configured with nx-anonymous role!"
    else
        print_warning "Failed to configure anonymous user. HTTP code: $http_code"
        echo "Response: $(echo "$response" | cut -d':' -f2-)"
    fi
    
    # Step 3: Create or update Docker pull privilege (optional, as nx-anonymous should cover this)
    print_status "Creating Docker pull privilege..."
    local privilege_config='{
        "name": "docker-registry-pull",
        "description": "Allow pulling from docker-registry",
        "type": "repository-view",
        "format": "docker",
        "repository": "'$REPO_NAME'",
        "actions": ["READ", "BROWSE"]
    }'
    
    response=$(nexus_api_call "POST" "/service/rest/v1/security/privileges/repository-view" "$privilege_config")
    http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "201" ]; then
        print_status "Docker pull privilege created successfully!"
    elif [ "$http_code" = "400" ]; then
        print_warning "Docker pull privilege may already exist"
    fi
}

# Function to disable Docker token authentication (for local testing)
disable_docker_token_auth() {
    print_status "Configuring Docker token authentication settings..."
    
    # Note: Nexus doesn't have a direct API to disable Docker token auth
    # This is typically handled by not including DockerToken realm in active realms
    # We already configured this in the configure_security_realms function
    
    print_status "Docker token authentication is disabled (not in active realms)"
}

# Function to verify configuration
verify_configuration() {
    print_status "Verifying Nexus configuration..."
    
    # Check if repository exists and is accessible
    if check_repository_exists "$REPO_NAME"; then
        print_status "✓ Repository '$REPO_NAME' exists"
    else
        print_error "✗ Repository '$REPO_NAME' not found"
        return 1
    fi
    
    # Check if Docker port is accessible
    if curl -f -s "http://localhost:$DOCKER_PORT/v2/" > /dev/null 2>&1; then
        print_status "✓ Docker registry is accessible on port $DOCKER_PORT"
    else
        print_error "✗ Docker registry is NOT accessible on port $DOCKER_PORT"
        print_error "🚨 MANUAL ACTION REQUIRED: EULA not accepted!"
        echo ""
        echo "To activate Docker registry on port $DOCKER_PORT:"
        echo "1. Open in browser: $NEXUS_URL"
        echo "2. Login: admin/admin123"
        echo "3. Complete Setup Wizard and MUST accept EULA"
        echo "4. Verify: curl http://localhost:$DOCKER_PORT/v2/"
        echo ""
        print_warning "Docker repository created but blocked until EULA acceptance"
    fi
    
    # Check anonymous access
    local response=$(nexus_api_call "GET" "/service/rest/v1/security/anonymous")
    local http_code=$(echo "$response" | cut -d':' -f1)
    
    if [ "$http_code" = "200" ]; then
        local body=$(echo "$response" | cut -d':' -f2-)
        if echo "$body" | grep -q '"enabled":true'; then
            print_status "✓ Anonymous access is enabled"
        else
            print_warning "✗ Anonymous access is disabled"
        fi
    fi
    
    print_status "Configuration verification completed"
}

# Main execution
main() {
    echo "=========================================="
    echo "Nexus Docker Registry Configuration Script"
    echo "=========================================="
    echo ""
    
    # Wait for Nexus to be ready
    wait_for_nexus
    
    # Check if repository already exists
    if check_repository_exists "$REPO_NAME"; then
        print_warning "Repository '$REPO_NAME' already exists. Skipping creation."
    else
        # Create Docker repository
        if ! create_docker_repository; then
            print_error "Failed to create Docker repository"
            exit 1
        fi
    fi
    
    # Configure security realms
    configure_security_realms
    
    # Configure anonymous access
    configure_anonymous_access
    
    # Disable Docker token authentication
    disable_docker_token_auth
    
    # Verify configuration
    verify_configuration
    
    echo ""
    print_status "=========================================="
    print_status "Nexus Docker Registry Configuration Complete!"
    print_status "=========================================="
    echo ""
    echo "Configuration Summary:"
    echo "- Nexus URL: $NEXUS_URL"
    echo "- Docker Registry: localhost:$DOCKER_PORT"
    echo "- Repository Name: $REPO_NAME"
    echo "- Anonymous Pull: Enabled"
    echo "- Docker Token Auth: Disabled"
    echo ""
    
    # Check if EULA needs to be accepted
    if ! curl -f -s "http://localhost:$DOCKER_PORT/v2/" > /dev/null 2>&1; then
        print_error "🚨 IMPORTANT: You need to accept EULA to complete setup!"
        echo ""
        echo "NEXT STEPS:"
        echo "1. Open in browser: $NEXUS_URL"
        echo "2. Login: admin/admin123"
        echo "3. Complete Setup Wizard and MUST accept EULA"
        echo "4. After accepting EULA verify: curl http://localhost:$DOCKER_PORT/v2/"
        echo ""
        print_warning "Docker registry will be blocked until EULA acceptance"
        echo ""
    else
        echo "✅ Docker registry is ready to use!"
        echo ""
        echo "Test your registry:"
        echo "1. Pull a test image:"
        echo "   docker pull hello-world"
        echo ""
        echo "2. Tag it for your registry:"
        echo "   docker tag hello-world localhost:$DOCKER_PORT/hello-world:latest"
        echo ""
        echo "3. Push to your registry:"
        echo "   docker push localhost:$DOCKER_PORT/hello-world:latest"
        echo ""
        echo "4. Pull from your registry:"
        echo "   docker pull localhost:$DOCKER_PORT/hello-world:latest"
        echo ""
    fi
    
    print_warning "Note: You may need to configure Docker daemon to allow insecure registries:"
    echo "Add \"localhost:$DOCKER_PORT\" to your Docker daemon insecure-registries list"
}

# Run main function
main "$@"