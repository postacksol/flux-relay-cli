# Windows Installation

## Quick Install (One Command)

```powershell
Invoke-WebRequest -Uri https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 -OutFile install.ps1; .\install.ps1
```

## Manual Install

1. **Download the installer:**
   ```powershell
   Invoke-WebRequest -Uri https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 -OutFile install.ps1
   ```

2. **Run the installer:**
   ```powershell
   .\install.ps1
   ```

## Alternative: Download Binary Directly

1. Go to: https://github.com/postacksol/flux-relay-cli/releases
2. Download `flux-relay-windows-amd64.exe` (or appropriate for your architecture)
3. Rename to `flux-relay.exe`
4. Add to your PATH or use directly

## Alternative: Go Install

```powershell
go install github.com/postacksol/flux-relay-cli@v1.0.0
```

## Alternative: Build from Source

```powershell
git clone https://github.com/postacksol/flux-relay-cli.git
cd flux-relay-cli
go build -o flux-relay.exe .
```
