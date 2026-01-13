#!/bin/bash

set -e

# Default to interactive mode
QUIET=false

# Check arguments
if [[ "$1" == "-q" ]]; then
    QUIET=true
fi

# Function for confirmation
confirm() {
    local msg="$1"
    if [ "$QUIET" = true ]; then
        echo "Auto-confirming: $msg"
        return 0
    fi
    read -p "$msg [y/N] " response
    case "$response" in
        [yY][eE][sS]|[yY]) 
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Ensure running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

echo "Scanning for unused disks..."
UNUSED_DISKS=()

# Get all block devices of type disk
for dev_name in $(lsblk -d -n -o NAME); do
    dev_path="/dev/$dev_name"
    
    # 1. Check if mounted (check the device or any of its children/partitions)
    # lsblk output for MOUNTPOINT is empty if not mounted
    mountpoints=$(lsblk "$dev_path" -n -o MOUNTPOINT)
    if [ -n "${mountpoints//[[:space:]]/}" ]; then 
        # If variable contains any non-whitespace chars, it's mounted
        continue
    fi

    # 2. Check for partitions
    # List types of children. If 'part' exists, skip.
    children_types=$(lsblk "$dev_path" -n -o TYPE)
    if echo "$children_types" | grep -q "part"; then
        continue
    fi
    
    # 3. Optional: Check if it's held by other things like LVM (type lvm)
    if echo "$children_types" | grep -q "lvm"; then
        continue
    fi
    
    # Additional check: ensure it is a disk, not a loop device or rom if lsblk -d failed filtering
    if [[ $dev_name == loop* ]] || [[ $dev_name == sr* ]]; then
        continue
    fi

    UNUSED_DISKS+=("$dev_path")
done

if [ ${#UNUSED_DISKS[@]} -eq 0 ]; then
    echo "No unused disks found."
    exit 0
fi

echo "Found unused disks:"
for disk in "${UNUSED_DISKS[@]}"; do
    # Get generic info for display
    size=$(lsblk -d -n -o SIZE "$disk")
    model=$(lsblk -d -n -o MODEL "$disk")
    echo "  - $disk (Size: $size, Model: $model)"
done

echo ""
if ! confirm "Do you want to proceed with formatting and mounting these disks?"; then
    echo "Aborted by user."
    exit 0
fi

# Stage 2 & 3: Format and Mount
for disk in "${UNUSED_DISKS[@]}"; do
    echo "Processing $disk..."
    
    if confirm "Format $disk as XFS?"; then
        echo "Formatting $disk..."
        mkfs.xfs -f "$disk"
    else
        echo "Skipping format for $disk."
    fi
    
    # Get UUID
    uuid=$(blkid -s UUID -o value "$disk")
    if [ -z "$uuid" ]; then
        echo "Error: Could not obtain UUID for $disk. Was it formatted correctly?"
        continue
    fi
    
    dev_name=$(basename "$disk")
    mount_point="/mnt/longhorn/${dev_name}"
    
    if confirm "Mount $disk to $mount_point?"; then
        echo "Creating mount point $mount_point..."
        mkdir -p "$mount_point"
        
        echo "Mounting..."
        mount "$disk" "$mount_point"
        
        echo "Mounted successfully."
        
        # Add to fstab
        if confirm "Add to /etc/fstab?"; then
             # Check if already in fstab to avoid duplicates
             if grep -q "$uuid" /etc/fstab; then
                 echo "Entry for UUID=$uuid already exists in /etc/fstab."
             else
                 echo "Adding to /etc/fstab..."
                 echo "UUID=$uuid $mount_point xfs defaults 0 0" >> /etc/fstab
                 echo "Added to /etc/fstab."
             fi
        fi
    else
        echo "Skipping mount for $disk."
    fi
done

echo "Done."
