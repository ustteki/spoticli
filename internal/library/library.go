package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/hajimehoshi/go-mp3"
	"clispot/internal/albumart"
)

// ItemType represents the type of item (folder or song)
type ItemType int

const (
	ItemTypeSong ItemType = iota
	ItemTypeFolder
)

// LibraryItem represents either a song or a folder
type LibraryItem struct {
	Type        ItemType
	Name        string
	Path        string
	Song        *Song  // Only set for songs
	SongCount   int    // Only set for folders
}

type Song struct {
	FilePath string
	Title    string
	Artist   string
	Album    string
	Year     string
	Duration time.Duration
	Genre    string
	Track    int
	FileSize int64
	AlbumArt *albumart.ASCIIArt 
	Playlist string // The folder/playlist this song belongs to
}

type Library struct {
	rootPath    string
	songs       []Song
	currentPath string
}


func NewLibrary(rootPath string) *Library {
	return &Library{
		rootPath:    rootPath,
		songs:       make([]Song, 0),
		currentPath: rootPath,
	}
}


func (l *Library) ScanDirectory() ([]Song, error) {
	l.songs = make([]Song, 0)

	err := filepath.Walk(l.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mp3" {
			song, err := l.extractMetadata(path, info)
			if err != nil {
				fmt.Printf("Error reading metadata for %s: %v\n", filepath.Base(path), err)
				song = Song{
					FilePath: path,
					Title:    strings.TrimSuffix(filepath.Base(path), ".mp3"),
					Artist:   "Unknown Artist",
					Album:    "Unknown Album",
					Duration: 0,
				}
			}
			l.songs = append(l.songs, song)
		}
		return nil
	})

	return l.songs, err
}

// GetCurrentItems returns items (folders and songs) in the current directory
func (l *Library) GetCurrentItems() ([]LibraryItem, error) {
	var items []LibraryItem
	
	// If not at root, add ".." to go back
	if l.currentPath != l.rootPath {
		items = append(items, LibraryItem{
			Type: ItemTypeFolder,
			Name: "..",
			Path: filepath.Dir(l.currentPath),
		})
	}
	
	// Read current directory
	entries, err := os.ReadDir(l.currentPath)
	if err != nil {
		return nil, err
	}
	
	// Add folders first
	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(l.currentPath, entry.Name())
			songCount := l.countSongsInFolder(fullPath)
			
			items = append(items, LibraryItem{
				Type:      ItemTypeFolder,
				Name:      entry.Name(),
				Path:      fullPath,
				SongCount: songCount,
			})
		}
	}
	
	// Add songs in current directory
	for _, entry := range entries {
		if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".mp3" {
			fullPath := filepath.Join(l.currentPath, entry.Name())
			
			// Find the song in our scanned songs
			for _, song := range l.songs {
				if song.FilePath == fullPath {
					items = append(items, LibraryItem{
						Type: ItemTypeSong,
						Name: fmt.Sprintf("%s - %s", song.Artist, song.Title),
						Path: fullPath,
						Song: &song,
					})
					break
				}
			}
		}
	}
	
	return items, nil
}

// NavigateToFolder changes the current directory
func (l *Library) NavigateToFolder(path string) error {
	// Validate the path is within our root
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	
	absRoot, err := filepath.Abs(l.rootPath)
	if err != nil {
		return err
	}
	
	if !strings.HasPrefix(absPath, absRoot) {
		return fmt.Errorf("path outside of music library")
	}
	
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist")
	}
	
	l.currentPath = path
	return nil
}

// GetCurrentPath returns the current directory path
func (l *Library) GetCurrentPath() string {
	return l.currentPath
}

// GetRootPath returns the root library path
func (l *Library) GetRootPath() string {
	return l.rootPath
}

// CanGoBack returns true if we can navigate to parent directory
func (l *Library) CanGoBack() bool {
	return l.currentPath != l.rootPath
}

// GetRelativePath returns the path relative to the root
func (l *Library) GetRelativePath() string {
	rel, err := filepath.Rel(l.rootPath, l.currentPath)
	if err != nil {
		return l.currentPath
	}
	if rel == "." {
		return "/"
	}
	return "/" + rel
}

// countSongsInFolder recursively counts MP3 files in a folder
func (l *Library) countSongsInFolder(folderPath string) int {
	count := 0
	filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mp3" {
			count++
		}
		return nil
	})
	return count
}

// GetSongsInCurrentFolder returns only songs in the current folder (no subfolders)
func (l *Library) GetSongsInCurrentFolder() []Song {
	var songs []Song
	
	for _, song := range l.songs {
		if filepath.Dir(song.FilePath) == l.currentPath {
			songs = append(songs, song)
		}
	}
	
	return songs
}


func (l *Library) extractMetadata(filePath string, fileInfo os.FileInfo) (Song, error) {
	
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return Song{}, err
	}
	defer tag.Close()

	
	title := tag.Title()
	artist := tag.Artist()
	album := tag.Album()
	year := tag.Year()
	genre := tag.Genre()

	
	if title == "" {
		title = l.getFileNameWithoutExt(fileInfo.Name())
	}

	
	if artist == "" {
		artist = "Unknown Artist"
	}

	
	if album == "" {
		album = "Unknown Album"
	}

	
	trackText := tag.GetTextFrame(tag.CommonID("Track number/Position in set")).Text
	track := 0
	if trackText != "" {
		
		if idx := strings.Index(trackText, "/"); idx > 0 {
			trackText = trackText[:idx]
		}
		fmt.Sscanf(trackText, "%d", &track)
	}

	// Determine playlist from folder structure
	playlistName := l.getPlaylistName(filePath)

	song := Song{
		FilePath: filePath,
		Title:    title,
		Artist:   artist,
		Album:    album,
		Year:     year,
		Genre:    genre,
		Track:    track,
		FileSize: fileInfo.Size(),
		Duration: l.getActualDuration(filePath),
		AlbumArt: l.extractAlbumArt(tag, title, artist),
		Playlist: playlistName,
	}

	return song, nil
}


func (l *Library) extractAlbumArt(tag *id3v2.Tag, title, artist string) *albumart.ASCIIArt {
	
	converter := albumart.NewConverter(32, 16)
	
	
	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	if len(pictures) > 0 {
		if picture, ok := pictures[0].(id3v2.PictureFrame); ok {
			
			if art, err := converter.ConvertImageToASCII(picture.Picture); err == nil {
				return art
			}
		}
	}
	
	
	return albumart.CreateSimpleArt(title, artist, 32, 16)
}


func (l *Library) getFileNameWithoutExt(filename string) string {
	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}


func (l *Library) estimateDuration(fileSize int64) time.Duration {
	// More accurate estimation for MP3 files
	// Assume average bitrate of 128 kbps (16 KB/s)
	// This is still an estimate, but much more accurate
	if fileSize < 1024 {
		return 0 // Very small files
	}
	
	// Calculate duration based on average MP3 bitrate
	// 128 kbps = 16 KB/s, but account for metadata overhead
	avgBytesPerSecond := int64(14000) // Slightly lower to account for metadata
	durationSeconds := fileSize / avgBytesPerSecond
	
	// Cap maximum duration to something reasonable (2 hours)
	if durationSeconds > 7200 {
		durationSeconds = 7200
	}
	
	return time.Duration(durationSeconds) * time.Second
}

// getActualDuration tries to get the actual duration from the MP3 file
func (l *Library) getActualDuration(filePath string) time.Duration {
	// Try to open and decode the MP3 file to get actual duration
	file, err := os.Open(filePath)
	if err != nil {
		// Fallback to file size estimation
		if fileInfo, statErr := os.Stat(filePath); statErr == nil {
			return l.estimateDuration(fileInfo.Size())
		}
		return 0
	}
	defer file.Close()

	// Create MP3 decoder
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		// Fallback to file size estimation
		if fileInfo, statErr := os.Stat(filePath); statErr == nil {
			return l.estimateDuration(fileInfo.Size())
		}
		return 0
	}

	// Get the actual length and sample rate
	length := decoder.Length()
	sampleRate := decoder.SampleRate()
	
	if length > 0 && sampleRate > 0 {
		// Calculate duration: length is in bytes, and we have 4 bytes per sample (16-bit stereo)
		samples := length / 4
		durationSeconds := float64(samples) / float64(sampleRate)
		return time.Duration(durationSeconds * float64(time.Second))
	}

	// If we can't get length from decoder, fallback to estimation
	if fileInfo, statErr := os.Stat(filePath); statErr == nil {
		return l.estimateDuration(fileInfo.Size())
	}
	
	return 0
}


func (l *Library) GetSongs() []Song {
	return l.songs
}


func (l *Library) GetSongsByArtist(artist string) []Song {
	var result []Song
	for _, song := range l.songs {
		if strings.EqualFold(song.Artist, artist) {
			result = append(result, song)
		}
	}
	return result
}


func (l *Library) GetSongsByAlbum(album string) []Song {
	var result []Song
	for _, song := range l.songs {
		if strings.EqualFold(song.Album, album) {
			result = append(result, song)
		}
	}
	return result
}


func (l *Library) SearchSongs(query string) []Song {
	var result []Song
	query = strings.ToLower(query)
	
	for _, song := range l.songs {
		if strings.Contains(strings.ToLower(song.Title), query) ||
		   strings.Contains(strings.ToLower(song.Artist), query) ||
		   strings.Contains(strings.ToLower(song.Album), query) {
			result = append(result, song)
		}
	}
	return result
}


func (l *Library) GetUniqueArtists() []string {
	artistMap := make(map[string]bool)
	for _, song := range l.songs {
		artistMap[song.Artist] = true
	}
	
	var artists []string
	for artist := range artistMap {
		artists = append(artists, artist)
	}
	return artists
}


func (l *Library) GetUniqueAlbums() []string {
	albumMap := make(map[string]bool)
	for _, song := range l.songs {
		albumMap[song.Album] = true
	}
	
	var albums []string
	for album := range albumMap {
		albums = append(albums, album)
	}
	return albums
}

// getPlaylistName determines the playlist name from the file path
func (l *Library) getPlaylistName(filePath string) string {
	dir := filepath.Dir(filePath)
	
	// If the file is directly in the root music folder, no playlist
	if dir == l.rootPath {
		return ""
	}
	
	// Get the immediate parent folder name as playlist
	rel, err := filepath.Rel(l.rootPath, dir)
	if err != nil {
		return ""
	}
	
	// If nested folders, use the top-level folder as playlist name
	parts := strings.Split(rel, string(filepath.Separator))
	return parts[0]
}