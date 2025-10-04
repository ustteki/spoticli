package playlist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"clispot/internal/library"
)


type Playlist struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Songs       []library.Song  `json:"songs"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}


type Manager struct {
	playlists   []Playlist
	playlistDir string
}


func NewManager(playlistDir string) *Manager {
	
	os.MkdirAll(playlistDir, 0755)
	
	return &Manager{
		playlists:   make([]Playlist, 0),
		playlistDir: playlistDir,
	}
}


func (m *Manager) CreatePlaylist(name, description string) (*Playlist, error) {
	
	for _, playlist := range m.playlists {
		if playlist.Name == name {
			return nil, fmt.Errorf("playlist '%s' already exists", name)
		}
	}
	
	playlist := Playlist{
		Name:        name,
		Description: description,
		Songs:       make([]library.Song, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	m.playlists = append(m.playlists, playlist)
	
	
	if err := m.savePlaylist(&playlist); err != nil {
		return nil, fmt.Errorf("error saving playlist: %v", err)
	}
	
	return &playlist, nil
}


func (m *Manager) AddSongToPlaylist(playlistName string, song library.Song) error {
	for i := range m.playlists {
		if m.playlists[i].Name == playlistName {
			
			for _, existingSong := range m.playlists[i].Songs {
				if existingSong.FilePath == song.FilePath {
					return fmt.Errorf("song already exists in playlist")
				}
			}
			
			m.playlists[i].Songs = append(m.playlists[i].Songs, song)
			m.playlists[i].UpdatedAt = time.Now()
			
			return m.savePlaylist(&m.playlists[i])
		}
	}
	
	return fmt.Errorf("playlist '%s' not found", playlistName)
}


func (m *Manager) GetPlaylist(name string) (*Playlist, error) {
	for i := range m.playlists {
		if m.playlists[i].Name == name {
			return &m.playlists[i], nil
		}
	}
	
	return nil, fmt.Errorf("playlist '%s' not found", name)
}


func (m *Manager) GetAllPlaylists() []Playlist {
	return m.playlists
}


func (m *Manager) savePlaylist(playlist *Playlist) error {
	filename := m.getPlaylistFileName(playlist.Name)
	
	data, err := json.MarshalIndent(playlist, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling playlist: %v", err)
	}
	
	return os.WriteFile(filename, data, 0644)
}


func (m *Manager) getPlaylistFileName(name string) string {
	
	safeName := filepath.Base(name)
	if safeName == "" || safeName == "." || safeName == ".." {
		safeName = "playlist"
	}
	
	return filepath.Join(m.playlistDir, safeName+".json")
}