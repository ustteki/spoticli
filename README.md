# CLiSpot 🎵

A terminal-based music player written in Go that brings the Spotify experience to your command line. Play your local MP3 collection with a beautiful, interactive interface featuring album art, visualizers, and full playback controls.

## ✨ Features

### 🎧 Core Playback
- **High-quality audio playback** using Go's oto/v2 library
- **MP3 support** with ID3 metadata extraction
- **Play/pause, stop, next/previous** track controls
- **Volume control** with +/- keys
- **Playlist management** with automatic library scanning

### 🎨 Visual Experience
- **ASCII album art** displayed in the interface
- **Real-time audio visualizer** with frequency-based animation
- **Progress bar** with seeking capability (click to seek)
- **Color-coded UI** with yellow highlights for playing tracks
- **Dynamic status updates** showing current playback state

### ⚙️ Advanced Features
- **Settings system** with persistent configuration
- **Repeat modes**: None, Single track, All tracks (cycle with 'L')
- **Toggleable components**: Progress bar (B key) and visualizer (V key)
- **Search functionality** to filter your music library
- **Keyboard shortcuts** for all operations

### 🔧 Technical Features
- **Fast library scanning** with metadata caching
- **Responsive terminal UI** built with tview
- **Cross-platform support** (macOS, Linux, Windows)
- **Configuration persistence** in ~/.config/clispot/

## 🚀 Installation

### Quick Install
```bash
# Clone the repository
git clone https://github.com/ustteki/spoticli

cd spoticli

# Build and install globally
go build -o clispot ./cmd
cp clispot ~/bin/  # Make sure ~/bin is in your PATH
```

### Usage
```bash
# Navigate to a directory containing MP3 files and run:
clispot

# Or specify a music directory:
clispot /path/to/your/music
```

## 🎮 Controls

### Basic Playback
| Key | Action |
|-----|--------|
| `Space` | Play/Pause current track |
| `Enter` | Play selected track |
| `n` | Next track |
| `p` | Previous track |
| `s` | Stop playback |

### Navigation
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate song list |
| `/` | Search mode |
| `Esc` | Exit search mode |

### Volume & Settings
| Key | Action |
|-----|--------|
| `+/=` | Increase volume |
| `-` | Decrease volume |
| `L` | Cycle repeat mode (None → Single → All → None) |
| `V` | Toggle audio visualizer on/off |
| `B` | Toggle progress bar on/off |
| `?` | Show current settings |

### Other
| Key | Action |
|-----|--------|
| `q` | Quit application |

## 🎵 Features in Detail

### Audio Visualizer
- **Real-time frequency display** with ASCII characters
- **Color-coded bars**: Red/Yellow (bass), Green/Blue (mid), Cyan/Blue (treble)
- **Peak hold indicators** for visual appeal
- **Beat-responsive animation** that pulses with the music
- **Graceful fade-out** when pausing or stopping

### Progress Bar
- **Time display** showing current position and total duration
- **Visual progress indicator** with color coding
- **Percentage display** for precise position tracking
- **Seeking support** (basic implementation)

### Repeat Modes
- **None**: Play through playlist once
- **Single**: Repeat current track indefinitely
- **All**: Repeat entire playlist when finished

### Settings System
- **Persistent configuration** saved to `~/.config/clispot/settings.json`
- **Toggleable UI components** with instant visual feedback
- **Volume preferences** remembered between sessions
- **Customizable visualizer height** and update intervals

## 📁 Project Structure

```
spoticli/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── albumart/
│   │   └── converter.go     # ASCII art conversion
│   ├── library/
│   │   └── library.go       # Music library scanning
│   ├── player/
│   │   └── player.go        # Audio playback engine
│   ├── playlist/
│   │   └── playlist.go      # Playlist management
│   ├── progressbar/
│   │   └── progressbar.go   # Progress bar component
│   ├── settings/
│   │   └── settings.go      # Settings persistence
│   ├── ui/
│   │   └── ui.go           # Terminal user interface
│   └── visualizer/
│       └── visualizer.go    # Audio visualizer
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
└── README.md               # This file
```

## 🛠️ Technical Details

### Dependencies
- `github.com/hajimehoshi/oto/v2` - Audio playback
- `github.com/hajimehoshi/go-mp3` - MP3 decoding
- `github.com/bogem/id3v2/v2` - ID3 metadata parsing
- `github.com/rivo/tview` - Terminal UI framework
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/nfnt/resize` - Image resizing for album art

### Performance
- **Efficient audio streaming** with configurable buffer sizes
- **Optimized UI updates** at 100ms intervals for smooth visualizer
- **Lazy album art loading** to improve startup time
- **Memory-efficient** metadata caching

## 🎨 Customization

### Settings File Location
`~/.config/clispot/settings.json`

### Example Settings
```json
{
  "show_progress_bar": true,
  "show_visualizer": false,
  "visualizer_height": 6,
  "repeat_mode": 0,
  "volume": 0.8,
  "theme": "default",
  "compact_mode": false,
  "buffer_size": 4096,
  "update_interval_ms": 100
}
```

## 🔧 Troubleshooting

### Audio Issues
- **No sound**: Check your system's audio settings and volume
- **Crackling audio**: Try adjusting the buffer size in settings
- **Slow playback**: Ensure your MP3 files aren't corrupted

### UI Issues
- **Display problems**: Ensure your terminal supports color and Unicode
- **Layout issues**: Try resizing your terminal window
- **Missing album art**: Ensure your MP3 files have embedded artwork

### Performance Issues
- **Slow scanning**: Large music libraries may take time to scan initially
- **High CPU usage**: Disable the visualizer if system resources are limited

## 🤝 Contributing

Contributions are welcome! Areas for improvement:

- **Additional audio formats** (FLAC, OGG, etc.)
- **Enhanced visualizer effects** and themes
- **Playlist import/export** functionality
- **Remote control** via web interface
- **Plugin system** for extensibility

## 📝 License

This project is open source. Feel free to use, modify, and distribute.

## 🙏 Acknowledgments

- **oto/v2** team for excellent Go audio library
- **tview** developers for the terminal UI framework
- **Spotify** for the original inspiration

---

**CLiSpot** - Because music sounds better in the terminal! 🎵✨
