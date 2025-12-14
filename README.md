# CFLIP - Claude Provider Switcher

[![Go Report Card](https://goreportcard.com/badge/github.com/vanducng/cflip)](https://goreportcard.com/report/github.com/vanducng/cflip)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)

CFLIP is a Go CLI tool that enables seamless switching between different Claude Code providers (Anthropic, GLM/z.ai, and future providers). It manages the `~/.claude/settings.json` configuration file to toggle between different API endpoints and authentication methods.

## Why CFLIP?

- **Quick Switching**: Change providers in under 2 seconds
- **Safe Operations**: Automatic backups before any changes
- **Extensible**: Easy to add new providers
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### From Source
```bash
go install github.com/vanducng/cflip@latest
```

### From Release (Coming Soon)
Download the appropriate binary for your platform from the [releases page](https://github.com/vanducng/cflip/releases).

## Usage

### List Available Providers
```bash
cflip list
```

### Switch to a Provider
```bash
# Switch to GLM (z.ai)
cflip switch glm

# Switch to Anthropic
cflip switch anthropic
```

### Check Current Status
```bash
cflip status
```

### Backup Management
```bash
# Create a backup
cflip backup create

# List backups
cflip backup list

# Restore from backup
cflip backup restore <backup-id>
```

## Supported Providers

### Anthropic (Default)
- Models: Claude 3.5 Haiku, Claude 3.5 Sonnet, Claude 3 Opus
- Endpoint: `https://api.anthropic.com`

### GLM by z.ai
- Models: GLM-4.5-air, GLM-4.6
- Endpoint: `https://api.z.ai/api/anthropic`
- Requires: GLM Coding Plan subscription from [Z.AI Platform](https://platform.z.ai)

## Configuration

CFLIP automatically manages your `~/.claude/settings.json` file. The first time you switch to a provider, you'll be prompted for:

- API Key
- Optional: Custom models or endpoints

## Examples

```bash
# Initial setup for GLM
cflip switch glm
> Enter your GLM API Key: zai-xxxxxxxxxxxxxxxxxxx
> Provider switched to GLM successfully!

# Quick switch back to Anthropic
cflip switch anthropic
> Provider switched to Anthropic successfully!

# Check your current provider
cflip status
> Current provider: GLM (z.ai)
> Models: haiku=glm-4.5-air, sonnet=glm-4.6, opus=glm-4.6
```

## Development

### Prerequisites
- Go 1.21 or later

### Build from Source
```bash
# Clone the repository
git clone https://github.com/vanducng/cflip.git
cd cflip

# Install dependencies
make deps

# Build
make build

# Run tests
make test
```

### Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [claude-code-switch](https://github.com/foreveryh/claude-code-switch)
- Built with [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper)