# Flux Relay CLI - Installation Guide

## 🚀 Quick Install (Choose Your Method)

### Method 1: Go Install (Easiest - If you have Go)

**First install:**
```bash
go install github.com/postacksol/flux-relay-cli@latest
```

**Update later:**
```bash
flux-relay install
```

> **Note:** Go requires the full module path (`github.com/postacksol/flux-relay-cli@latest`) for `go install`. After the first install, you can use the simpler `flux-relay install` command for updates.

**Note:** Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

### Method 2: One-Line Installer

#### Windows (PowerShell):
```powershell
irm https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 | iex
```

#### Linux/macOS (Bash):
```bash
curl -fsSL https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.sh | bash
```

### Method 3: Download and Run Installer

#### Windows:
```powershell
iwr https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 -OutFile install.ps1
.\install.ps1
```

#### Linux/macOS:
```bash
curl -fsSL https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

## 📋 What the Installer Does

1. **Checks for pre-built binaries** from GitHub Releases
2. **Downloads and installs** if binaries are available
3. **Automatically builds from source** if binaries aren't available (requires Go and Git)
4. **Adds to your PATH** automatically
5. **Ready to use** - just restart your terminal or run `flux-relay --version`

## 🔧 Manual Installation

### Build from Source

```bash
git clone https://github.com/postacksol/flux-relay-cli.git
cd flux-relay-cli
go build -o flux-relay .
```

Then add to your PATH or use directly.

### Download Pre-built Binary

1. Go to [Releases](https://github.com/postacksol/flux-relay-cli/releases)
2. Download the binary for your OS/architecture
3. Rename to `flux-relay` (or `flux-relay.exe` on Windows)
4. Add to your PATH or use directly

## ✅ Verify Installation

After installation, verify it works:

```bash
flux-relay --version
```

You should see: `flux-relay version 1.0.0`

## 🆘 Troubleshooting

### "command not found" after installation

- **Windows:** Restart your PowerShell/terminal, or run:
  ```powershell
  $env:Path += ";$env:USERPROFILE\.flux-relay\bin"
  ```

- **Linux/macOS:** Restart your terminal, or run:
  ```bash
  export PATH="$PATH:$HOME/.flux-relay/bin"
  source ~/.bashrc  # or ~/.zshrc
  ```

### Go install doesn't work

Make sure:
- Go is installed: `go version`
- `$GOPATH/bin` or `$HOME/go/bin` is in your PATH
- You have internet connectivity

### Installer fails to download

- Check your internet connection
- Try downloading the installer script manually
- Use `go install` method instead

## 📚 Next Steps

After installation, see the [README.md](README.md) for usage instructions.
