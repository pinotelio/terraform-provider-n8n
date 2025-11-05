#!/bin/bash

# Installation script for terraform-provider-n8n
# This script builds and installs the provider to the local Terraform plugins directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Map OS names
case $OS in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

PROVIDER_NAME="terraform-provider-n8n"
VERSION="0.1.0"
INSTALL_PATH="$HOME/.terraform.d/plugins/registry.terraform.io/pinotelio/n8n/$VERSION/${OS}_${ARCH}"

echo -e "${GREEN}Building terraform-provider-n8n...${NC}"
echo "OS: $OS"
echo "Architecture: $ARCH"
echo "Install path: $INSTALL_PATH"
echo ""

# Build the provider
echo -e "${YELLOW}Building provider...${NC}"
go build -o $PROVIDER_NAME

if [ ! -f $PROVIDER_NAME ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"
echo ""

# Create installation directory
echo -e "${YELLOW}Creating installation directory...${NC}"
mkdir -p "$INSTALL_PATH"

# Copy binary
echo -e "${YELLOW}Installing provider...${NC}"
cp $PROVIDER_NAME "$INSTALL_PATH/"

# Make it executable
chmod +x "$INSTALL_PATH/$PROVIDER_NAME"

echo -e "${GREEN}Installation successful!${NC}"
echo ""
echo "Provider installed to: $INSTALL_PATH"
echo ""
echo "You can now use the provider in your Terraform configurations:"
echo ""
echo "terraform {"
echo "  required_providers {"
echo "    n8n = {"
echo "      source = \"pinotelio/n8n\""
echo "      version = \"~> 0.1\""
echo "    }"
echo "  }"
echo "}"
echo ""
echo "provider \"n8n\" {"
echo "  endpoint = \"https://your-n8n-instance.com\""
echo "  api_key  = \"your-api-key\""
echo "}"
echo ""
echo -e "${GREEN}Happy automating! 🚀${NC}"

