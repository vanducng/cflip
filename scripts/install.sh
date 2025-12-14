#!/bin/bash

# CFLIP Install Script
# Installs the Claude Provider Switcher (cflip) CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
REPO="vanducng/cflip"
VERSION="latest"
INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="cflip"

# Help message
help() {
    cat << EOF
CFLIP Install Script

USAGE:
    curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash

OPTIONS:
    VERSION=<version>    Install specific version (default: latest)
    INSTALL_DIR=<dir>    Installation directory (default: $HOME/.local/bin)
    BINARY_NAME=<name>   Binary name (default: cflip)

ENVIRONMENT VARIABLES:
    CFLIP_VERSION        Version to install (default: latest)
    CFLIP_INSTALL_DIR    Installation directory (default: $HOME/.local/bin)
    CFLIP_BINARY_NAME    Binary name (default: cflip)

EXAMPLES:
    # Install latest version
    curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash

    # Install specific version
    curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash -s -- --version=v1.1.2

    # Install to custom directory
    curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash -s -- --install-dir=/usr/local/bin

    # Install with custom binary name
    curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash -s -- --binary-name=claude-flip

    # Using environment variables
    CFLIP_VERSION=v1.1.2 curl -sSL https://raw.githubusercontent.com/vanducng/cflip/main/scripts/install.sh | bash

EOF
}

# Print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version=*)
                VERSION="${1#*=}"
                shift
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --install-dir=*)
                INSTALL_DIR="${1#*=}"
                shift
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --binary-name=*)
                BINARY_NAME="${1#*=}"
                shift
                ;;
            --binary-name)
                BINARY_NAME="$2"
                shift 2
                ;;
            -h|--help)
                help
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                help
                exit 1
                ;;
        esac
    done

    # Override with environment variables if set
    if [[ -n "$CFLIP_VERSION" ]]; then
        VERSION="$CFLIP_VERSION"
    fi
    if [[ -n "$CFLIP_INSTALL_DIR" ]]; then
        INSTALL_DIR="$CFLIP_INSTALL_DIR"
    fi
    if [[ -n "$CFLIP_BINARY_NAME" ]]; then
        BINARY_NAME="$CFLIP_BINARY_NAME"
    fi
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $OS in
        darwin)
            OS="Darwin"
            ;;
        linux)
            OS="Linux"
            ;;
        mingw*|msys*|cygwin*)
            OS="Windows"
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            ARCH="x86_64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
}

# Get latest version from GitHub API
get_latest_version() {
    print_status "Fetching latest version..."
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4)
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4)
    else
        echo -e "${RED}Neither curl nor wget is available${NC}"
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to get latest version${NC}"
        exit 1
    fi
}

# Download function
download() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$output" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$output" "$url"
    else
        echo -e "${RED}Neither curl nor wget is available${NC}"
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksum_url="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$checksum_url" > "${file}.sha256"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "${file}.sha256" "$checksum_url"
    fi

    if [ -f "${file}.sha256" ]; then
        if command -v shasum >/dev/null 2>&1; then
            EXPECTED=$(grep "$file" "${file}.sha256" | awk '{print $1}')
            ACTUAL=$(shasum -a 256 "$file" | awk '{print $1}')
            if [ "$EXPECTED" = "$ACTUAL" ]; then
                print_status "Checksum verified"
            else
                echo -e "${RED}Checksum verification failed${NC}"
                echo "Expected: $EXPECTED"
                echo "Actual: $ACTUAL"
                exit 1
            fi
        else
            print_warning "shasum not available, skipping checksum verification"
        fi
        rm -f "${file}.sha256"
    fi
}

# Install binary
install_binary() {
    local os="$1"
    local arch="$2"
    local version="$3"

    # Determine archive format
    if [ "$os" = "Windows" ]; then
        ARCHIVE_EXT="zip"
    else
        ARCHIVE_EXT="tar.gz"
    fi

    # Construct download URL
    local archive_name="${BINARY_NAME}_${os}_${arch}.${ARCHIVE_EXT}"
    local download_url="https://github.com/$REPO/releases/download/$version/$archive_name"
    local checksum_url="https://github.com/$REPO/releases/download/$version/checksums.txt"

    print_status "Downloading: $archive_name"

    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download archive
    local archive_path="$tmp_dir/$archive_name"
    if ! download "$download_url" "$archive_path"; then
        echo -e "${RED}Failed to download: $download_url${NC}"
        exit 1
    fi

    # Download and verify checksum
    verify_checksum "$archive_path" "$checksum_url"

    # Extract archive
    print_status "Extracting archive..."
    cd "$tmp_dir"

    if [ "$ARCHIVE_EXT" = "zip" ]; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$archive_path"
        else
            echo -e "${RED}unzip is required to extract ZIP files${NC}"
            exit 1
        fi
    else
        if command -v tar >/dev/null 2>&1; then
            tar -xzf "$archive_path"
        else
            echo -e "${RED}tar is required to extract tar.gz files${NC}"
            exit 1
        fi
    fi

    # Find the binary
    local binary_path=$(find . -type f -name "$BINARY_NAME*" -executable | head -n 1)
    if [ -z "$binary_path" ]; then
        # Fallback: look for any file with the binary name
        binary_path=$(find . -type f -name "$BINARY_NAME*" | head -n 1)
    fi

    if [ -z "$binary_path" ]; then
        echo -e "${RED}Binary not found in archive${NC}"
        echo "Contents:"
        ls -la
        exit 1
    fi

    # Make binary executable (for Unix systems)
    if [ "$os" != "Windows" ]; then
        chmod +x "$binary_path"
    fi

    # Create install directory
    mkdir -p "$INSTALL_DIR"

    # Move binary to install directory
    local final_path="$INSTALL_DIR/$BINARY_NAME"
    if [ "$os" = "Windows" ]; then
        final_path="$INSTALL_DIR/$BINARY_NAME.exe"
        # If the binary already has .exe extension, keep it
        if [[ "$binary_path" == *.exe ]]; then
            mv "$binary_path" "$final_path"
        else
            mv "$binary_path" "$final_path"
        fi
    else
        mv "$binary_path" "$final_path"
    fi

    echo -e "${GREEN}Installed to: $final_path${NC}"
}

# Add to PATH if needed
add_to_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
        echo -e "${YELLOW}Add the following to your shell profile:${NC}"
        echo ""
        if [ -n "$ZSH_VERSION" ] || [ -n "$BASH_VERSION" ]; then
            echo "export PATH=\"\$PATH:$INSTALL_DIR\""
            if [ -f "$HOME/.zshrc" ]; then
                echo "# Then run: source ~/.zshrc"
            elif [ -f "$HOME/.bashrc" ]; then
                echo "# Then run: source ~/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                echo "# Then run: source ~/.bash_profile"
            fi
        else
            echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        fi
        echo ""
    fi
}

# Main installation
main() {
    echo -e "${GREEN}CFLIP Installer${NC}"
    echo "=================="
    echo ""

    # Parse arguments
    parse_args "$@"

    # Detect OS and architecture
    detect_os
    detect_arch

    echo "OS: $OS"
    echo "Architecture: $ARCH"
    echo "Install Directory: $INSTALL_DIR"
    echo "Binary Name: $BINARY_NAME"
    echo ""

    # Get version if not specified
    if [ "$VERSION" = "latest" ]; then
        get_latest_version
    else
        # Ensure version starts with v
        if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            VERSION="v$VERSION"
        fi
    fi

    echo -e "${GREEN}Installing CFLIP $VERSION${NC}"
    echo ""

    # Check if install directory exists and is writable
    if [ ! -d "$INSTALL_DIR" ]; then
        print_status "Creating installation directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR" || {
            echo -e "${RED}Cannot create installation directory: $INSTALL_DIR${NC}"
            exit 1
        }
    fi

    # Check if we can write to install directory
    if [ ! -w "$INSTALL_DIR" ]; then
        echo -e "${RED}Cannot write to installation directory: $INSTALL_DIR${NC}"
        echo "Try running with sudo or choose a different directory"
        exit 1
    fi

    # Install binary
    install_binary "$OS" "$ARCH" "$VERSION"

    # Add to PATH warning if needed
    add_to_path

    echo ""
    echo -e "${GREEN}Installation complete!${NC}"
    echo ""
    echo -e "${YELLOW}To get started:${NC}"
    echo "  $BINARY_NAME --help"
    echo ""

    # Show version if available
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ] || [ -f "$INSTALL_DIR/$BINARY_NAME.exe" ]; then
        echo -e "${YELLOW}Version:${NC}"
        if [ "$OS" = "Windows" ]; then
            "$INSTALL_DIR/$BINARY_NAME.exe" --version 2>/dev/null || echo "Version not available"
        else
            "$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null || echo "Version not available"
        fi
    fi
}

# Run main function with all arguments
main "$@"