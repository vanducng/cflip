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

#### List Available Providers
```bash
cflip list
```

Output:
```
PROVIDER     NAME         DESCRIPTION                                         MODELS
  anthropic  Anthropic    Official Anthropic Claude API provider              claude-3-5-haiku-20241022/claude-3-5-sonnet-20241022/claude-3-opus-20240229
* glm        GLM by z.ai  GLM models from z.ai with Anthropic-compatible API  glm-4.5-air/glm-4.6/glm-4.6
```

#### Check Current Status
```bash
cflip status
```

Output:
```
Current provider: GLM by z.ai
CONFIGURATION  VALUE
Base URL       https://api.z.ai/api/anthropic
Haiku Model    glm-4.5-air
Sonnet Model   glm-4.6
Opus Model     glm-4.6
API Timeout    3000000 ms

API Key: Configured ✓
```

#### Switch Providers
```bash
# Interactive mode - prompts for selection
cflip switch

# Direct switch
cflip switch anthropic
cflip switch glm
```

## Command Reference

### switch
Switch between Claude providers.

```bash
cflip switch [provider]
```

**Options:**
- `--verbose, -v`: Show detailed output
- `--quiet, -q`: Suppress output except errors

**Examples:**
```bash
# Interactive selection
cflip switch

# Direct switch to Anthropic
cflip switch anthropic

# Verbose output
cflip switch glm --verbose
```

### list
List all available providers.

```bash
cflip list
```

**Output columns:**
- `*`: Indicates currently active provider
- `PROVIDER`: Provider identifier
- `NAME`: Display name
- `DESCRIPTION`: Brief description
- `MODELS`: Available models (haiku/sonnet/opus)

### status
Show current provider status and configuration.

```bash
cflip status
```

**Output includes:**
- Current provider name
- Base URL and models
- API key status
- Settings file location
- Available features (for GLM)

### backup
Manage configuration backups.

```bash
cflip backup [subcommand]
```

**Subcommands:**

#### backup create
Create a backup of current settings.
```bash
cflip backup create [--description "text"]
```

**Example:**
```bash
# Simple backup
cflip backup create

# Backup with description
cflip backup create -d "Before switching to GLM"
```

#### backup list
List all available backups.
```bash
cflip backup list
```

**Output:**
```
ID                      TIMESTAMP           PROVIDER  SIZE
backup-20250114-200000  2025-01-14 20:00:00  anthropic  245 bytes
backup-20250114-201500  2025-01-14 20:15:00  glm       238 bytes
```

#### backup restore
Restore settings from a backup.
```bash
cflip backup restore <backup-id>
```

**Example:**
```bash
cflip backup restore backup-20250114-200000
```

#### backup delete
Delete a specific backup.
```bash
cflip backup delete <backup-id>
```

#### backup prune
Delete old backups.
```bash
cflip backup prune [--older-than duration]
```

**Duration formats:**
- `7d` - 7 days (default)
- `24h` - 24 hours
- `30m` - 30 minutes

## Global Options

All commands support these global options:

- `--verbose, -v`: Show detailed output
- `--quiet, -q`: Suppress output except errors
- `--help, -h`: Show help for the command

## Provider Setup

### Anthropic
1. Get API key from [Anthropic Console](https://console.anthropic.com/)
2. Ensure you have credits or active subscription
3. Run: `cflip switch anthropic`
4. Enter API key when prompted

### GLM by z.ai
1. Visit [Z.AI Platform](https://platform.z.ai) and register
2. Subscribe to GLM Coding Plan
3. Generate API key from dashboard
4. Run: `cflip switch glm`
5. Enter API key when prompted

## Configuration File

CFLIP manages the `~/.claude/settings.json` file:

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your-api-key",
    "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
    "API_TIMEOUT_MS": "3000000",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.5-air",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.6",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-4.6"
  }
}
```

## Troubleshooting

### API Key Not Accepted
- Ensure API key is correct for the provider
- Check subscription status (especially for GLM)
- Verify key permissions

### Can't Determine Current Provider
- Check if `~/.claude/settings.json` exists
- Verify file has correct JSON format
- Run `cflip status --verbose` for details

### Backup Failed
- Check permissions on `~/.claude/` directory
- Ensure disk space is available
- Run with `--verbose` for error details

### Permission Denied
- Ensure write permissions on `~/.claude/settings.json`
- Check if Claude Code is running (may lock the file)
- Try restarting with a new terminal

## Examples

### Switching Workflow
```bash
# Check current setup
cflip status

# Create backup before switching
cflip backup create -d "Before GLM switch"

# Switch to GLM
cflip switch glm

# Verify switch
cflip status

# List backups
cflip backup list

# Switch back if needed
cflip switch anthropic

# Restore from backup
cflip backup restore backup-20250114-200000
```

### Automated Script
```bash
#!/bin/bash
# Switch provider and verify

PROVIDER=$1
if [ -z "$PROVIDER" ]; then
    echo "Usage: $0 <provider>"
    exit 1
fi

# Create backup with timestamp
cflip backup create -d "Auto-backup before $PROVIDER"

# Switch provider
cflip switch $PROVIDER

# Verify
cflip status --quiet
```

## Integration with Claude Code

After switching providers:
1. Close all Claude Code windows
2. Start Claude Code in a new terminal
3. Verify with `/status` command in Claude Code

## Tips

- Always create backups before switching
- Use `cflip status --verbose` for debugging
- Regular backup pruning: `cflip backup prune --older-than 30d`
- Check provider setup requirements before switching