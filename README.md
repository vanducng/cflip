# CFLIP - Claude Provider Switcher

[![Go Report Card](https://goreportcard.com/badge/github.com/vanducng/cflip)](https://goreportcard.com/report/github.com/vanducng/cflip)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest Release](https://img.shields.io/github/release/vanducng/cflip)](https://github.com/vanducng/cflip/releases)

Quickly switch between Claude Code providers (Anthropic, GLM/z.ai). Manages your `~/.claude/settings.json` automatically.

## Install

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash
```

### Install Specific Version
```bash
curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash -s -- --version=v1.1.3
```

### Other Options
```bash
# macOS with Homebrew
brew install vanducng/tap/cflip

# Go install
go install github.com/vanducng/cflip@latest

# Download binaries
# https://github.com/vanducng/cflip/releases
```

## Usage

```bash
# List providers
cflip list

# Switch provider
cflip switch glm
cflip switch anthropic

# Check current provider
cflip status

# Get help
cflip --help
```

## Supported Providers

- **Anthropic** - Official Claude API
- **GLM (z.ai)** - GLM models with Anthropic-compatible API
  - Requires GLM subscription from [Z.AI Platform](https://platform.z.ai)

## Architecture Support

- ✅ macOS (Intel & Apple Silicon)
- ✅ Linux (x86_64 & ARM64)
- ✅ Windows (x86_64)

## License

MIT