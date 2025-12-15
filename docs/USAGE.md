# CFLIP Usage Guide

## Quick Start

### Installation
```bash
# Install from source
go install github.com/vanducng/cflip@latest

# Or build from source
git clone https://github.com/vanducng/cflip.git
cd cflip
make install
```

### Basic Usage

#### Switch Providers
```bash
# Interactive mode - prompts for selection
cflip switch

# Direct switch to Anthropic (official provider)
cflip switch anthropic

# Switch to GLM (z.ai provider)
cflip switch glm

# Switch to a custom provider
cflip switch my-provider
```

#### Check Current Status
```bash
# Check configuration location (current version only supports switch command)
cflip switch --help
```

## Command Reference

### switch
Switch between Claude providers.

```bash
cflip switch [provider] [flags]
```

**Description:**
Switch the active Claude provider. This will update your `~/.cflip/config.toml` file and generate the appropriate Claude settings for the specified provider.

**Available providers:**
- `anthropic` - Official Anthropic Claude API (optional API key, uses default endpoint)
- `glm` - GLM models by z.ai (requires API key and base URL)
- `custom` - Any custom provider (requires API key and base URL)

**Options:**
- `--verbose, -v`: Show detailed output
- `--quiet, -q`: Suppress output except errors
- `--help, -h`: Show help for the command

**Examples:**
```bash
# Interactive selection
cflip switch

# Direct switch to Anthropic
cflip switch anthropic

# Switch to GLM with verbose output
cflip switch glm --verbose

# Quiet mode
cflip switch anthropic --quiet
```

## Provider Configuration

### Anthropic (Official)
The Anthropic provider is special - it's the default/owner provider:

```bash
$ cflip switch anthropic
Configure API key for Anthropic? (optional, Y/n): Y
Enter Anthropic API key (optional): [hidden]

✓ Successfully switched to anthropic

Configuration: Using Anthropic with default endpoint
Authentication: API Key configured
```

**Features:**
- Uses Claude Code's default endpoint (no `ANTHROPIC_BASE_URL` set)
- Optional API key (can use Claude Code subscription)
- No model mappings needed (uses defaults)

### External Providers (GLM, Custom)

#### First-time Setup
```bash
$ cflip switch glm

Configuring glm provider
? Enter glm API token: [hidden]
? Enter glm base URL: https://api.z.ai/api/anthropic

Configure model mappings? (Y/n): Y
? Enter model for haiku category (optional): glm-4.6-air
? Enter model for sonnet category (optional): glm-4.6
? Enter model for opus category (optional): [leave empty]

✓ Successfully switched to glm

Configuration:
  Base URL: https://api.z.ai/api/anthropic
  Model Mappings:
    haiku: glm-4.6-air
    sonnet: glm-4.6

Authentication: API Key
```

**Features:**
- Required: API token and base URL
- Optional: Model mappings for haiku/sonnet/opus categories
- Updates all 5 environment variables in settings.json

## Configuration Files

### CFLIP Configuration (`~/.cflip/config.toml`)
```toml
provider = "glm"

[providers]
[providers.anthropic]
token = "sk-ant-api..."  # Optional API key

[providers.glm]
token = "your-glm-token"
base_url = "https://api.z.ai/api/anthropic"

[providers.glm.model_map]
haiku = "glm-4.6-air"
sonnet = "glm-4.6"
```

### Claude Settings (`~/.claude/settings.json`)
CFLIP automatically manages this file:

**For Anthropic:**
```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-ant-api..."
  }
}
```

**For External Providers:**
```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your-token",
    "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.6-air",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.6"
  }
}
```

## Provider Setup

### Anthropic
1. Optional: Get API key from [Anthropic Console](https://console.anthropic.com/)
2. Run: `cflip switch anthropic`
3. Optionally enter API key (can use Claude Code subscription)

### GLM by z.ai
1. Visit [Z.AI Platform](https://platform.z.ai) and register
2. Subscribe to GLM Coding Plan
3. Generate API key from dashboard
4. Run: `cflip switch glm`
5. Enter API key: `[your-token]`
6. Enter base URL: `https://api.z.ai/api/anthropic`

### Custom Provider
Any Anthropic-compatible API:
```bash
cflip switch my-custom-provider
# Follow prompts to configure:
# - API token
# - Base URL
# - Optional model mappings
```

## Model Categories

External providers can map their models to Anthropic's categories:

- **haiku**: Fast, efficient models for quick tasks
- **sonnet**: Advanced models for most tasks
- **opus**: Most capable models for complex tasks

## Examples

### Basic Provider Switching
```bash
# Start with Anthropic
cflip switch anthropic

# Switch to GLM for testing
cflip switch glm

# Switch back
cflip switch anthropic
```

### Custom Provider Setup
```bash
# Add a custom OpenAI-compatible provider
cflip switch openai-compatible
# ? Enter openai-compatible API token: sk-...
# ? Enter openai-compatible base URL: https://api.openai.com/v1
# ? Configure model mappings? Y
# ? Enter model for haiku category: gpt-4o-mini
# ? Enter model for sonnet category: gpt-4o
```

### Verbose Output
```bash
$ cflip switch glm --verbose

✓ Successfully switched to glm

Configuration:
  Base URL: https://api.z.ai/api/anthropic
  Model Mappings:
    haiku: glm-4.6-air
    sonnet: glm-4.6

Authentication: API Key

Configuration saved to: /Users/user/.cflip/config.toml
Claude settings updated at: /Users/user/.claude/settings.json
```

## Integration with Claude Code

After switching providers:
1. Close all Claude Code windows
2. Start Claude Code in a new terminal
3. Claude Code will automatically use the new provider configuration

## Troubleshooting

### API Key Issues
- Verify API key is correct for the provider
- Check subscription status
- Ensure token has necessary permissions

### Configuration Not Applied
- Check `~/.claude/settings.json` exists
- Restart Claude Code after switching
- Use `--verbose` flag for debugging

### Custom Provider Not Working
- Verify base URL is correct
- Check API compatibility with Anthropic format
- Test with curl or similar tool first

## Tips

- Use `--verbose` flag to see what files are being updated
- Store sensitive tokens only in the config file (not in shell history)
- Test with Anthropic first to ensure CFLIP is working
- Always verify custom provider API compatibility
