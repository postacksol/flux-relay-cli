# Flux Relay CLI v1.0.0 Release

## Installation

### From Source

```bash
git clone https://github.com/fluxrelay/flux-relay-cli.git
cd flux-relay-cli
go build -o flux-relay .
```

### Binary Releases

Download the appropriate binary for your platform from the [Releases](https://github.com/fluxrelay/flux-relay-cli/releases) page.

## Quick Start

1. **Login:**
   ```bash
   flux-relay login
   ```

2. **Headless Login (for WSL/headless environments):**
   ```bash
   flux-relay login --headless
   # Visit the URL shown, then run:
   flux-relay config set token "YOUR_TOKEN"
   ```

3. **Check Status:**
   ```bash
   flux-relay login  # Shows current user if already logged in
   ```

4. **Logout:**
   ```bash
   flux-relay logout
   ```

## Features

- ✅ Device code authentication flow
- ✅ Automatic browser opening
- ✅ Headless mode for WSL/SSH environments
- ✅ Token validation and storage
- ✅ Already logged in detection
- ✅ Dad jokes while waiting (optional)

## Configuration

Tokens are stored in `~/.flux-relay/config.json` with secure permissions (0600).

## Requirements

- Go 1.21+ (for building from source)
- Access to Flux Relay API (default: http://localhost:3000)
