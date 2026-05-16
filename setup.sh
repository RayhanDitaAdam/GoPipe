#!/bin/bash

# Warna untuk output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up GoPipe dependencies...${NC}"

# Jalankan go mod tidy
if go mod tidy; then
    echo -e "${GREEN}✅ Dependencies are ready!${NC}"
else
    echo -e "❌ Gagal menjalankan go mod tidy."
    exit 1
fi
