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

### Managing Projects

1. **List all projects:**
   ```bash
   flux-relay pr list
   ```
   Shows all projects in your account with details like ID, name, description, and creation date.

2. **Select a project:**
   ```bash
   flux-relay pr MyProject        # Select by name
   flux-relay pr 56OSXXQH        # Select by ID
   flux-relay pr                  # Show current project
   ```

3. **List servers in selected project:**
   ```bash
   flux-relay server list
   # or
   flux-relay srv list
   ```
   Shows all servers in the selected project with nameserver counts.

4. **Select a server:**
   ```bash
   flux-relay server MyServer        # Select by name
   flux-relay srv server_123        # Select by ID (using alias)
   flux-relay server                 # Show current server
   ```

5. **List nameservers in selected server:**
   ```bash
   flux-relay ns list
   ```
   Shows all nameservers (databases) in the selected server.

6. **Select a nameserver:**
   ```bash
   flux-relay ns db                  # Select by name
   flux-relay ns db_123              # Select by ID
   flux-relay ns                     # Show current nameserver
   ```

7. **Execute SQL queries:**
   ```bash
   flux-relay sql "SELECT * FROM conversations_db WHERE server_id = ? LIMIT 10"
   ```
   Executes SQL queries on the selected server. Queries automatically filter by server_id.

## Commands

- `flux-relay install` - Install or update the CLI
- `flux-relay login` - Authenticate with Flux Relay
- `flux-relay login --headless` - Headless authentication mode
- `flux-relay logout` - Log out and remove stored token
- `flux-relay config set token <token>` - Set access token manually
- `flux-relay pr list` - List all projects in your account
- `flux-relay pr <project-name-or-id>` - Select a project to work with (supports names with spaces)
- `flux-relay pr` - Show currently selected project
- `flux-relay server list` or `flux-relay srv list` - List all servers in the selected project (with nameserver counts)
- `flux-relay server <server-name-or-id>` or `flux-relay srv <server-name-or-id>` - Select a server
- `flux-relay server` or `flux-relay srv` - Show currently selected server
- `flux-relay ns list` - List all nameservers in the selected server
- `flux-relay ns <nameserver-name-or-id>` - Select a nameserver
- `flux-relay ns` - Show currently selected nameserver
- `flux-relay sql <query>` - Execute SQL query on the selected server/nameserver

## Configuration

- Config location: `~/.flux-relay/config.json`
- API URL: Defaults to `http://localhost:3000` (can be set via `--api-url` flag)

## Features

- ✅ OAuth 2.0 Device Authorization Grant flow
- ✅ Automatic browser opening
- ✅ Headless mode for WSL/SSH environments
- ✅ Token validation and secure storage
- ✅ Already logged in detection
- ✅ Project management (list projects)
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
