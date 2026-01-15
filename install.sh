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
    if ! git clone --depth 1 --branch main https://github.com/postacksol/flux-relay-cli.git "$TEMP_DIR" 2>/dev/null; then
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

# Clean up old PATH entries and add new one
echo "Checking for existing installations..."

# Detect shell
if [[ "$SHELL" == *"zsh"* ]]; then
    SHELL_RC="$HOME/.zshrc"
else
    SHELL_RC="$HOME/.bashrc"
fi

# Remove old flux-relay entries from shell RC file
if [[ -f "$SHELL_RC" ]]; then
    # Create backup
    cp "$SHELL_RC" "$SHELL_RC.backup.$(date +%s)" 2>/dev/null || true
    
    # Remove old flux-relay PATH entries
    sed -i.tmp '/# Flux Relay CLI/,/export PATH.*flux-relay/d' "$SHELL_RC" 2>/dev/null || \
    sed -i '/# Flux Relay CLI/,/export PATH.*flux-relay/d' "$SHELL_RC" 2>/dev/null || true
    rm -f "$SHELL_RC.tmp" 2>/dev/null || true
fi

# Add to PATH (always add, even if it seems to be there, to ensure it's at the end)
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Adding to PATH..."
else
    echo "Updating PATH entry..."
fi

# Remove any existing Flux Relay CLI section and add fresh one
if [[ -f "$SHELL_RC" ]]; then
    # Remove old entries
    grep -v "Flux Relay CLI" "$SHELL_RC" > "$SHELL_RC.tmp" 2>/dev/null || cp "$SHELL_RC" "$SHELL_RC.tmp"
    grep -v "flux-relay" "$SHELL_RC.tmp" > "$SHELL_RC" 2>/dev/null || cp "$SHELL_RC.tmp" "$SHELL_RC"
    rm -f "$SHELL_RC.tmp" 2>/dev/null || true
fi

# Add new entry
echo "" >> "$SHELL_RC"
echo "# Flux Relay CLI" >> "$SHELL_RC"
echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"

echo "✅ Updated PATH in $SHELL_RC"
echo "Run: source $SHELL_RC (or restart your terminal)"

echo ""
echo "Installation complete!"
echo "Binary installed to: $BIN_PATH"
echo ""
echo "Try running: flux-relay --version"
