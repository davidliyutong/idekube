#!/usr/bin/env bash

# Script to generate API client from swagger.json
# This script handles Java environment issues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating API client from swagger.json...${NC}"

# Check if swagger.json exists
SWAGGER_FILE="../controller/docs/api/swagger.json"
if [ ! -f "$SWAGGER_FILE" ]; then
    echo -e "${RED}Error: swagger.json not found at $SWAGGER_FILE${NC}"
    echo "Please make sure the controller documentation is generated first."
    exit 1
fi

# Try to find Java installation
JAVA_CMD=""

# First, try /usr/libexec/java_home (macOS)
if command -v /usr/libexec/java_home &> /dev/null; then
    JAVA_HOME=$(/usr/libexec/java_home 2>/dev/null)
    if [ -n "$JAVA_HOME" ]; then
        export PATH="$JAVA_HOME/bin:$PATH"
        JAVA_CMD="$JAVA_HOME/bin/java"
    fi
fi

# If not found, try java in PATH
if [ -z "$JAVA_CMD" ] && command -v java &> /dev/null; then
    JAVA_CMD="java"
fi

# If still not found, error out
if [ -z "$JAVA_CMD" ]; then
    echo -e "${RED}Error: Java not found${NC}"
    echo "Please install Java (version 8 or higher) to generate API client."
    echo "On macOS: brew install openjdk"
    exit 1
fi

# Verify Java version
if ! $JAVA_CMD -version &> /dev/null; then
    echo -e "${RED}Error: Java command failed${NC}"
    exit 1
fi

echo -e "${GREEN}Using Java: $JAVA_CMD${NC}"
$JAVA_CMD -version 2>&1 | head -1

# Run the generation
echo -e "${YELLOW}Running OpenAPI Generator...${NC}"
yarn generate-api

echo -e "${GREEN}✓ API client generated successfully in src/api/client/${NC}"
