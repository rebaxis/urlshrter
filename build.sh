#!/bin/bash

# Build script для установки версии, даты и коммита через ldflags

VERSION=${VERSION:-"v1.0.0"}
BUILD_DATE=$(date +'%Y/%m/%d %H:%M:%S')
BUILD_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building with:"
echo "  Version: $VERSION"
echo "  Date: $BUILD_DATE"
echo "  Commit: $BUILD_COMMIT"
echo ""

go build -ldflags "\
  -X 'main.buildVersion=$VERSION' \
  -X 'main.buildDate=$BUILD_DATE' \
  -X 'main.buildCommit=$BUILD_COMMIT'" \
  -o shortener.exe \
  ./cmd/shortener

echo "Build complete: shortener.exe"
