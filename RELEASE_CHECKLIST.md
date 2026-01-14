# Release Checklist for v1.0.0

## Pre-Release

- [x] Update version in `cmd/root.go` to `1.0.0`
- [x] Update `go.mod` module path to `github.com/fluxrelay/flux-relay-cli`
- [x] Update README.md with v1.0.0 information
- [x] Create CHANGELOG.md
- [x] Run `go mod tidy` to ensure dependencies are clean
- [x] Build and test the binary: `go build -o flux-relay.exe .`

## Git Release (if using Git)

```bash
# Initialize git repo (if not already)
cd flux-relay-cli
git init
git add .
git commit -m "Release v1.0.0 - Initial CLI with authentication"

# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0 - Initial CLI with authentication"
git push origin v1.0.0
```

## GitHub Release

1. Go to GitHub repository
2. Click "Releases" → "Create a new release"
3. Tag: `v1.0.0`
4. Title: `v1.0.0 - Initial Release`
5. Description: Copy from `RELEASE.md`
6. Upload binaries for:
   - Windows: `flux-relay.exe`
   - Linux: `flux-relay` (build with `GOOS=linux go build`)
   - macOS: `flux-relay` (build with `GOOS=darwin go build`)

## Binary Builds

### Windows
```bash
go build -o flux-relay.exe .
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o flux-relay-linux-amd64 .
```

### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o flux-relay-darwin-amd64 .
```

### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o flux-relay-darwin-arm64 .
```

## Post-Release

- [ ] Verify installation works: `go install github.com/fluxrelay/flux-relay-cli@v1.0.0`
- [ ] Test authentication flow
- [ ] Test headless mode
- [ ] Update documentation if needed
