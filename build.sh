#!/bin/bash
# Recode agent builder
set -euo pipefail

log () {
  echo -e "${1}" >&2
}

SUPPORTED_PLATFORMS=("linux/amd64" "linux/386" "linux/arm" "linux/arm64")

for platform in "${SUPPORTED_PLATFORMS[@]}"
do
	platform_parts=(${platform//\// })
	platform_os="${platform_parts[0]}"
	platform_arch="${platform_parts[1]}"
	bin_name="agent-${platform_os}-${platform_arch}"

  log "Building agent for ${platform_os}/${platform_arch}..."

	env GOOS="${platform_os}" GOARCH="${platform_arch}" go build -o "out/${bin_name}" main.go
	if [ $? -ne 0 ]; then
    log "An error occured duing build for ${platform_os}/${platform_arch}!"
		exit 1
	fi
done