package player

import (
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
	"clispot/internal/settings"
)


type Player struct {
	context     *oto.Context
	player      oto.Player
	isPlaying   bool
	isPaused    bool
	currentSong string
	volume      float64
	position    time.Duration
	duration    time.Duration
	startTime   time.Time
	pausedTime  time.Duration
	repeatMode  settings.RepeatMode
}


type PlaybackState struct {
	IsPlaying   bool
	IsPaused    bool
	CurrentSong string
	Position    time.Duration
	Duration    time.Duration
	Volume      float64
	RepeatMode  settings.RepeatMode
}


func NewPlayer() *Player {
	
	ctx, ready, err := oto.NewContext(44100, 2, 2)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize audio context: %v", err))
	}
	<-ready

	return &Player{
		context:    ctx,
		isPlaying:  false,
		isPaused:   false,
		volume:     0.8, 
		repeatMode: settings.RepeatNone,
	}
}


func (p *Player) Play(filePath string) error {
	
	p.Stop()

	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return fmt.Errorf("failed to decode MP3: %v", err)
	}

	
	
	sampleRate := decoder.SampleRate()
	length := decoder.Length()
	if length > 0 && sampleRate > 0 {
		p.duration = time.Duration(length/int64(sampleRate)/4) * time.Second 
	} else {
		p.duration = 0 
	}

	
	file.Close()
	file, err = os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to reopen file: %v", err)
	}

	
	decoder, err = mp3.NewDecoder(file)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to decode MP3 for playback: %v", err)
	}

	
	p.player = p.context.NewPlayer(decoder)
	p.currentSong = filePath
	p.isPlaying = true
	p.isPaused = false
	p.startTime = time.Now()
	p.pausedTime = 0

	
	p.player.Play()

	return nil
}


func (p *Player) Pause() {
	if p.player != nil && p.isPlaying && !p.isPaused {
		p.player.Pause()
		p.isPaused = true
		
		p.updatePosition()
	}
}


func (p *Player) Resume() {
	if p.player != nil && p.isPlaying && p.isPaused {
		p.player.Play()
		p.isPaused = false
		
		p.startTime = time.Now().Add(-p.position)
	}
}

// Stop stops the current playback completely (original behavior)
func (p *Player) Stop() {
	if p.player != nil {
		p.player.Close()
		p.player = nil
	}
	p.isPlaying = false
	p.isPaused = false
	p.currentSong = ""
	p.position = 0
	p.pausedTime = 0
}

// PauseAndStop pauses the song and shows it as stopped, but keeps position
func (p *Player) PauseAndStop() {
	if p.player != nil && p.isPlaying {
		// Update position before pausing
		p.updatePosition()
		p.player.Pause()
		p.isPaused = true
		p.isPlaying = false // Show as stopped but keep the song loaded
	}
}

// ResumeFromStop resumes from a paused/stopped state
func (p *Player) ResumeFromStop() {
	if p.player != nil && p.currentSong != "" && !p.isPlaying {
		p.player.Play()
		p.isPlaying = true
		p.isPaused = false
		// Adjust start time to account for the current position
		p.startTime = time.Now().Add(-p.position)
	}
}


func (p *Player) TogglePlayPause() {
	if p.isPaused {
		p.Resume()
	} else if p.isPlaying {
		p.Pause()
	}
}


func (p *Player) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}
	p.volume = volume
	
	if p.player != nil {
		p.player.SetVolume(volume)
	}
}


func (p *Player) GetVolume() float64 {
	return p.volume
}


func (p *Player) IsPlaying() bool {
	return p.isPlaying && !p.isPaused
}


func (p *Player) IsPaused() bool {
	return p.isPaused
}


func (p *Player) GetCurrentSong() string {
	return p.currentSong
}


func (p *Player) SetRepeatMode(mode settings.RepeatMode) {
	p.repeatMode = mode
}


func (p *Player) GetRepeatMode() settings.RepeatMode {
	return p.repeatMode
}


func (p *Player) CycleRepeatMode() settings.RepeatMode {
	switch p.repeatMode {
	case settings.RepeatNone:
		p.repeatMode = settings.RepeatSingle
	case settings.RepeatSingle:
		p.repeatMode = settings.RepeatAll
	case settings.RepeatAll:
		p.repeatMode = settings.RepeatNone
	}
	return p.repeatMode
}


func (p *Player) ShouldRepeat() bool {
	return p.repeatMode == settings.RepeatSingle
}


func (p *Player) ShouldRepeatPlaylist() bool {
	return p.repeatMode == settings.RepeatAll
}


func (p *Player) Seek(position time.Duration) error {
	if p.player == nil || !p.isPlaying {
		return fmt.Errorf("no track currently playing")
	}
	
	if position < 0 {
		position = 0
	} else if position > p.duration {
		position = p.duration
	}
	
	
	
	p.position = position
	p.startTime = time.Now().Add(-position)
	
	return nil
}


func (p *Player) updatePosition() {
	if p.isPlaying && !p.isPaused {
		p.position = time.Since(p.startTime)
		if p.position > p.duration {
			p.position = p.duration
		}
	}
}


func (p *Player) GetState() PlaybackState {
	
	if p.isPlaying && !p.isPaused {
		p.updatePosition()
	}
	
	return PlaybackState{
		IsPlaying:   p.isPlaying && !p.isPaused,
		IsPaused:    p.isPaused,
		CurrentSong: p.currentSong,
		Position:    p.position,
		Duration:    p.duration,
		Volume:      p.volume,
		RepeatMode:  p.repeatMode,
	}
}


func (p *Player) IsFinished() bool {
	if p.player == nil {
		return true
	}
	return !p.player.IsPlaying()
}


func (p *Player) Close() {
	p.Stop()
	if p.context != nil {
		
		
	}
}