#!/bin/bash
set -e

AUR_DIR="${AUR_DIR:-../lota-aur}"
VERSION="{{version}}"

if [ ! -d "$AUR_DIR" ]; then
    echo "AUR directory not found: $AUR_DIR"
    exit 1
fi

cd "$AUR_DIR"

# Clean old build artifacts
rm -rf lota/ src/ pkg/

# Update .SRCINFO with new version
makepkg --printsrcinfo > .SRCINFO

# Commit and push to AUR
git add PKGBUILD .SRCINFO
git commit -m "upgpkg: lota ${VERSION}-1"
git push origin master

echo "AUR updated to version ${VERSION}"
