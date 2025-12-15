# CFLIP - Claude Provider Switcher

[![CI](https://github.com/vanducng/cflip/workflows/CI/badge.svg)](https://github.com/vanducng/cflip/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vanducng/cflip)](https://goreportcard.com/report/github.com/vanducng/cflip)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest Release](https://img.shields.io/github/release/vanducng/cflip)](https://github.com/vanducng/cflip/releases)

A simple CLI tool to switch between Claude Code providers. Automatically manages your `~/.claude/settings.json` configuration.

**✨ Features:**
- 🔄 Instantly switch between Claude providers
- 🎯 Special handling for Anthropic (uses default Claude Code endpoint)
- 🔧 Support for external providers with custom endpoints
- 📝 Optional model mapping configuration
- 🚀 Minimal configuration required

## Install

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash
```

### Install Specific Version
```bash
curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash -s -- --version=v1.6.0
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

### Quick Start

```bash
# Interactive provider selection
cflip switch

# Switch directly to a provider
cflip switch anthropic

# Configure a new external provider
cflip switch glm

# Get help
cflip switch --help
```

### Provider Configuration

#### Anthropic (Official)
```bash
# Switch to Anthropic with default settings
cflip switch anthropic

# Optionally configure an API key
cflip switch anthropic
# ? Configure API key for Anthropic? (optional, Y/n): Y
# ? Enter Anthropic API key (optional): [your-api-key]
```

#### External Providers (e.g., GLM, Custom)
```bash
# Switch to an external provider
cflip switch glm

# You'll be prompted for:
# ? Enter glm API token: [your-token]
# ? Enter glm base URL: https://api.z.ai/api/anthropic

# Optionally configure model mappings:
# ? Configure model mappings? (Y/n): Y
# ? Enter model for haiku category (optional): glm-4.6-air
# ? Enter model for sonnet category (optional): glm-4.6
# ? Enter model for opus category (optional): [enter]
```

### Configuration File

Your configuration is stored in `~/.cflip/config.toml`:

```toml
provider = "glm"

[providers]
[providers.anthropic]
# token = ""  # Optional API key

[providers.glm]
token = "your-api-token"
base_url = "https://api.z.ai/api/anthropic"

[providers.glm.model_map]
haiku = "glm-4.6-air"
sonnet = "glm-4.6"
```

### What CFLIP Updates

When you switch providers, CFLIP updates your `~/.claude/settings.json`:

**For Anthropic:**
- Only sets `ANTHROPIC_AUTH_TOKEN` (if provided)
- Uses Claude Code's default endpoint (no `ANTHROPIC_BASE_URL`)
- No model mappings (uses defaults)

**For External Providers:**
- Sets `ANTHROPIC_AUTH_TOKEN`
- Sets `ANTHROPIC_BASE_URL`
- Optionally sets model mappings (`ANTHROPIC_DEFAULT_HAIKU_MODEL`, etc.)

## Supported Providers

### Built-in Support
- **anthropic** - Official Anthropic Claude API (uses Claude Code default endpoint)
- **glm** - GLM models by z.ai

### Custom Providers
You can add any Anthropic-compatible provider:
```bash
cflip switch my-provider
# ? Enter my-provider API token: [token]
# ? Enter my-provider base URL: [url]
```

### Provider Details

#### GLM (z.ai)
- **Provider name:** `glm`
- **Base URL:** `https://api.z.ai/api/anthropic`
- **Requirements:** GLM subscription from [Z.AI Platform](https://platform.z.ai)
- **Example models:** `glm-4.6-air` (haiku), `glm-4.6` (sonnet)

## Advanced Usage

### Verbose Output
```bash
# See detailed information about the switch process
cflip switch glm --verbose

# Quiet mode - minimal output
cflip switch anthropic --quiet
```

### Example: Setting up GLM Provider
```bash
# First time setup for GLM
$ cflip switch glm

Configuring glm provider
? Enter glm API token: [hidden]
? Enter glm base URL: https://api.z.ai/api/anthropic

Configure model mappings? (Y/n): Y
? Enter model for haiku category (optional): glm-4.6-air
? Enter model for sonnet category (optional): glm-4.6
? Enter model for opus category (optional):

✓ Successfully switched to glm

Configuration:
  Base URL: https://api.z.ai/api/anthropic
  Model Mappings:
    haiku: glm-4.6-air
    sonnet: glm-4.6

Authentication: API Key

Configuration saved to: /Users/username/.cflip/config.toml
Claude settings updated at: /Users/username/.claude/settings.json
```

### Manual Configuration
You can also manually edit `~/.cflip/config.toml`:

```toml
# Active provider
provider = "anthropic"

# Provider configurations
[providers]
[providers.anthropic]
# token = "sk-ant-api..."  # Optional

[providers.glm]
token = "your-glm-token"
base_url = "https://api.z.ai/api/anthropic"

[providers.custom]
token = "your-token"
base_url = "https://your-api-endpoint.com"
[providers.custom.model_map]
haiku = "your-haiku-model"
sonnet = "your-sonnet-model"
opus = "your-opus-model"
```

## How It Works

1. **Configuration Storage**: CFLIP stores provider configurations in `~/.cflip/config.toml`
2. **Settings Update**: When switching providers, CFLIP updates `~/.claude/settings.json` with the appropriate environment variables
3. **Model Categories**: External providers can map their models to Anthropic's categories (haiku, sonnet, opus)

## Architecture Support

- ✅ macOS (Intel & Apple Silicon)
- ✅ Linux (x86_64 & ARM64)
- ✅ Windows (x86_64)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
