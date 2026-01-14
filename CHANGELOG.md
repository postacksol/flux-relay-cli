# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-14

### Added
- Initial release of Flux Relay CLI
- OAuth 2.0 Device Authorization Grant authentication flow
- `login` command with automatic browser opening
- `login --headless` flag for headless/WSL environments
- `logout` command to remove stored tokens
- `config set token` command for manual token configuration
- Automatic detection of already logged-in users
- Token validation before saving
- Secure token storage in `~/.flux-relay/config.json`
- Dad jokes display during authentication wait (optional)
- ASCII logo display
- Support for custom API URLs via `--api-url` flag

### Security
- Tokens stored with secure file permissions (0600)
- Token validation before acceptance
- Secure HTTP client with proper headers
