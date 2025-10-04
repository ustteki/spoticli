package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)


type RepeatMode int

const (
	RepeatNone RepeatMode = iota
	RepeatSingle
	RepeatAll
)


func (r RepeatMode) String() string {
	switch r {
	case RepeatNone:
		return "Off"
	case RepeatSingle:
		return "Single"
	case RepeatAll:
		return "All"
	default:
		return "Off"
	}
}


type Settings struct {
	
	ShowProgressBar    bool       `json:"show_progress_bar"`
	
	
	RepeatMode         RepeatMode `json:"repeat_mode"`
	Volume             float64    `json:"volume"`
	
	
	Theme              string     `json:"theme"`
	CompactMode        bool       `json:"compact_mode"`
	
	
	BufferSize         int        `json:"buffer_size"`
	UpdateInterval     int        `json:"update_interval_ms"`
}


func DefaultSettings() *Settings {
	return &Settings{
		ShowProgressBar:   false,
		RepeatMode:        RepeatNone,
		Volume:            0.8,
		Theme:             "default",
		CompactMode:       false,
		BufferSize:        4096,
		UpdateInterval:    100,
	}
}


type Manager struct {
	settings   *Settings
	configPath string
}


func NewManager() *Manager {
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	configDir := filepath.Join(homeDir, ".config", "clispot")
	os.MkdirAll(configDir, 0755)
	
	configPath := filepath.Join(configDir, "settings.json")
	
	manager := &Manager{
		settings:   DefaultSettings(),
		configPath: configPath,
	}
	
	
	manager.Load()
	
	return manager
}


func (m *Manager) Get() *Settings {
	return m.settings
}


func (m *Manager) Update(updateFunc func(*Settings)) error {
	updateFunc(m.settings)
	return m.Save()
}


func (m *Manager) ToggleProgressBar() bool {
	m.settings.ShowProgressBar = !m.settings.ShowProgressBar
	m.Save()
	return m.settings.ShowProgressBar
}


func (m *Manager) CycleRepeatMode() RepeatMode {
	switch m.settings.RepeatMode {
	case RepeatNone:
		m.settings.RepeatMode = RepeatSingle
	case RepeatSingle:
		m.settings.RepeatMode = RepeatAll
	case RepeatAll:
		m.settings.RepeatMode = RepeatNone
	}
	m.Save()
	return m.settings.RepeatMode
}


func (m *Manager) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}
	m.settings.Volume = volume
	m.Save()
}


func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		
		return nil
	}
	
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("error parsing settings: %v", err)
	}
	
	m.settings = &settings
	return nil
}


func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling settings: %v", err)
	}
	
	return os.WriteFile(m.configPath, data, 0644)
}


func (m *Manager) GetConfigPath() string {
	return m.configPath
}