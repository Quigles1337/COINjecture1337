#!/bin/bash

# COINjecture Enhanced macOS Installer v3.6.4
# This script completely eliminates macOS security warnings

set -e

echo "🚀 COINjecture Enhanced macOS Installer v3.6.4"
echo "=============================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_special() {
    echo -e "${PURPLE}[SPECIAL]${NC} $1"
}

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This installer is for macOS only."
    exit 1
fi

# Create installation directory
INSTALL_DIR="$HOME/Applications/COINjecture"
print_status "Installing COINjecture to: $INSTALL_DIR"

# Create directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Copy the executable
if [ -f "./COINjecture" ]; then
    print_status "Copying COINjecture executable..."
    cp "./COINjecture" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/COINjecture"
    print_success "Executable copied successfully"
else
    print_error "COINjecture executable not found in current directory"
    exit 1
fi

# COMPREHENSIVE SECURITY WARNING ELIMINATION
print_special "🔒 ELIMINATING macOS SECURITY WARNINGS..."

# Remove ALL quarantine and security attributes
print_status "Removing quarantine attributes..."
xattr -d com.apple.quarantine "$INSTALL_DIR/COINjecture" 2>/dev/null || true
xattr -d com.apple.provenance "$INSTALL_DIR/COINjecture" 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemWhereFroms "$INSTALL_DIR/COINjecture" 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemDownloadedDate "$INSTALL_DIR/COINjecture" 2>/dev/null || true

# Clear ALL extended attributes
print_status "Clearing all extended attributes..."
xattr -c "$INSTALL_DIR/COINjecture" 2>/dev/null || true

# Set proper permissions
print_status "Setting proper permissions..."
chmod 755 "$INSTALL_DIR/COINjecture"

# Create a wrapper script that bypasses security warnings
print_status "Creating security-bypass wrapper..."
cat > "$INSTALL_DIR/COINjecture-Safe" << 'EOF'
#!/bin/bash
# COINjecture Security-Bypass Wrapper
# This script eliminates macOS security warnings

cd "$HOME/Applications/COINjecture"

# Remove any quarantine attributes that might have been re-added
xattr -d com.apple.quarantine ./COINjecture 2>/dev/null || true
xattr -d com.apple.provenance ./COINjecture 2>/dev/null || true
xattr -c ./COINjecture 2>/dev/null || true

# Run COINjecture
exec ./COINjecture "$@"
EOF
chmod +x "$INSTALL_DIR/COINjecture-Safe"

print_success "Security-bypass wrapper created"

# Create launcher script with security bypass
print_status "Creating enhanced launcher script..."
cat > "$INSTALL_DIR/launch_coinjecture.sh" << 'EOF'
#!/bin/bash
# Enhanced COINjecture Launcher with Security Bypass

cd "$HOME/Applications/COINjecture"

# Comprehensive security attribute removal
xattr -d com.apple.quarantine ./COINjecture 2>/dev/null || true
xattr -d com.apple.provenance ./COINjecture 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemWhereFroms ./COINjecture 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemDownloadedDate ./COINjecture 2>/dev/null || true
xattr -c ./COINjecture 2>/dev/null || true

# Run COINjecture
./COINjecture
EOF
chmod +x "$INSTALL_DIR/launch_coinjecture.sh"

# Create desktop shortcut with security bypass
print_status "Creating enhanced desktop shortcut..."
cat > "$HOME/Desktop/COINjecture.command" << 'EOF'
#!/bin/bash
# COINjecture Desktop Launcher with Security Bypass

cd "$HOME/Applications/COINjecture"

# Remove all security attributes
xattr -d com.apple.quarantine ./COINjecture 2>/dev/null || true
xattr -d com.apple.provenance ./COINjecture 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemWhereFroms ./COINjecture 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemDownloadedDate ./COINjecture 2>/dev/null || true
xattr -c ./COINjecture 2>/dev/null || true

# Run COINjecture
./COINjecture
EOF
chmod +x "$HOME/Desktop/COINjecture.command"

print_success "Enhanced desktop shortcut created: ~/Desktop/COINjecture.command"

# Create Applications folder shortcut
print_status "Creating Applications folder shortcut..."
ln -sf "$INSTALL_DIR/COINjecture-Safe" "$HOME/Applications/COINjecture-CLI"

print_success "Applications shortcut created"

# Create a system-wide security bypass script
print_status "Creating system-wide security bypass..."
cat > "$INSTALL_DIR/fix-security.sh" << 'EOF'
#!/bin/bash
# COINjecture Security Fix Script
# Run this if you ever get security warnings again

echo "🔒 Fixing COINjecture security attributes..."

cd "$HOME/Applications/COINjecture"

# Remove ALL security attributes
xattr -d com.apple.quarantine ./COINjecture 2>/dev/null || true
xattr -d com.apple.provenance ./COINjecture 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemWhereFroms ./COINjecture 2>/dev/null || true
xattr -d com.apple.metadata:kMDItemDownloadedDate ./COINjecture 2>/dev/null || true
xattr -c ./COINjecture 2>/dev/null || true

# Set proper permissions
chmod 755 ./COINjecture

echo "✅ Security attributes cleared!"
echo "✅ COINjecture should now run without warnings!"
EOF
chmod +x "$INSTALL_DIR/fix-security.sh"

print_success "Security fix script created: $INSTALL_DIR/fix-security.sh"

# Create uninstaller
print_status "Creating enhanced uninstaller..."
cat > "$INSTALL_DIR/uninstall.sh" << 'EOF'
#!/bin/bash
echo "🗑️  Uninstalling COINjecture..."
rm -rf "$HOME/Applications/COINjecture"
rm -f "$HOME/Desktop/COINjecture.command"
rm -f "$HOME/Applications/COINjecture-CLI"
echo "✅ COINjecture uninstalled successfully"
EOF
chmod +x "$INSTALL_DIR/uninstall.sh"

print_success "Enhanced uninstaller created: $INSTALL_DIR/uninstall.sh"

# Final security check
print_special "🔍 FINAL SECURITY CHECK..."
if xattr -l "$INSTALL_DIR/COINjecture" 2>/dev/null | grep -q "com.apple"; then
    print_warning "Some security attributes still present - running final cleanup..."
    xattr -c "$INSTALL_DIR/COINjecture" 2>/dev/null || true
else
    print_success "✅ All security attributes successfully removed!"
fi

echo ""
echo "🎉 Enhanced Installation Complete!"
echo "================================"
echo ""
echo "📁 Installation Location: $INSTALL_DIR"
echo "🖥️  Desktop Shortcut: ~/Desktop/COINjecture.command"
echo "📱 Applications Shortcut: ~/Applications/COINjecture-CLI"
echo "🔒 Security Fix Script: $INSTALL_DIR/fix-security.sh"
echo ""
echo "🚀 How to Run (Multiple Options):"
echo "   • Double-click desktop shortcut (RECOMMENDED)"
echo "   • Run: $INSTALL_DIR/COINjecture-Safe"
echo "   • Run: $INSTALL_DIR/launch_coinjecture.sh"
echo "   • Run: $INSTALL_DIR/COINjecture"
echo ""
echo "🔒 Security Warning Solutions:"
echo "   • All launchers automatically remove security attributes"
echo "   • If you still see warnings, run: $INSTALL_DIR/fix-security.sh"
echo "   • Right-click → 'Open' (one-time override)"
echo ""
echo "🗑️  To Uninstall:"
echo "   Run: $INSTALL_DIR/uninstall.sh"
echo ""
print_success "COINjecture is ready to use with ZERO security warnings! 🎯"
