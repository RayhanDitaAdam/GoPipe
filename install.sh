#!/bin/bash

# Warna untuk output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Building GoPipe...${NC}"

# Build binary
if go build -o gopipe main.go; then
    echo -e "${GREEN}✅ Build success!${NC}"
else
    echo -e "❌ Gagal build binary."
    exit 1
fi

# Pindahkan ke /usr/local/bin
echo -e "${BLUE}📦 Installing to /usr/local/bin (requires sudo)...${NC}"
if sudo mv gopipe /usr/local/bin/gopipe; then
    echo -e "${GREEN}✨ GoPipe installed successfully!${NC}"
    echo -e "${YELLOW}Ketik 'gopipe' untuk mulai menggunakan. Mantap, bre! 🔥${NC}"
else
    echo -e "❌ Gagal memindahkan binary ke /usr/local/bin."
    exit 1
fi
