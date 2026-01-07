#!/bin/bash

# RKE2 Airgap Installation Download Script
# This script downloads RKE2 images and binaries for airgap installation

set -e

# Variables
RKE2_VERSION="v1.33.1+rke2r1"
ARCH="arm64"
CNI_PLUGIN="cilium"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOWNLOAD_DIR="$SCRIPT_DIR/files"

# Create download directory
echo "Creating download directory: $DOWNLOAD_DIR"
mkdir -p "$DOWNLOAD_DIR"
cd "$DOWNLOAD_DIR"

# Function to check if file exists and matches checksum
check_file_integrity() {
    local filename=$1
    local checksum_file="sha256sum-${ARCH}.txt"

    if [ ! -f "$filename" ]; then
        return 1  # File doesn't exist
    fi

    if [ ! -f "$checksum_file" ]; then
        return 1  # Checksum file doesn't exist, can't verify
    fi

    # Check if file is in checksum file and verify
    if grep -q "$filename" "$checksum_file"; then
        echo "  Verifying existing file: $filename"
        local expected_checksum=$(grep "$filename" "$checksum_file" | awk '{print $1}')
        local actual_checksum=$(shasum -a 256 "$filename" | awk '{print $1}')

        if [ "$expected_checksum" = "$actual_checksum" ]; then
            echo "  ✓ Checksum verified, skipping download"
            return 0
        else
            echo "  ✗ Checksum mismatch, will re-download"
            echo "    Expected: $expected_checksum"
            echo "    Actual:   $actual_checksum"
            return 1
        fi
    fi

    return 1  # File not in checksum list
}

# Download checksums for verification
echo "Downloading SHA256 checksums..."
curl -LO "https://github.com/rancher/rke2/releases/download/${RKE2_VERSION}/sha256sum-${ARCH}.txt"

# Download RKE2 images tarball
echo "Downloading RKE2 images archive..."
if ! check_file_integrity "rke2-images.linux-${ARCH}.tar.zst"; then
    curl -LO "https://github.com/rancher/rke2/releases/download/${RKE2_VERSION}/rke2-images.linux-${ARCH}.tar.zst"
fi

if ! check_file_integrity "rke2-images-${CNI_PLUGIN}.linux-${ARCH}.tar.zst"; then
    curl -LO "https://github.com/rancher/rke2/releases/download/${RKE2_VERSION}/rke2-images-${CNI_PLUGIN}.linux-${ARCH}.tar.zst"
fi

# Download RKE2 binary
echo "Downloading RKE2 binary..."
if ! check_file_integrity "rke2.linux-${ARCH}.tar.gz"; then
    curl -LO "https://github.com/rancher/rke2/releases/download/${RKE2_VERSION}/rke2.linux-${ARCH}.tar.gz"
fi

# Download RKE2 install script
echo "Downloading RKE2 install script..."
if [ ! -f "install.sh" ]; then
    curl -sfL "https://get.rke2.io" --output "install.sh"
else
    echo "  ✓ install.sh already exists, skipping download"
fi

echo ""
echo "Download completed successfully!"
echo "Files downloaded to: $DOWNLOAD_DIR"
echo ""
echo "Next steps:"
echo "1. Verify checksums: sha256sum -c sha256sum-${ARCH}.txt --ignore-missing"
echo "2. Transfer these files to your airgap nodes"
echo "3. On airgap nodes, run:"
echo "   sudo mkdir -p /var/lib/rancher/rke2/agent/images/"
echo "   sudo cp rke2-images.linux-${ARCH}.tar.zst /var/lib/rancher/rke2/agent/images/"
echo "   sudo cp rke2.linux-${ARCH} /usr/local/bin/rke2"
echo "   sudo chmod +x /usr/local/bin/rke2"
