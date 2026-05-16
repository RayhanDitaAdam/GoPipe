#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building GoPipe${NC}"

# Build binary
if go build -o gopipe ./cmd/gopipe; then
    echo -e "${GREEN}Build success${NC}"
else
    echo -e "Failed to build binary"
    exit 1
fi

# Move to /usr/local/bin
echo -e "${BLUE}Installing to /usr/local/bin (requires sudo)${NC}"
if sudo mv gopipe /usr/local/bin/gopipe; then
    echo -e "${GREEN}GoPipe installed successfully${NC}"
    echo -e "${YELLOW}Type 'gopipe' to start using${NC}"
else
    echo -e "Failed to move binary to /usr/local/bin"
    exit 1
fi
