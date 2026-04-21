#!/usr/bin/env bash
set -euo pipefail

BINARY="lota"
OUTPUT_DIR="dist"
MODULE="lota"

TARGETS=(
  "linux   amd64"
  "linux   arm64"
  "windows amd64"
  "windows arm64"
  "darwin  amd64"
  "darwin  arm64"
)

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-s -w"
LDFLAGS+=" -X ${MODULE}/shared.Version=${VERSION}"
LDFLAGS+=" -X ${MODULE}/shared.Commit=${COMMIT}"
LDFLAGS+=" -X ${MODULE}/shared.BuildDate=${BUILD_DATE}"

mkdir -p "${OUTPUT_DIR}"

echo "Building ${BINARY} ${VERSION} (${COMMIT})"
echo "-------------------------------------------"

for target in "${TARGETS[@]}"; do
  OS=$(echo "${target}" | awk '{print $1}')
  ARCH=$(echo "${target}" | awk '{print $2}')

  FILE_OS="${OS}"
  [[ "${OS}" == "darwin" ]] && FILE_OS="macos"
  OUT="${OUTPUT_DIR}/${BINARY}-${FILE_OS}-${ARCH}"
  [[ "${OS}" == "windows" ]] && OUT="${OUT}.exe"

  printf "  %-10s %-8s -> %s\n" "${OS}" "${ARCH}" "${OUT}"

  GOOS="${OS}" GOARCH="${ARCH}" CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "${LDFLAGS}" \
    -o "${OUT}" \
    .
done

echo "-------------------------------------------"
echo "Creating archives..."

for target in "${TARGETS[@]}"; do
  OS=$(echo "${target}" | awk '{print $1}')
  ARCH=$(echo "${target}" | awk '{print $2}')

  FILE_OS="${OS}"
  [[ "${OS}" == "darwin" ]] && FILE_OS="macos"
  BINARY_NAME="${BINARY}-${FILE_OS}-${ARCH}"
  [[ "${OS}" == "windows" ]] && BINARY_NAME="${BINARY_NAME}.exe"

  ARCHIVE="${OUTPUT_DIR}/${BINARY_NAME}.tar.gz"

  printf "  %-10s %-8s -> %s\n" "${OS}" "${ARCH}" "${ARCHIVE}"

  (cd "${OUTPUT_DIR}" && tar -czf "${BINARY_NAME}.tar.gz" "${BINARY_NAME}")
done

echo "-------------------------------------------"
echo "Generating checksums..."
(cd "${OUTPUT_DIR}" && sha256sum *.tar.gz > checksums.txt)
echo "  checksums.txt"

echo "-------------------------------------------"
echo "Done. Artifacts in ./${OUTPUT_DIR}/"
