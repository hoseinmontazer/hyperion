#!/usr/bin/env bash
set -e

# Variables
REPO_OWNER="hoseinmontazer"
REPO_NAME="hyperion"
CONFIG_DIR="$HOME/.hyperion"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
INSTALL_DIR="/usr/local/bin"

# Fetch the latest release information from GitHub API
echo "Fetching latest release information..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
TAG_NAME=$(echo "$LATEST_RELEASE" | jq -r '.tag_name')
ASSET_URL=$(echo "$LATEST_RELEASE" | jq -r '.assets[] | select(.name | test("hyperion-linux-amd64.*\\.tar\\.gz$")) | .browser_download_url')

# Check if the asset URL was found
if [ -z "$ASSET_URL" ]; then
  echo "Error: No suitable asset found for Linux 64-bit."
  exit 1
fi

# Download and extract the asset
TMP_DIR=$(mktemp -d)
echo "Downloading Hyperion $TAG_NAME..."
curl -L "$ASSET_URL" -o "$TMP_DIR/hyperion.tar.gz"
echo "Extracting..."
tar -xzf "$TMP_DIR/hyperion.tar.gz" -C "$TMP_DIR"

# Install the binary
echo "Installing Hyperion..."
sudo mv "$TMP_DIR/hyperion" "$INSTALL_DIR/hyperion"
sudo chmod +x "$INSTALL_DIR/hyperion"

# Create config directory and sample config if not exists
if [ ! -d "$CONFIG_DIR" ]; then
  echo "Creating default config at $CONFIG_FILE..."
  mkdir -p "$CONFIG_DIR"
  cat <<EOL > "$CONFIG_FILE"
# Hyperion config example
esxi:
  - host: "192.168.1.100"
    username: "root"
    password_env: "ESXI_PASS"
    insecure: true
EOL
fi

# Clean up
rm -rf "$TMP_DIR"

echo "Hyperion $TAG_NAME installed successfully!"
echo "You can now run it using: hyperion serve"
