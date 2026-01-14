#!/bin/bash
# Flux Relay CLI Installer for Linux/macOS
# Usage: ./install.sh

set -e

echo "Flux Relay CLI Installer"
echo "========================"
echo ""

# Detect OS and architecture
OS="linux"
ARCH="amd64"

if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
fi

if [[ $(uname -m) == "arm64" ]] || [[ $(uname -m) == "aarch64" ]]; then
    ARCH="arm64"
fi

# Get latest release
echo "Fetching latest release..."
RELEASE=$(curl -s https://api.github.com/repos/postacksol/flux-relay-cli/releases/latest)
VERSION=$(echo $RELEASE | grep -oP '"tag_name": "\K[^"]+')

if [ -z "$VERSION" ]; then
    echo "Error: Could not fetch latest release. Using v1.0.0"
    VERSION="v1.0.0"
else
    echo "Latest version: $VERSION"
fi

# Find appropriate binary
ASSET_URL=$(echo $RELEASE | grep -oP '"browser_download_url": "\K[^"]*' | grep -i "$OS" | grep -i "$ARCH" | head -1)

if [ -z "$ASSET_URL" ]; then
    echo "Error: Could not find binary for $OS/$ARCH"
    exit 1
fi

# Determine install location
INSTALL_DIR="$HOME/.flux-relay/bin"
BIN_PATH="$INSTALL_DIR/flux-relay"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download binary
echo "Downloading binary..."
curl -L -o "$BIN_PATH" "$ASSET_URL"
chmod +x "$BIN_PATH"

# Add to PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Adding to PATH..."
    
    # Detect shell
    if [[ "$SHELL" == *"zsh"* ]]; then
        SHELL_RC="$HOME/.zshrc"
    else
        SHELL_RC="$HOME/.bashrc"
    fi
    
    echo "" >> "$SHELL_RC"
    echo "# Flux Relay CLI" >> "$SHELL_RC"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
    
    echo "Added $INSTALL_DIR to PATH in $SHELL_RC"
    echo "Run: source $SHELL_RC (or restart your terminal)"
else
    echo "Already in PATH"
fi

echo ""
echo "Installation complete!"
echo "Binary installed to: $BIN_PATH"
echo ""
echo "Try running: flux-relay --version"
