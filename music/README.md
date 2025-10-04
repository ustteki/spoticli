# Sample Music Directory

# Put your MP3 files here!

This directory is where you should place your MP3 music files.

## Playlist Organization

Create folders to organize your music into playlists:

```
music/
├── Rock/
│   ├── song1.mp3
│   └── song2.mp3
├── Electronic/
│   ├── track1.mp3
│   └── track2.mp3
└── Chill Vibes/
    └── relaxing.mp3
```

Each folder becomes a browsable playlist in the application!

## Supported Formats

- **MP3** files with ID3 tags (recommended)
- Files should have proper metadata (title, artist, album) for the best experience

## Getting Started

1. Add your MP3 files to this directory
2. Organize them into folders by genre, mood, or any way you like
3. Run `spotify` (or `./spotify -dir ./music` if running locally)
4. Browse your music library with folder navigation!

The application will automatically scan this directory and all subdirectories for MP3 files when it starts.