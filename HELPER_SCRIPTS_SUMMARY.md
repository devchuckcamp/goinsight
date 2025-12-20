# Helper Scripts Documentation

## Overview

Two convenience shell scripts have been created to simplify GoInsight project management:

1. **`build-start.sh`** - Builds and starts the entire application
2. **`stop-service.sh`** - Stops all services gracefully

---

## Files Created

### 1. `build-start.sh` (3.0 KB)

**Purpose**: Automate the complete build and startup process

**Features**:
- Docker status verification
- Go installation verification
- Automatic Go application build
- Docker Compose container startup
- Service health checking
- Helpful error messages
- Colored output for clarity
- Progress indication

**Usage**:
```bash
./build-start.sh
```

**Steps Performed**:
1. Checks if Docker is running
2. Verifies Go is installed
3. Builds Go application (`go build ./cmd/api`)
4. Starts Docker containers (`docker-compose up -d`)
5. Waits for services to be ready (3 seconds)
6. Performs health check on API (up to 20 seconds)
7. Displays useful next steps and command examples

**Exit Codes**:
- `0` - Success
- `1` - Docker not running, Go not found, build failed, or health check failed

**Success Output**:
```
================================================
✓ GoInsight is running!
================================================

Services:
  ✓ API Server: http://localhost:8080
  ✓ PostgreSQL: localhost:5432

Useful Commands:
  View logs:     docker-compose logs -f
  View API logs: docker-compose logs -f goinsight-api
  Stop service:  ./stop-service.sh
```

---

### 2. `stop-service.sh` (1.8 KB)

**Purpose**: Gracefully stop all containers and clean up

**Features**:
- docker-compose.yml verification
- Safe container shutdown
- Verification of cleanup
- Data preservation (volumes kept)
- Helpful next steps

**Usage**:
```bash
./stop-service.sh
```

**Steps Performed**:
1. Verifies `docker-compose.yml` exists
2. Stops all Docker containers (`docker-compose down`)
3. Verifies all containers are stopped
4. Displays next steps

**Exit Codes**:
- `0` - Success
- `1` - docker-compose.yml not found or stop command failed

**Success Output**:
```
================================================
✓ GoInsight service stopped
================================================

Useful Commands:
  Start again:   ./build-start.sh
  View logs:     docker-compose logs
  Remove data:   docker-compose down -v
  Restart:       docker-compose restart
```

---

### 3. `SCRIPTS_README.md` (8.0 KB)

**Purpose**: Comprehensive guide for the helper scripts

**Contents**:
- Script descriptions and features
- Usage instructions
- Requirements and setup
- Troubleshooting guide
- Advanced usage examples
- Performance tips
- CI/CD integration examples
- Exit code reference

---

## Quick Start

### One-Time Setup
```bash
# Make scripts executable (if needed)
chmod +x build-start.sh stop-service.sh

# Start the application
./build-start.sh
```

### Daily Workflow
```bash
# Start with full build
./build-start.sh

# Work on code...

# Stop when done
./stop-service.sh

# Next day, start again
./build-start.sh
```

### Faster Restart (code changes only)
```bash
# If you only modified Go code
./build-start.sh

# If nothing changed, just restart containers
docker-compose restart
```

---

## Key Benefits

### For Developers
- ✅ One-command startup (no manual steps)
- ✅ Automatic health checking
- ✅ Clear error messages
- ✅ Helpful command suggestions
- ✅ Colored output for readability

### For DevOps/CI-CD
- ✅ Exit codes for automation
- ✅ Idempotent operations
- ✅ Dependency verification
- ✅ Clean shutdown process
- ✅ Data preservation options

### For Troubleshooting
- ✅ Step-by-step progress
- ✅ Service health verification
- ✅ Container status checking
- ✅ Log access instructions
- ✅ Reset options documented

---

## Usage Examples

### Start Application
```bash
$ ./build-start.sh
================================================
GoInsight - Build and Start Script
================================================

[1/4] Checking Docker status...
✓ Docker is running

[2/4] Verifying Go installation...
✓ Go go1.25.3 found

[3/4] Building Go application...
✓ Build successful

[4/4] Starting Docker containers...
✓ Docker containers started

================================================
✓ GoInsight is running!
================================================

Services:
  ✓ API Server: http://localhost:8080
  ✓ PostgreSQL: localhost:5432
```

### Stop Application
```bash
$ ./stop-service.sh
================================================
GoInsight - Stop Service Script
================================================

[1/3] Stopping Docker containers...
✓ Containers stopped

[2/3] Verifying all containers are stopped...
✓ All containers stopped

[3/3] Cleanup complete

================================================
✓ GoInsight service stopped
================================================
```

### Test API After Startup
```bash
$ curl -s http://localhost:8080/api/health | jq .
{
  "status": "healthy"
}
```

---

## Script Features Comparison

| Feature | build-start.sh | stop-service.sh |
|---------|----------------|-----------------|
| Dependency checking | ✓ | ✓ |
| Error handling | ✓ | ✓ |
| Colored output | ✓ | ✓ |
| Progress indication | ✓ | ✓ |
| Health checking | ✓ | - |
| Helpful messages | ✓ | ✓ |
| Exit codes | ✓ | ✓ |
| Data preservation | - | ✓ |

---

## Common Tasks

### View Live Logs
```bash
docker-compose logs -f
```

### Check Container Status
```bash
docker-compose ps
```

### Restart a Service
```bash
docker-compose restart goinsight-api
```

### Clean Rebuild
```bash
docker-compose down -v
./build-start.sh
```

### Reset Database
```bash
docker-compose down -v
./build-start.sh
```

---

## Troubleshooting

### Script Permission Denied
```bash
chmod +x build-start.sh stop-service.sh
```

### Docker Not Running
- macOS/Windows: Open Docker Desktop
- Linux: `sudo systemctl start docker`

### Build Fails
```bash
# Check error messages
./build-start.sh
# If build fails, verify Go code
go build ./cmd/api
```

### API Not Responding
```bash
# Check container logs
docker-compose logs goinsight-api

# Restart containers
./stop-service.sh
./build-start.sh
```

### Port Already in Use
```bash
# View what's using port 8080
lsof -i :8080          # macOS/Linux
netstat -ano | grep :8080  # Windows
```

---

## Requirements

### System Requirements
- **Docker** & **Docker Compose** - For containerization
- **Go 1.21+** - For building
- **Bash** - For running scripts
  - macOS: Built-in
  - Linux: Built-in
  - Windows: Git Bash, WSL, or similar

### Environment Setup
Create `.env` file in project root:

```bash
POSTGRES_USER=goinsight
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=goinsight
DATABASE_URL=postgresql://goinsight:yourpassword@localhost:5432/goinsight
LLM_PROVIDER=mock
```

---

## Advanced Usage

### Rebuild Without Cache
```bash
docker-compose build --no-cache
docker-compose up -d
```

### Execute Commands in Container
```bash
# Connect to database
docker-compose exec goinsight-postgres psql -U goinsight -d goinsight

# Run shell in API container
docker-compose exec goinsight-api sh
```

### Monitor Resource Usage
```bash
docker stats
```

---

## Performance Tips

- Use `./build-start.sh` only when code changes
- Use `docker-compose restart` for faster restart
- Avoid `--no-cache` unless necessary
- Close unused Docker containers

---

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Start Services
  run: ./build-start.sh

- name: Run Tests
  run: pytest tests/

- name: Stop Services
  run: ./stop-service.sh
```

### GitLab CI Example
```yaml
before_script:
  - ./build-start.sh

script:
  - pytest tests/

after_script:
  - ./stop-service.sh
```

---

## File Permissions

Scripts are created with proper permissions:
```bash
-rwxr-xr-x build-start.sh    # Executable by owner, readable by others
-rwxr-xr-x stop-service.sh   # Executable by owner, readable by others
```

---

## Related Documentation

- [README.md](./README.md) - Project overview
- [DESIGN_PATTERNS.md](./DESIGN_PATTERNS.md) - Architecture guide
- [SCRIPTS_README.md](./SCRIPTS_README.md) - Detailed script guide
- [docker-compose.yml](./docker-compose.yml) - Container configuration
- [Dockerfile](./Dockerfile) - Image build configuration

---

## Summary

These helper scripts significantly improve the developer experience by:
1. **Automating** the build and startup process
2. **Verifying** dependencies are available
3. **Checking** service health
4. **Providing** helpful feedback
5. **Enabling** quick starts and stops

### Time Saved
- **Without scripts**: 5-10 manual commands, troubleshooting needed
- **With scripts**: 1 command, automatic verification

---

**Created**: December 20, 2025  
**Version**: 1.0  
**Status**: ✅ Production Ready
