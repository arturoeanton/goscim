#!/bin/sh
# One-line installer for GoSCIM precompiled binaries (Linux & macOS).
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/arturoeanton/goscim/main/scripts/install.sh | sh
#   GOSCIM_INSTALL_DIR=$HOME/.local/bin sh install.sh

set -eu

REPO="arturoeanton/goscim"
INSTALL_DIR="${GOSCIM_INSTALL_DIR:-${INSTALL_DIR:-/usr/local/bin}}"
VERSION="${GOSCIM_VERSION:-latest}"

uname_s="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$uname_s" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  *)
    echo "Unsupported operating system: $uname_s" >&2
    exit 1
    ;;
esac

uname_m="$(uname -m)"
case "$uname_m" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $uname_m" >&2
    exit 1
    ;;
esac

if [ "$VERSION" = "latest" ]; then
  TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
else
  TAG="$VERSION"
fi

if [ -z "$TAG" ]; then
  echo "Unable to determine latest release tag from GitHub." >&2
  exit 1
fi

VER_NUM="${TAG#v}"
ARCHIVE="goscim_${VER_NUM}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

echo "Downloading GoSCIM ${TAG} (${OS}/${ARCH})..."
if ! curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "${TMP_DIR}/${ARCHIVE}" 2>/dev/null; then
  if command -v go >/dev/null 2>&1; then
    echo "Precompiled archive not yet attached to ${TAG}; building via 'go install github.com/${REPO}@${TAG}'..."
    GOBIN="$TMP_DIR" CGO_ENABLED=0 go install "github.com/${REPO}@${TAG}"
  else
    echo "Archive ${BASE_URL}/${ARCHIVE} not found and Go toolchain is not installed." >&2
    exit 1
  fi
else
  if curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMP_DIR}/checksums.txt" 2>/dev/null; then
    EXPECTED="$(grep "  ${ARCHIVE}\$" "${TMP_DIR}/checksums.txt" | awk '{print $1}')"
    if [ -n "$EXPECTED" ]; then
      if command -v sha256sum >/dev/null 2>&1; then
        ACTUAL="$(sha256sum "${TMP_DIR}/${ARCHIVE}" | awk '{print $1}')"
      elif command -v shasum >/dev/null 2>&1; then
        ACTUAL="$(shasum -a 256 "${TMP_DIR}/${ARCHIVE}" | awk '{print $1}')"
      else
        ACTUAL="$EXPECTED"
      fi
      if [ "$ACTUAL" != "$EXPECTED" ]; then
        echo "SHA-256 checksum mismatch for ${ARCHIVE}" >&2
        exit 1
      fi
    fi
  fi
  tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "${TMP_DIR}"
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || true
if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "${TMP_DIR}/goscim" "${INSTALL_DIR}/goscim"
else
  echo "Installing into ${INSTALL_DIR} (requires sudo)..."
  sudo install -m 0755 "${TMP_DIR}/goscim" "${INSTALL_DIR}/goscim"
fi

echo "Installed GoSCIM to ${INSTALL_DIR}/goscim"
echo "Run 'goscim -demo' to launch the SCIM 2.0 server and HTMX console on http://127.0.0.1:8080/ui"
