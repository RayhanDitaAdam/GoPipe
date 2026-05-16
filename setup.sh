#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up GoPipe dependencies${NC}"

# Run go mod tidy
if go mod tidy; then
    echo -e "${GREEN}Dependencies are ready${NC}"
else
    echo -e "Failed to run go mod tidy"
    exit 1
fi
