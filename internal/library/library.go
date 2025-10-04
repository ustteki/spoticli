package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"clispot/internal/albumart"
)


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
}


type Library struct {
	rootPath string
	songs    []Song
}


func NewLibrary(rootPath string) *Library {
	return &Library{
		rootPath: rootPath,
		songs:    make([]Song, 0),
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

	song := Song{
		FilePath: filePath,
		Title:    title,
		Artist:   artist,
		Album:    album,
		Year:     year,
		Genre:    genre,
		Track:    track,
		FileSize: fileInfo.Size(),
		Duration: l.estimateDuration(fileInfo.Size()),
		AlbumArt: nil, 
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
	
	
	durationSeconds := fileSize / 16000
	return time.Duration(durationSeconds) * time.Second
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