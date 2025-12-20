#!/bin/bash

# GoInsight Stop Service Script
# This script stops the Docker containers and cleans up resources

set -e

echo "================================================"
echo "GoInsight - Stop Service Script"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Docker Compose file exists
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}✗ docker-compose.yml not found in current directory${NC}"
    echo "  Please run this script from the GoInsight root directory"
    exit 1
fi

# Stop containers
echo -e "${BLUE}[1/3] Stopping Docker containers...${NC}"
if docker-compose down; then
    echo -e "${GREEN}✓ Containers stopped${NC}"
else
    echo -e "${RED}✗ Failed to stop containers${NC}"
    exit 1
fi
echo ""

# Optional: Remove volumes (for clean shutdown)
echo -e "${YELLOW}Note: Volume data is preserved for next startup${NC}"
echo ""

# Verify containers are stopped
echo -e "${BLUE}[2/3] Verifying all containers are stopped...${NC}"
RUNNING=$(docker ps -q | wc -l)
if [ $RUNNING -eq 0 ]; then
    echo -e "${GREEN}✓ All containers stopped${NC}"
else
    echo -e "${YELLOW}⚠ ${RUNNING} container(s) still running${NC}"
fi
echo ""

# Display final status
echo -e "${BLUE}[3/3] Cleanup complete${NC}"
echo ""
echo "================================================"
echo -e "${GREEN}✓ GoInsight service stopped${NC}"
echo "================================================"
echo ""
echo "Useful Commands:"
echo "  Start again:   ./build-start.sh"
echo "  View logs:     docker-compose logs"
echo "  Remove data:   docker-compose down -v"
echo "  Restart:       docker-compose restart"
echo ""
echo "================================================"
