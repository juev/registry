# Nexus Repository Manager - Container Registry Setup

This project sets up a Nexus Repository Manager 3 instance configured for Docker registry functionality, perfect for experimenting with container registries.

## Files Overview

- `docker-compose.yml` - Main Docker Compose configuration for Nexus
- `Dockerfile` - Multi-layer test image for pushing to the registry
- `setup.sh` - Automated setup script that includes REST API configuration
- `configure-nexus.sh` - REST API script for automated Nexus configuration
- `test-registry.sh` - Test script to verify registry functionality
- `.dockerignore` - Optimizes Docker build context
- `docker-daemon-example.json` - Example Docker daemon configuration

## Quick Start

1. **Start Nexus Repository Manager:**

   ```bash
   ./setup.sh
   ```

   Or manually:

   ```bash
   docker-compose up -d
   ```

2. **Wait for Nexus to start** (usually takes 2-3 minutes on first run)

3. **Access the Web Interface:**
   - Open <http://localhost:8081> in your browser
   - Username: `admin`
   - Get the initial password:

     ```bash
     docker exec nexus-registry cat /nexus-data/admin.password
     ```

4. **🚨 IMPORTANT: Complete the setup wizard** and set a new admin password
   - **You MUST accept the EULA (End User License Agreement)**
   - Configure Anonymous Access (recommended to enable)
   - Without accepting EULA, Docker registry on port 8082 will not work

## Configure Docker Registry

### Automated Configuration (Recommended)

The setup script now automatically configures Nexus via REST API:

```bash
./setup.sh
```

**⚠️ Note:** Automated configuration will create the Docker repository, but **EULA must be accepted manually via web interface** to activate port 8082.

This will:

1. Start Nexus container
2. Wait for Nexus to be ready
3. Automatically configure Docker registry via REST API
4. Create a Docker hosted repository named "docker-registry"
5. Configure it to listen on port 8082 with forceBasicAuth enabled
6. Configure anonymous user with proper nx-anonymous role assignment
7. Set up appropriate security realms and authentication

**🔧 After running the script:**

1. Go to <http://localhost:8081>  
2. Login as admin/admin123
3. Complete Setup Wizard and accept EULA
4. Verify: `curl http://localhost:8082/v2/` (should return `{}` or require authentication)

**✅ Authentication Fixed:** The script now automatically resolves the anonymous user authorization issues that previously caused "unauthorized" errors during Docker push operations.

### Manual Configuration (Alternative)

If you prefer manual configuration:

1. Login to Nexus web interface (<http://localhost:8081>)
2. Username: `admin`, Password: `admin123`
3. Go to **Settings** (gear icon) → **Repositories**
4. Click **Create repository**
5. Choose **docker (hosted)**
6. Configure:
   - **Name**: `docker-registry`
   - **HTTP Port**: `8082`
   - **Enable Docker V1 API**: Check if needed
   - **Allow anonymous docker pull**: Check for easier testing
7. Click **Create repository**

### Test Configuration

After setup, test your registry configuration:

```bash
./test-registry.sh
```

This script will verify:

- Nexus accessibility
- Docker registry endpoint
- Repository creation
- Push/pull functionality

### 2. Configure Docker Client for Insecure Registry

Since we're using HTTP (not HTTPS), configure Docker to allow insecure registries:

**On macOS/Linux:**

1. Edit or create `/etc/docker/daemon.json`:

   ```bash
   sudo cp docker-daemon-example.json /etc/docker/daemon.json
   ```

   Or add manually:

   ```json
   {
     "insecure-registries": [
       "localhost:8082",
       "127.0.0.1:8082"
     ]
   }
   ```

2. Restart Docker daemon:

   ```bash
   sudo systemctl restart docker  # Linux
   # Or restart Docker Desktop on macOS
   ```

**On Windows:**

- Open Docker Desktop settings
- Go to Docker Engine
- Add the insecure-registries configuration to the JSON

## Testing the Registry

### 1. Build the Test Image

```bash
docker build -t nexus-test:latest .
```

### 2. Tag and Push to Nexus Registry

```bash
# Tag the image for your registry
docker tag nexus-test:latest localhost:8082/nexus-test:latest

# Push to registry
docker push localhost:8082/nexus-test:latest
```

### 3. Pull from Registry

```bash
# Remove local image first
docker rmi nexus-test:latest localhost:8082/nexus-test:latest

# Pull from registry
docker pull localhost:8082/nexus-test:latest
```

### 4. Test with Authentication (Optional)

If you didn't enable anonymous access:

```bash
# Login to registry
docker login localhost:8082

# Then push/pull as above
```

## Test Image Details

The included `Dockerfile` creates a multi-layer image with:

- **Base Layer**: Alpine Linux 3.18
- **Layer 1**: Basic utilities (curl, wget, ca-certificates)
- **Layer 2**: User and directory setup
- **Layer 3**: Configuration files
- **Layer 4**: Additional tools (bash, jq, git)
- **Layer 5**: Application script
- **Layer 6**: Test data files
- **Layer 7**: Final permissions and workspace setup

This creates a realistic multi-layer image perfect for testing registry functionality.

## Useful Commands

```bash
# View Nexus logs
docker-compose logs -f nexus

# Stop Nexus
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v

# View registry catalog (if enabled)
curl http://localhost:8082/v2/_catalog

# View image tags
curl http://localhost:8082/v2/nexus-test/tags/list

# Check image layers
docker history localhost:8082/nexus-test:latest
```

## Troubleshooting

### Common Issues

1. **🚨 "EULA not accepted" or port 8082 unavailable**
   - **Solution:** Open <http://localhost:8081> in browser
   - Login as admin/admin123
   - Complete Setup Wizard and **must accept EULA**
   - Verify: `curl http://localhost:8082/v2/`

2. **"connection refused" when pushing**
   - Ensure Nexus is fully started (check logs)
   - Verify Docker registry repository is created
   - Check insecure-registries configuration
   - **Make sure EULA is accepted (see #1)**

3. **Repository creation errors via API**

   ```plain
   [WARN] Repository 'docker-registry' may already exist or configuration error
   [ERROR] Failed to create Docker repository
   ```

   - **This is normal!** Repository is created, but EULA not accepted
   - Solution: accept EULA via web interface

4. **Authentication errors**
   - Verify credentials with `docker login localhost:8082`
   - Check repository permissions in Nexus

5. **Nexus won't start**
   - Check available memory (Nexus needs ~2GB)
   - View logs: `docker-compose logs nexus`
   - Ensure ports 8081 and 8082 are available

6. **Push fails with "unauthorized"**
   - Run `./configure-nexus.sh` to automatically fix anonymous user permissions
   - The script assigns the nx-anonymous role to the anonymous user
   - Alternatively, manually enable "Allow anonymous docker pull" in repository settings

### Resource Requirements

- **Memory**: Minimum 2GB RAM for Nexus
- **Disk**: 10GB+ for persistent storage
- **Ports**: 8081 (Web UI), 8082 (Docker Registry)

## Accessing Nexus Data

Nexus data is persisted in a Docker volume. To access:

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect registry_nexus-data

# Backup volume
docker run --rm -v registry_nexus-data:/data -v $(pwd):/backup alpine tar czf /backup/nexus-backup.tar.gz -C /data .

# Restore volume
docker run --rm -v registry_nexus-data:/data -v $(pwd):/backup alpine tar xzf /backup/nexus-backup.tar.gz -C /data
```

## Cleanup

To completely remove everything:

```bash
# Stop and remove containers, networks, and volumes
docker-compose down -v

# Remove built images
docker rmi nexus-test:latest localhost:8082/nexus-test:latest

# Remove Nexus image (optional)
docker rmi sonatype/nexus3:latest
```

This setup provides a complete Docker registry environment perfect for learning and experimentation!
