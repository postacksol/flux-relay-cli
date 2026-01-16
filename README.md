# Flux Relay CLI

<div align="center">

![Flux Relay CLI](./flux-relay-logo-white.svg)

**A comprehensive command-line interface for managing the Flux Relay messaging platform, including database operations, nameserver management, and interactive SQL queries.**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Version](https://img.shields.io/badge/Version-1.0.0-blue)](https://github.com/postacksol/flux-relay-cli)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Commands Reference](#commands-reference)
- [Interactive SQL Shell](#interactive-sql-shell)
- [Nameserver Management](#nameserver-management)
- [Table Management](#table-management)
- [Multi-Tenant Isolation](#multi-tenant-isolation)
- [Project Structure](#project-structure)
- [Development](#development)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)

---

## Overview

Flux Relay CLI is a powerful command-line tool for managing your Flux Relay messaging platform. It provides a comprehensive interface for database operations, nameserver management, and SQL query execution. Built with Go and featuring an interactive SQL shell similar to Turso, the CLI enables developers to efficiently manage their messaging infrastructure from the terminal.

The CLI follows a hierarchical structure: **Projects → Servers → Nameservers**, allowing you to organize and manage your messaging platform resources with context-aware commands.

---

## Features

### Authentication & Security
- **OAuth 2.0 Device Authorization Grant**: Secure authentication flow
- **Automatic Browser Opening**: Seamless login experience
- **Headless Mode**: Support for WSL, SSH, and CI/CD environments
- **Token Management**: Secure token storage and validation
- **Session Management**: Automatic detection of existing sessions

### Project Management
- **List Projects**: View all projects in your account
- **Select Projects**: Switch between projects by name or ID
- **Project Context**: Maintain current project selection across commands

### Server Management
- **List Servers**: View all servers in a project with nameserver counts
- **Select Servers**: Switch between servers by name or ID
- **Server Shell**: Interactive SQL shell for server-level operations
- **Context Awareness**: Commands automatically use selected server

### Nameserver Management
- **List Nameservers**: View all nameservers (databases) in a server
- **Create Nameservers**: Create new nameserver instances
- **Initialize Schema**: Set up default messaging platform tables
- **Select Nameservers**: Switch between nameservers by name or ID
- **Nameserver Shell**: Interactive SQL shell for nameserver-specific operations

### Interactive SQL Shell
- **Turso-like Interface**: Familiar shell experience for SQL operations
- **Multi-line Queries**: Support for complex SQL statements
- **Query History**: Access previous queries
- **Table Inspection**: List tables, view schemas, and explore structure
- **Context Switching**: Switch between nameservers within the shell
- **Helper Commands**: Built-in commands for common operations

### Database Operations
- **SQL Query Execution**: Execute SELECT, INSERT, UPDATE, DELETE queries
- **DDL Operations**: Create, alter, and drop tables
- **Schema Customization**: Modify default tables and create custom tables
- **Data Type Changes**: Support for complex schema migrations
- **Table Validation**: Automatic validation of table names and operations

### Multi-Tenant Isolation
- **Complete Data Isolation**: Each user's data is isolated by server_id
- **Secure Table Operations**: Table validation ensures operations only affect your data
- **Nameserver Scoping**: All operations are scoped to your nameservers
- **Cross-User Protection**: Prevents access to other users' data

---

## Prerequisites

Before installing Flux Relay CLI, ensure you have:

- **Go 1.21+** (for building from source)
- **Access to Flux Relay API server** (default: `http://localhost:3000`)
- **Git** (for cloning the repository)
- **Terminal/Command Prompt** with appropriate permissions

For Windows users:
- **PowerShell 5.1+** or **PowerShell Core 7+**

For Linux/macOS users:
- **Bash** shell

---

## Installation

### Quick Install

**Windows (PowerShell one-liner):**
```powershell
irm https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 | iex
```

**Linux/macOS (Bash one-liner):**
```bash
curl -fsSL https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.sh | bash
```

**Using Go (if you have Go installed):**
```bash
go install github.com/postacksol/flux-relay-cli@latest
```

> **Note:** Go requires the full module path. After first install, you can use `flux-relay install` for updates.

### Update Existing Installation

```bash
flux-relay install
```

### Manual Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/postacksol/flux-relay-cli.git
   cd flux-relay-cli
   ```

2. **Build from source**
   ```bash
   go build -o flux-relay .
   ```

3. **Add to PATH** (optional)
   - **Linux/macOS**: `sudo mv flux-relay /usr/local/bin/`
   - **Windows**: Add the directory containing `flux-relay.exe` to your PATH

For detailed installation instructions, see [INSTALL.md](INSTALL.md) or [INSTALL_WINDOWS.md](INSTALL_WINDOWS.md).

---

## Quick Start

### 1. Authentication

**Normal login (opens browser):**
```bash
flux-relay login
```

**Headless mode (for WSL/SSH):**
```bash
flux-relay login --headless
# Visit the URL shown, then set the token:
flux-relay config set token "YOUR_TOKEN"
```

**Check current session:**
```bash
flux-relay login
# Shows current user if already authenticated
```

**Logout:**
```bash
flux-relay logout
```

### 2. Select Project and Server

```bash
# List all projects
flux-relay pr list

# Select a project
flux-relay pr MyProject        # By name
flux-relay pr 56OSXXQH         # By ID

# List servers in project
flux-relay server list

# Select a server
flux-relay server MyServer     # By name
flux-relay server 6BDJ4YBK     # By ID
```

### 3. Open Interactive SQL Shell

```bash
# Server-level shell
flux-relay server shell MyServer

# Or nameserver-level shell
flux-relay ns shell db
```

### 4. Execute SQL Queries

```bash
# One-off query
flux-relay sql "SELECT * FROM conversations_db WHERE server_id = ? LIMIT 10"

# In interactive shell
→ SELECT * FROM conversations_db WHERE server_id = ? LIMIT 5;
→ .tables
→ .quit
```

---

## Configuration

### Config File Location

- **Linux/macOS**: `~/.flux-relay/config.json`
- **Windows**: `%USERPROFILE%\.flux-relay\config.json`

### Environment Variables

- `FLUX_RELAY_API_URL`: API base URL (default: `http://localhost:3000`)
- `FLUX_RELAY_CONFIG`: Custom config file path

### Command-Line Flags

- `--api-url <url>`: Override API base URL
- `--config <path>`: Use custom config file
- `--verbose, -v`: Enable verbose output

### Manual Token Configuration

```bash
flux-relay config set token "YOUR_ACCESS_TOKEN"
```

---

## Commands Reference

### Authentication Commands

| Command | Description |
|--------|-------------|
| `flux-relay login` | Authenticate with Flux Relay (opens browser) |
| `flux-relay login --headless` | Headless authentication mode |
| `flux-relay logout` | Log out and remove stored token |
| `flux-relay config set token <token>` | Set access token manually |

### Project Commands

| Command | Description |
|--------|-------------|
| `flux-relay pr list` | List all projects in your account |
| `flux-relay pr <name-or-id>` | Select a project to work with |
| `flux-relay pr` | Show currently selected project |

### Server Commands

| Command | Description |
|--------|-------------|
| `flux-relay server list` | List all servers in the selected project |
| `flux-relay server <name-or-id>` | Select a server |
| `flux-relay server` | Show currently selected server |
| `flux-relay server shell <name-or-id>` | Open interactive SQL shell for a server |
| `flux-relay srv` | Alias for `server` command |

### Nameserver Commands

| Command | Description |
|--------|-------------|
| `flux-relay ns list` | List all nameservers in the selected server |
| `flux-relay ns <name-or-id>` | Select a nameserver |
| `flux-relay ns` | Show currently selected nameserver |
| `flux-relay ns shell <name-or-id>` | Open interactive SQL shell for a nameserver |

### SQL Commands

| Command | Description |
|--------|-------------|
| `flux-relay sql <query>` | Execute a single SQL query on the selected server/nameserver |

### Utility Commands

| Command | Description |
|--------|-------------|
| `flux-relay install` | Install or update the CLI |
| `flux-relay --version` | Show version information |
| `flux-relay --help` | Show help message |

---

## Interactive SQL Shell

The interactive SQL shell provides a Turso-like experience for executing SQL queries and managing your database.

### Starting the Shell

```bash
# Server-level shell
flux-relay server shell MyServer

# Nameserver-level shell
flux-relay ns shell db
```

### Shell Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `.help` | `.h` | Show help message |
| `.examples` | `.ex` | Show example queries and operations |
| `.quit` | `.exit`, `.q` | Exit the shell |
| `.clear` | `.c` | Clear the current query |
| `.context` | `.ctx` | Show current context (server/nameserver) |
| `.tables` | | List all tables |
| `.schema <table>` | | Show schema for a table |
| `.nameservers` | `.ns` | List available nameservers |
| `.use <nameserver>` | | Switch to a nameserver context |
| `.create_ns <name>` | | Create a new nameserver |
| `.init_ns <name>` | | Initialize schema for a nameserver |
| `.drop_table <name>` | | Drop a table (with confirmation) |

### Example Shell Session

```
Welcome to Flux Relay SQL shell!
Type ".quit" to exit the shell and ".help" to list all available commands.

📌 Server context: TestServer (6BDJ4YBK)
📌 Nameserver context: db2

→ SELECT * FROM conversations_db2 WHERE server_id = ? LIMIT 5;
id          server_id    title              created_at
──          ─────────    ─────              ───────────
conv_1      6BDJ4YBK     Test Conversation  2024-01-01T10:00:00Z

Rows returned: 1 (15ms)

→ .tables
Showing tables for 1 nameserver(s) in this server:
db2
name
──
conversations_db2
end_users_db2
messages_db2
files_db2
custom_products_db2

Rows returned: 5 (12ms)

→ .quit
Goodbye!
```

### SQL Query Features

- **Multi-line Queries**: Press Enter to continue on next line, semicolon or double Enter to execute
- **Parameterized Queries**: Use `?` for server_id parameters (automatically injected)
- **Query History**: Access previous queries with arrow keys
- **Auto-completion**: Tab completion for table names and commands

---

## Nameserver Management

Nameservers (databases) are logical namespaces within a server that organize your data.

### Creating Nameservers

**In the shell:**
```bash
→ .create_ns db2
Creating nameserver 'db2'...
✅ Nameserver 'db2' created successfully! ID: 6BDJ5ABC
```

**Using CLI commands:**
```bash
# After creating via shell, you can switch to it
flux-relay ns db2
```

### Initializing Schema

Initialize default messaging platform tables:

```bash
→ .init_ns db2
Initializing schema for nameserver 'db2'...
✅ Schema initialized successfully!
Tables created: 6
Verified tables: conversations_db2, end_users_db2, messages_db2, files_db2
```

### Switching Nameserver Context

```bash
→ .use db2
Switched to nameserver: db2

→ .context
📌 Server: TestServer (6BDJ4YBK)
📌 Nameserver: db2 (6BDJ5ABC)
```

### Listing Nameservers

```bash
→ .nameservers
Active nameservers:
  db2 (6BDJ5ABC) [current]
  name1 (6BDJ6DEF)

Inactive nameservers:
  old_db (6BDJ7GHI)
```

---

## Table Management

Tables in Flux Relay follow a naming pattern: `{baseName}_{nameserverName}`.

### Creating Tables

**Default tables** (created by `.init_ns`):
- `conversations_{nameserver}`
- `end_users_{nameserver}`
- `messages_{nameserver}`
- `files_{nameserver}`

**Custom tables:**
```sql
CREATE TABLE custom_products_db2 (
  id TEXT PRIMARY KEY,
  server_id TEXT NOT NULL,
  name TEXT,
  price REAL,
  created_at TEXT NOT NULL
);
```

### Altering Tables

**Add a column:**
```sql
ALTER TABLE conversations_db2 ADD COLUMN priority INTEGER;
```

**Rename a column:**
```sql
ALTER TABLE conversations_db2 RENAME COLUMN priority TO importance;
```

**Change data type** (requires table recreation):
```sql
-- Step 1: Create new table
CREATE TABLE conversations_db2_new (
  id TEXT PRIMARY KEY,
  server_id TEXT NOT NULL,
  priority INTEGER,  -- Changed from TEXT
  created_at TEXT NOT NULL
);

-- Step 2: Copy data with type conversion
INSERT INTO conversations_db2_new 
SELECT id, server_id, CAST(priority AS INTEGER), created_at
FROM conversations_db2 WHERE server_id = ?;

-- Step 3: Drop old table
DROP TABLE conversations_db2;

-- Step 4: Rename new table
ALTER TABLE conversations_db2_new RENAME TO conversations_db2;
```

### Dropping Tables

```bash
→ .drop_table custom_products_db2
⚠️  Warning: This will permanently delete the table 'custom_products_db2' and all its data.
Are you sure? (yes/no): yes
✅ Table 'custom_products_db2' dropped successfully
```

Or using SQL:
```sql
DROP TABLE custom_products_db2;
```

### Table Naming Rules

- Tables must follow pattern: `{baseName}_{nameserverName}`
- Base name can be any valid identifier (e.g., `conversations`, `custom_products`)
- Nameserver name must match an existing nameserver in your server
- Temporary tables for migrations can use suffixes: `_new`, `_old`, `_temp`, `_backup`, `_migrated`

---

## Multi-Tenant Isolation

Flux Relay CLI ensures complete data isolation between users through multiple security layers.

### Isolation Layers

1. **Project-level**: Users can only access their own projects
2. **Server-level**: Users can only access servers in their projects
3. **Nameserver-level**: Table validation only checks nameservers in your server
4. **Data-level**: All queries are automatically filtered by `server_id`

### How It Works

- **Automatic server_id Filtering**: All SELECT/UPDATE/DELETE queries require `WHERE server_id = ?`
- **Table Validation**: DDL operations only validate against your own nameservers
- **Cross-User Protection**: Even with the same nameserver/table names, users cannot access each other's data

### Example

User A and User B can both have:
- Nameserver named `db2`
- Table named `conversations_db2`

But they are completely isolated because:
- Different `server_id` values
- Different `project_id` values
- All queries filtered by their own `server_id`

---

## Project Structure

```
flux-relay-cli/
├── cmd/                    # Command implementations
│   ├── config.go          # Configuration management
│   ├── install.go         # Installation command
│   ├── login.go           # Authentication
│   ├── logout.go          # Logout command
│   ├── ns.go              # Nameserver commands
│   ├── pr.go              # Project commands
│   ├── projects.go        # Project listing
│   ├── root.go            # Root command and flags
│   ├── server.go          # Server commands
│   ├── shell.go           # Interactive SQL shell
│   └── sql.go             # SQL query execution
├── internal/              # Internal packages
│   ├── api/               # API client
│   │   └── client.go      # HTTP client implementation
│   └── config/            # Configuration storage
│       └── storage.go     # Config file management
├── main.go                # Entry point
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile               # Build automation
├── install.sh             # Linux/macOS installer
├── install.ps1            # Windows installer
└── README.md              # This file
```

---

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/postacksol/flux-relay-cli.git
cd flux-relay-cli

# Download dependencies
go mod download

# Build
go build -o flux-relay .

# Run tests (if available)
go test ./...
```

### Running Locally

```bash
# Run a command
go run main.go login

# Or build and run
go build -o flux-relay .
./flux-relay login
```

### Development Requirements

- Go 1.21 or higher
- Access to Flux Relay API server
- Git for version control

### Code Style

- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Write meaningful commit messages
- Add comments for exported functions and types

---

## Contributing

Contributions are welcome! If you have access to the repository:

1. **Fork the repository** (if applicable)
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Test thoroughly**
5. **Commit your changes** (`git commit -m 'Add some amazing feature'`)
6. **Push to the branch** (`git push origin feature/amazing-feature`)
7. **Open a Pull Request**

### Contribution Guidelines

- Follow existing code style and patterns
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass
- Write clear commit messages

---

## Support

For support and questions:

- **Documentation**: See [INSTALL.md](INSTALL.md) for installation help
- **Issues**: Report issues via GitHub Issues (if repository is public)
- **Email**: Contact the development team for private support

### Common Issues

**Login fails in headless mode:**
- Ensure you visit the URL and copy the token correctly
- Check that the token is set: `flux-relay config set token "YOUR_TOKEN"`

**Cannot find project/server:**
- Verify you're logged in: `flux-relay login`
- Check project selection: `flux-relay pr`
- List available projects: `flux-relay pr list`

**SQL query errors:**
- Ensure you include `WHERE server_id = ?` in SELECT/UPDATE/DELETE queries
- Verify table names follow the pattern: `{baseName}_{nameserverName}`
- Use `.tables` to see available tables

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

<div align="center">

**Built with Go by Postack Solutions**

[GitHub](https://github.com/postacksol/flux-relay-cli) • [Documentation](./INSTALL.md) • [Support](#support)

</div>
