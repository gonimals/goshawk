#!/bin/bash

set -e

# Check for root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (or via sudo)"
  exit 1
fi

echo "Installing Goshawk..."

# Determine architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64)
    ARCH_REGEX="(amd64|x86_64)"
    ;;
  aarch64|arm64)
    ARCH_REGEX="(arm64|aarch64)"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "Fetching latest release information..."
RELEASE_JSON=$(curl -sL https://api.github.com/repos/gonimals/goshawk/releases/latest)
URL=$(echo "$RELEASE_JSON" | grep -iE "browser_download_url.*linux.*${ARCH_REGEX}.*\.tar\.gz" | cut -d '"' -f 4 | head -n 1)

if [ -z "$URL" ]; then
    echo "Could not find a pre-compiled binary for architecture $ARCH."
    echo "Please compile manually or check the releases page."
    exit 1
fi

TMP_DIR=$(mktemp -d)
echo "Downloading $URL..."
curl -sL "$URL" -o "$TMP_DIR/goshawk.tar.gz"

echo "Extracting..."
tar -xzf "$TMP_DIR/goshawk.tar.gz" -C "$TMP_DIR"

echo "Installing binary to /usr/local/sbin/goshawk"
mkdir -p /usr/local/sbin
cp "$TMP_DIR/goshawk" /usr/local/sbin/goshawk
chmod +x /usr/local/sbin/goshawk

rm -rf "$TMP_DIR"

echo "Setting up configuration file..."
if [ ! -f /etc/goshawk.yml ]; then
    echo "Downloading example config to /etc/goshawk.yml..."
    curl -sL "https://raw.githubusercontent.com/gonimals/goshawk/main/example_config.yaml" -o /etc/goshawk.yml
else
    echo "Configuration file /etc/goshawk.yml already exists."
fi

chown root:root /etc/goshawk.yml
chmod 0600 /etc/goshawk.yml

if [ -d /etc/systemd/system ]; then
    echo "Setting up systemd service..."
    curl -sL "https://raw.githubusercontent.com/gonimals/goshawk/main/deploy/goshawk.service" -o /etc/systemd/system/goshawk.service

    systemctl daemon-reload
else
    echo "Systemd configuration directory not found. Skipping service setup."
fi

echo "Installation complete!"
echo "Please edit /etc/goshawk.yml with your desired configuration."
echo "Then, you can start the service using: systemctl start goshawk.service"
