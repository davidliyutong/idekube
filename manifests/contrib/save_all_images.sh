#!/usr/bin/env bash
set -euo pipefail

# Batch-save images for each subdirectory under this contrib folder.
# For each subdir containing an images.txt, call save_images.sh to pull and save
# those images into a tar.gz named <subdir>-images.tar.gz.
#
# Optional env vars:
#   PLATFORM     -> e.g. linux/amd64, passed through to save_images.sh
#   OUTPUT_DIR   -> directory to store exported tarballs (default: ./_images)
#   CONTAINER_CLI-> docker or podman (handled by save_images.sh)

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAVE_SCRIPT="${BASE_DIR}/save_images.sh"

if [[ ! -f "${SAVE_SCRIPT}" ]]; then
  echo "Error: save_images.sh not found at ${SAVE_SCRIPT}" >&2
  exit 1
fi

platform="${PLATFORM:-}"
output_dir="${OUTPUT_DIR:-${BASE_DIR}/_images}"
mkdir -p "${output_dir}"

echo "Using output directory: ${output_dir}"
if [[ -n "${platform}" ]]; then
  echo "Target platform: ${platform}"
fi

shopt -s nullglob
for dir in "${BASE_DIR}"/*/; do
  [[ -d "${dir}" ]] || continue
  images_file="${dir}images.txt"

  if [[ -f "${images_file}" ]]; then
    name="$(basename "${dir%/}")"
    out_tar="${output_dir}/${name}-images.tar.gz"
    echo "Processing ${name} (list: ${images_file}) -> ${out_tar}"

    if [[ -n "${platform}" ]]; then
      bash "${SAVE_SCRIPT}" -l "${images_file}" -i "${out_tar}" -p "${platform}"
    else
      bash "${SAVE_SCRIPT}" -l "${images_file}" -i "${out_tar}"
    fi
  else
    echo "Skipping $(basename "${dir%/}"): no images.txt"
  fi
done

echo "All done. Tarballs are in: ${output_dir}"
