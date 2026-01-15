# Flux Relay CLI

A command-line interface for managing Flux Relay messaging platform.

## Version

**v1.0.0** - Initial release with authentication system

> **Note:** This workflow is being tested via GitHub Actions

## Installation

### 🚀 Quick Install

**First time install:**

**If you have Go:**
```bash
go install github.com/postacksol/flux-relay-cli@latest
```
> **Note:** Go requires the full module path. After first install, you can use `flux-relay install` for updates.

**Windows (PowerShell one-liner):**
```powershell
irm https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 | iex
```

**Linux/macOS (Bash one-liner):**
```bash
curl -fsSL https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.sh | bash
```

**Update existing installation:**
```bash
flux-relay install
```

📖 **For detailed installation instructions, see [INSTALL.md](INSTALL.md)**

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

- `flux-relay install` - Install or update the CLI
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
