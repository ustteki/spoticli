# Sample Music Directory

This directory is where you should place your MP3 files.

## Recommended Structure:

```
music/
├── Artist Name/
│   ├── Album Name/
│   │   ├── 01 - Track Name.mp3
│   │   ├── 02 - Another Track.mp3
│   │   └── 03 - Final Track.mp3
│   └── Another Album/
│       ├── 01 - First Song.mp3
│       └── 02 - Second Song.mp3
└── Another Artist/
    └── Their Album/
        ├── 01 - Opening Track.mp3
        ├── 02 - Middle Track.mp3
        └── 03 - Closing Track.mp3
```

## Getting Started:

1. Copy your MP3 files into this directory
2. Organize them by Artist/Album if desired (optional but recommended)
3. Run the music player: `clispot`
4. The app will automatically scan and index all MP3 files

## Features:

- **Album Art Display** - ASCII art representations of album covers
- **Yellow Color Theme** - Faint yellow highlighting for currently playing music
- **Search Functionality** - Press `/` to search by title, artist, or album
- **Keyboard Controls** - Full keyboard navigation and control

## Notes:

- Files are scanned recursively, so any folder structure works
- Only MP3 files are currently supported
- Large libraries may take a moment to scan initially
- The player remembers your music library between sessions