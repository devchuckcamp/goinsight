#!/bin/bash

# GoInsight Build and Start Script
# This script builds the application and starts it via Docker Compose

set -e

echo "================================================"
echo "GoInsight - Build and Start Script"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Docker is running
echo -e "${BLUE}[1/4] Checking Docker status...${NC}"
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Verify Go is installed
echo -e "${BLUE}[2/4] Verifying Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed. Please install Go and try again.${NC}"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓ Go ${GO_VERSION} found${NC}"
echo ""

# Build the Go application
echo -e "${BLUE}[3/4] Building Go application...${NC}"
if go build -v ./cmd/api; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed. Please check the error messages above.${NC}"
    exit 1
fi
echo ""

# Start Docker containers
echo -e "${BLUE}[4/4] Starting Docker containers...${NC}"
if docker-compose up -d; then
    echo -e "${GREEN}✓ Docker containers started${NC}"
else
    echo -e "${RED}✗ Failed to start Docker containers.${NC}"
    exit 1
fi
echo ""

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 3

# Check service health
echo -e "${BLUE}Checking service health...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API is healthy${NC}"
        break
    fi
    
    if [ $i -lt 10 ]; then
        echo -e "${YELLOW}  Waiting for API... ($i/10)${NC}"
        sleep 2
    else
        echo -e "${RED}✗ API health check failed after 20 seconds${NC}"
        echo "  Try: docker-compose logs goinsight-api"
        exit 1
    fi
done
echo ""

# Display status and next steps
echo "================================================"
echo -e "${GREEN}✓ GoInsight is running!${NC}"
echo "================================================"
echo ""
echo "Services:"
echo -e "  ${GREEN}✓${NC} API Server: http://localhost:8080"
echo -e "  ${GREEN}✓${NC} PostgreSQL: localhost:5432"
echo ""
echo "Useful Commands:"
echo "  View logs:     docker-compose logs -f"
echo "  View API logs: docker-compose logs -f goinsight-api"
echo "  Stop service:  ./stop-service.sh"
echo "  Health check:  curl http://localhost:8080/api/health"
echo ""
echo "Try the API:"
echo "  curl -X POST http://localhost:8080/api/ask \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"question\": \"Show me negative feedback\"}'"
echo ""
echo "================================================"
