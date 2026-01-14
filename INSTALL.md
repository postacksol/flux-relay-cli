# Installation Instructions

## Option 1: Build from Source (Recommended)

Since the repository is public, you can clone and build directly:

```bash
git clone https://github.com/postacksol/flux-relay-cli.git
cd flux-relay-cli
go build -o flux-relay.exe .
```

Then add to your PATH or use directly:
```bash
.\flux-relay.exe login
```

## Option 2: Install via Go (if module resolution works)

```bash
go install github.com/postacksol/flux-relay-cli@v1.0.0
```

**Note:** If you encounter module resolution errors, use Option 1 instead.

## Option 3: Download Pre-built Binary

Once GitHub Releases are set up with binaries, you can download directly from:
https://github.com/postacksol/flux-relay-cli/releases
