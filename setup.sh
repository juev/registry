#!/bin/bash
set -e

echo "Setting up Nexus Registry Environment..."

# Create necessary directories
mkdir -p nexus-config

# Start Nexus
echo "Starting Nexus container..."
docker-compose up -d

echo "Waiting for Nexus to start up (this may take a few minutes)..."
sleep 60

# Check if Nexus is running
echo "Checking Nexus status..."
until curl -f http://localhost:8081/service/rest/v1/status 2>/dev/null; do
    echo "Waiting for Nexus to be ready..."
    sleep 10
done

echo ""
echo "Nexus is now running!"
echo ""
echo "Configuring Nexus via REST API..."
./configure-nexus.sh
echo ""
echo "Next steps (if you prefer manual configuration):"
echo "1. Open http://localhost:8081 in your browser"
echo "2. Login with username 'admin' and password 'admin123'"
echo "3. Create a Docker registry repository:"
echo "   - Go to Settings > Repositories > Create repository"
echo "   - Choose 'docker (hosted)'"
echo "   - Name it 'docker-registry'"
echo "   - Set HTTP port to 8082"
echo "   - Enable Docker V1 API if needed"
echo ""
echo "To build and test the sample Docker image:"
echo "   docker build -t nexus-test:latest ."
echo ""
echo "To tag and push to your Nexus registry (after setup):"
echo "   docker tag nexus-test:latest localhost:8082/nexus-test:latest"
echo "   docker push localhost:8082/nexus-test:latest"
echo ""
echo "Note: You may need to configure Docker to allow insecure registries"
echo "Add '\"insecure-registries\": [\"localhost:8082\"]' to your Docker daemon config"