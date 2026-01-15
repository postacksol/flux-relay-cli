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
    echo "Warning: Could not find binary for $OS/$ARCH"
    echo "The release may not have binaries yet. Building from source..."
    echo ""
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed. Please install Go first:"
        echo "  https://go.dev/dl/"
        echo ""
        echo "Or download a pre-built binary from:"
        echo "  https://github.com/postacksol/flux-relay-cli/releases"
        exit 1
    fi
    
    echo "Go found: $(go version)"
    echo ""
    
    # Create temporary directory for building
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    echo "Cloning repository..."
    if ! git clone --depth 1 https://github.com/postacksol/flux-relay-cli.git "$TEMP_DIR" 2>/dev/null; then
        echo "Error: Failed to clone repository. Make sure Git is installed."
        echo "  https://git-scm.com/download/"
        exit 1
    fi
    
    echo "Building from source..."
    cd "$TEMP_DIR"
    if ! go build -o flux-relay .; then
        echo "Error: Build failed"
        exit 1
    fi
    
    echo "Build successful!"
    echo ""
    
    # Set binary path to built file
    BUILT_BIN="$TEMP_DIR/flux-relay"
    
    # Determine install location
    INSTALL_DIR="$HOME/.flux-relay/bin"
    BIN_PATH="$INSTALL_DIR/flux-relay"
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Copy built binary to install location
    echo "Installing binary..."
    cp "$BUILT_BIN" "$BIN_PATH"
    chmod +x "$BIN_PATH"
    
    echo "Installed successfully!"
    echo ""
    
    # Skip to PATH section
    SKIP_DOWNLOAD=true
else
    SKIP_DOWNLOAD=false
fi

if [ "$SKIP_DOWNLOAD" != "true" ]; then
    # Determine install location
    INSTALL_DIR="$HOME/.flux-relay/bin"
    BIN_PATH="$INSTALL_DIR/flux-relay"
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Download binary
    echo "Downloading binary..."
    curl -L -o "$BIN_PATH" "$ASSET_URL"
    chmod +x "$BIN_PATH"
else
    # If we built from source, INSTALL_DIR and BIN_PATH are already set above
    # Just make sure they're set
    if [ -z "$INSTALL_DIR" ]; then
        INSTALL_DIR="$HOME/.flux-relay/bin"
        BIN_PATH="$INSTALL_DIR/flux-relay"
    fi
fi

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
