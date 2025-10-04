#!/bin/bash

# Spotify CLI Installation Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🎵 Installing Spotify CLI...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    echo -e "${YELLOW}Visit: https://golang.org/doc/install${NC}"
    exit 1
fi

# Build the application
echo -e "${YELLOW}📦 Building application...${NC}"
go build -o spotify ./cmd/main.go

# Make it executable
chmod +x spotify

# Determine installation directory
if [[ ":$PATH:" == *":/usr/local/bin:"* ]]; then
    INSTALL_DIR="/usr/local/bin"
elif [[ ":$PATH:" == *":$HOME/.local/bin:"* ]]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
elif [[ ":$PATH:" == *":$HOME/bin:"* ]]; then
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    echo -e "${YELLOW}⚠️  Created $INSTALL_DIR - you may need to add it to your PATH${NC}"
    echo -e "${YELLOW}   Add this to your ~/.zshrc or ~/.bashrc:${NC}"
    echo -e "${YELLOW}   export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
fi

# Copy the binary
echo -e "${YELLOW}📋 Installing to $INSTALL_DIR...${NC}"
if [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
    sudo cp spotify "$INSTALL_DIR/"
    echo -e "${GREEN}✅ Installed spotify to $INSTALL_DIR (requires sudo)${NC}"
else
    cp spotify "$INSTALL_DIR/"
    echo -e "${GREEN}✅ Installed spotify to $INSTALL_DIR${NC}"
fi

# Create default music directory
MUSIC_DIR="$HOME/Music/spotify-cli"
if [[ ! -d "$MUSIC_DIR" ]]; then
    mkdir -p "$MUSIC_DIR"
    echo -e "${GREEN}📁 Created default music directory: $MUSIC_DIR${NC}"
fi

# Create a sample folder structure
mkdir -p "$MUSIC_DIR/Rock" "$MUSIC_DIR/Pop" "$MUSIC_DIR/Electronic" "$MUSIC_DIR/Jazz"

echo -e "${GREEN}🎉 Installation complete!${NC}"
echo ""
echo -e "${BLUE}Usage:${NC}"
echo -e "  ${YELLOW}spotify${NC}                    # Use default music directory"
echo -e "  ${YELLOW}spotify -dir /path/to/music${NC} # Use custom directory"
echo ""
echo -e "${BLUE}Music Directory:${NC} $MUSIC_DIR"
echo -e "${BLUE}Playlist System:${NC} Create folders in your music directory"
echo -e "  • Each folder becomes a playlist"
echo -e "  • Navigate with Enter/Backspace"
echo -e "  • Add MP3 files to folders to organize by genre/mood"
echo ""
echo -e "${BLUE}Controls:${NC}"
echo -e "  • ${GREEN}Enter${NC} - Play song or browse folder"
echo -e "  • ${GREEN}Space${NC} - Play/Pause"
echo -e "  • ${GREEN}Backspace${NC} - Go back to parent folder"
echo -e "  • ${GREEN}n/p${NC} - Next/Previous song"
echo -e "  • ${GREEN}+/-${NC} - Volume up/down"
echo -e "  • ${GREEN}q${NC} - Quit"

# Test the installation
echo ""
echo -e "${YELLOW}🧪 Testing installation...${NC}"
if command -v spotify &> /dev/null; then
    echo -e "${GREEN}✅ 'spotify' command is available!${NC}"
    echo -e "${BLUE}Run ${YELLOW}'spotify'${BLUE} to start the music player${NC}"
else
    echo -e "${RED}❌ Command not found. You may need to:"
    echo -e "   1. Restart your terminal"
    echo -e "   2. Add $INSTALL_DIR to your PATH"
    echo -e "   3. Run: export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
fi