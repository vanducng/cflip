#!/bin/bash

# CFLIP Installation Script
# Installs cflip (Claude Provider Switcher)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="$HOME/.local/bin"
PLATFORM=""
ARCH=""
VERSION="latest"

# Help function
show_help() {
    cat << EOF
CFLIP Installation Script

USAGE:
    install.sh [OPTIONS]

OPTIONS:
    -d, --dir DIR        Installation directory (default: $HOME/.local/bin)
    -v, --version VER    Version to install (default: latest)
    -p, --platform PLAT  Override platform detection
    -a, --arch ARCH      Override architecture detection
    -h, --help           Show this help message

EXAMPLES:
    ./install.sh                    # Install latest to ~/.local/bin
    ./install.sh -d /usr/local/bin  # Install to /usr/local/bin
    ./install.sh -v v1.0.0         # Install specific version
    sudo ./install.sh -d /usr/local/bin  # Install system-wide

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

# Detect platform
detect_platform() {
    if [[ "$PLATFORM" != "" ]]; then
        return
    fi

    case "$(uname -s)" in
        Darwin*)
            PLATFORM="darwin"
            ;;
        Linux*)
            PLATFORM="linux"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            PLATFORM="windows"
            ;;
        *)
            print_error "Unsupported platform: $(uname -s)"
            exit 1
            ;;
    esac
}

# Detect architecture
detect_arch() {
    if [[ "$ARCH" != "" ]]; then
        return
    fi

    case "$(uname -m)" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

# Check if directory is in PATH
check_path() {
    if [[ ":$PATH:" == *":$1:"* ]]; then
        return 0
    else
        return 1
    fi
}

# Install from source
install_from_source() {
    print_status "Installing from source..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        echo "Visit: https://golang.org/dl/"
        exit 1
    fi

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # Clone repository
    print_status "Cloning repository..."
    if [[ "$VERSION" == "latest" ]]; then
        git clone https://github.com/vanducng/cflip.git
    else
        git clone --branch "$VERSION" https://github.com/vanducng/cflip.git
    fi

    cd cflip

    # Build
    print_status "Building cflip..."
    make build

    # Install
    mkdir -p "$INSTALL_DIR"
    cp bin/cflip "$INSTALL_DIR/"

    # Cleanup
    cd /
    rm -rf "$TMP_DIR"

    print_status "Installation complete!"
}

# Install from release
install_from_release() {
    print_status "Installing from release..."

    # Determine file name
    if [[ "$PLATFORM" == "windows" ]]; then
        FILENAME="cflip-${ARCH}.exe"
    else
        FILENAME="cflip-${ARCH}"
    fi

    # Get release download URL
    if [[ "$VERSION" == "latest" ]]; then
        DOWNLOAD_URL="https://github.com/vanducng/cflip/releases/latest/download/${FILENAME}"
    else
        DOWNLOAD_URL="https://github.com/vanducng/cflip/releases/download/${VERSION}/${FILENAME}"
    fi

    # Create installation directory
    mkdir -p "$INSTALL_DIR"

    # Download
    print_status "Downloading cflip..."
    if command -v curl &> /dev/null; then
        curl -L "$DOWNLOAD_URL" -o "$INSTALL_DIR/cflip${PLATFORM:+.exe}"
    elif command -v wget &> /dev/null; then
        wget "$DOWNLOAD_URL" -O "$INSTALL_DIR/cflip${PLATFORM:+.exe}"
    else
        print_error "Neither curl nor wget is available"
        exit 1
    fi

    # Make executable (not for Windows)
    if [[ "$PLATFORM" != "windows" ]]; then
        chmod +x "$INSTALL_DIR/cflip"
    fi

    print_status "Installation complete!"
}

# Main installation
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -p|--platform)
                PLATFORM="$2"
                shift 2
                ;;
            -a|--arch)
                ARCH="$2"
                shift 2
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    print_status "Installing CFLIP (Claude Provider Switcher)"
    print_status "Version: $VERSION"
    print_status "Install directory: $INSTALL_DIR"

    # Detect platform and architecture
    detect_platform
    detect_arch

    print_status "Platform: $PLATFORM-$ARCH"

    # Check if install directory exists and is writable
    if [[ ! -d "$INSTALL_DIR" ]]; then
        print_status "Creating installation directory..."
        mkdir -p "$INSTALL_DIR" || {
            print_error "Cannot create installation directory: $INSTALL_DIR"
            exit 1
        }
    fi

    # Check if we can write to install directory
    if [[ ! -w "$INSTALL_DIR" ]]; then
        print_error "Cannot write to installation directory: $INSTALL_DIR"
        print_status "Try running with sudo or choose a different directory"
        exit 1
    fi

    # Try to install from release first, fallback to source
    if install_from_release; then
        print_status "Installed from pre-built release"
    else
        print_warning "Pre-built release not available, installing from source"
        install_from_source
    fi

    # Check if installation directory is in PATH
    if ! check_path "$INSTALL_DIR"; then
        print_warning "Installation directory is not in your PATH"
        print_status "Add the following to your shell profile:"
        echo ""
        case "$(basename "$SHELL")" in
            bash)
                echo "export PATH=\"\$PATH:$INSTALL_DIR\""
                echo "# Add to ~/.bashrc"
                ;;
            zsh)
                echo "export PATH=\"\$PATH:$INSTALL_DIR\""
                echo "# Add to ~/.zshrc"
                ;;
            fish)
                echo "set -gx PATH \$PATH $INSTALL_DIR"
                echo "# Add to ~/.config/fish/config.fish"
                ;;
            *)
                echo "export PATH=\"\$PATH:$INSTALL_DIR\""
                echo "# Add to your shell profile"
                ;;
        esac
        echo ""
    fi

    # Verify installation
    if command -v "$INSTALL_DIR/cflip" &> /dev/null; then
        print_status "CFLIP installed successfully!"
        print_status "Run 'cflip --help' to get started"

        # Show version
        if [[ "$PLATFORM" != "windows" ]]; then
            "$INSTALL_DIR/cflip" --version 2>/dev/null || true
        fi
    else
        print_error "Installation verification failed"
        exit 1
    fi
}

# Run main function
main "$@"