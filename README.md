# Flux Relay CLI

A command-line interface for managing Flux Relay messaging platform.

## Version

**v1.0.0** - Initial release with authentication system

> **Note:** This workflow is being tested via GitHub Actions

## Installation

### Build from Source

```bash
git clone https://github.com/fluxrelay/flux-relay-cli.git
cd flux-relay-cli
go build -o flux-relay .
```

### Using Go Install

```bash
go install github.com/fluxrelay/flux-relay-cli@v1.0.0
```

## Quick Start

### Authentication

1. **Login (normal mode):**
   ```bash
   flux-relay login
   ```
   This will open your browser for authentication.

2. **Login (headless mode - for WSL/SSH):**
   ```bash
   flux-relay login --headless
   ```
   Visit the URL shown, then set the token:
   ```bash
   flux-relay config set token "YOUR_TOKEN"
   ```

3. **Check if already logged in:**
   ```bash
   flux-relay login
   ```
   Shows current user if already authenticated.

4. **Logout:**
   ```bash
   flux-relay logout
   ```

## Commands

- `flux-relay login` - Authenticate with Flux Relay
- `flux-relay login --headless` - Headless authentication mode
- `flux-relay logout` - Log out and remove stored token
- `flux-relay config set token <token>` - Set access token manually

## Configuration

- Config location: `~/.flux-relay/config.json`
- API URL: Defaults to `http://localhost:3000` (can be set via `--api-url` flag)

## Features

- ✅ OAuth 2.0 Device Authorization Grant flow
- ✅ Automatic browser opening
- ✅ Headless mode for WSL/SSH environments
- ✅ Token validation and secure storage
- ✅ Already logged in detection
- ✅ Dad jokes while waiting (optional entertainment)

## Requirements

- Go 1.21+ (for building from source)
- Access to Flux Relay API server

## Development

```bash
go mod download
go run main.go login
```

## License

MIT
