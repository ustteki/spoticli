package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	
	"clispot/internal/library"
	"clispot/internal/player"
	"clispot/internal/settings"
	"clispot/internal/progressbar"
)

type App struct {
	app         *tview.Application
	songs       []library.Song
	player      *player.Player
	currentIdx  int
	
	settingsManager *settings.Manager
	progressBar     *progressbar.ProgressBar
	
	songList        *tview.List
	infoPanel       *tview.TextView
	statusBar       *tview.TextView
	helpText        *tview.TextView
	searchInput     *tview.InputField
	progressPanel   *tview.TextView
	
	isSearchMode bool
	filteredSongs []library.Song
}

func NewApp(songs []library.Song, audioPlayer *player.Player) *App {
	settingsManager := settings.NewManager()
	
	return &App{
		app:             tview.NewApplication(),
		songs:           songs,
		player:          audioPlayer,
		currentIdx:      0,
		filteredSongs:   songs,
		settingsManager: settingsManager,
		progressBar:     progressbar.NewProgressBar(80),  // Default width
	}
}

func (a *App) Run() error {
	a.setupUI()
	a.setupKeyBindings()
	a.populateSongList()
	a.updateInfoPanel()
	a.updateStatusBar()
	
	go a.updateLoop()
	
	return a.app.Run()
}

func (a *App) setupUI() {
	a.songList = tview.NewList().ShowSecondaryText(true)
	a.infoPanel = tview.NewTextView().SetDynamicColors(true)
	a.statusBar = tview.NewTextView().SetDynamicColors(true)
	a.helpText = tview.NewTextView().SetDynamicColors(true)
	a.searchInput = tview.NewInputField().SetLabel("Search: ")
	a.progressPanel = tview.NewTextView().SetDynamicColors(true)
	
	a.songList.SetBorder(true).SetTitle(" Songs ")
	a.infoPanel.SetBorder(true).SetTitle(" Now Playing ")
	a.statusBar.SetBorder(false)
	a.helpText.SetBorder(true).SetTitle(" Controls ")
	a.searchInput.SetBorder(true).SetTitle(" Search ")
	a.progressPanel.SetBorder(false)
	
	helpStr := `[yellow]Controls:[white]
[green]Space[white] - Play/Pause    [green]Enter[white] - Play selected
[green]↑/↓[white] - Navigate        [green]n[white] - Next song
[green]p[white] - Previous song     [green]/[white] - Search
[green]Esc[white] - Exit search     [green]q[white] - Quit
[green]+/-[white] - Volume up/down  [green]s[white] - Stop
[green]L[white] - Repeat mode       [green]B[white] - Toggle progress
[green]?[white] - Settings info`
	
	a.helpText.SetText(helpStr)
	
	a.searchInput.SetText("")
	
	a.updateComponentVisibility()
	
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.songList, 0, 3, true).
		AddItem(a.searchInput, 3, 0, false)
	
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.infoPanel, 0, 2, false).
		AddItem(a.helpText, 12, 0, false)
	
	mainPanel := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftPanel, 0, 2, true).
		AddItem(rightPanel, 0, 1, false)
	
	var root tview.Primitive
	if a.settingsManager.Get().ShowProgressBar {
		root = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(mainPanel, 0, 1, true).
			AddItem(a.progressPanel, 1, 0, false).
			AddItem(a.statusBar, 1, 0, false)
	} else {
		root = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(mainPanel, 0, 1, true).
			AddItem(a.statusBar, 1, 0, false)
	}
	
	a.app.SetRoot(root, true)
}

func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.isSearchMode {
			switch event.Key() {
			case tcell.KeyEscape:
				a.exitSearchMode()
				return nil
			case tcell.KeyEnter:
				a.performSearch()
				a.exitSearchMode()
				return nil
			}
			return event // Let search input handle other keys
		}
		
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				a.app.Stop()
				return nil
			case ' ':
				a.togglePlayPause()
				return nil
			case 'n', 'N':
				a.nextSong()
				return nil
			case 'p', 'P':
				a.previousSong()
				return nil
			case 's', 'S':
				a.stopPlayback()
				return nil
			case '/':
				a.enterSearchMode()
				return nil
			case '+', '=':
				a.increaseVolume()
				return nil
			case '-':
				a.decreaseVolume()
				return nil
			case 'l', 'L':
				a.cycleRepeatMode()
				return nil
			case 'b', 'B':
				a.toggleProgressBar()
				return nil
			case '?':
				a.showSettingsInfo()
				return nil
				return nil
			}
		case tcell.KeyEnter:
			a.playSelected()
			return nil
		}
		return event
	})
	
	a.songList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		a.playSelected()
	})
}

func (a *App) populateSongList() {
	a.songList.Clear()
	
	for i, song := range a.filteredSongs {
		mainText := fmt.Sprintf("%s - %s", song.Artist, song.Title)
		secondaryText := fmt.Sprintf("%s | %s", song.Album, a.formatDuration(song.Duration))
		
		if a.player.GetCurrentSong() == song.FilePath {
			mainText = fmt.Sprintf("[yellow]♪ %s[white]", mainText)
		}
		
		a.songList.AddItem(mainText, secondaryText, 0, nil)
		
		if a.player.GetCurrentSong() == song.FilePath {
			a.songList.SetCurrentItem(i)
		}
	}
}

func (a *App) updateInfoPanel() {
	if len(a.filteredSongs) == 0 {
		a.infoPanel.SetText("[red]No songs found[white]")
		return
	}
	
	state := a.player.GetState()
	
	if state.CurrentSong == "" {
		currentIdx := a.songList.GetCurrentItem()
		if currentIdx >= 0 && currentIdx < len(a.filteredSongs) {
			song := a.filteredSongs[currentIdx]
			
			var info strings.Builder
			info.WriteString("[yellow]Selected:[white]\n")
			
			if song.AlbumArt != nil {
				info.WriteString(song.AlbumArt.GetColorizedASCII())
				info.WriteString("\n\n")
			}
			
			info.WriteString(fmt.Sprintf(`[green]Title:[white] %s
[green]Artist:[white] %s
[green]Album:[white] %s
[green]Year:[white] %s
[green]Genre:[white] %s
[green]Duration:[white] %s
[green]File:[white] %s

[dim]Press Enter to play[white]`,
				song.Title, song.Artist, song.Album, song.Year, 
				song.Genre, a.formatDuration(song.Duration),
				filepath.Base(song.FilePath)))
			
			a.infoPanel.SetText(info.String())
		}
	} else {
		var currentSong *library.Song
		for _, song := range a.filteredSongs {
			if song.FilePath == state.CurrentSong {
				currentSong = &song
				break
			}
		}
		
		if currentSong != nil {
			status := "[red]Stopped[white]"
			if state.IsPlaying {
				status = "[yellow]♪ Playing[white]"
			} else if state.IsPaused {
				status = "[yellow]⏸ Paused[white]"
			}
			
			var info strings.Builder
			info.WriteString(status + "\n")
			
			if currentSong.AlbumArt != nil {
				info.WriteString(currentSong.AlbumArt.GetColorizedASCII())
				info.WriteString("\n\n")
			}
			
			info.WriteString(fmt.Sprintf(`[yellow]Title:[white] %s
[yellow]Artist:[white] %s
[yellow]Album:[white] %s
[yellow]Year:[white] %s
[yellow]Genre:[white] %s
[yellow]Duration:[white] %s
[yellow]Volume:[white] %.0f%%
[yellow]Repeat:[white] %s
[yellow]File:[white] %s`,
				currentSong.Title, currentSong.Artist, currentSong.Album,
				currentSong.Year, currentSong.Genre, a.formatDuration(currentSong.Duration),
				state.Volume*100, repeatModeToString(state.RepeatMode), filepath.Base(currentSong.FilePath)))
			
			a.infoPanel.SetText(info.String())
		}
	}
}

func (a *App) updateStatusBar() {
	totalSongs := len(a.songs)
	filteredSongs := len(a.filteredSongs)
	
	var statusText string
	if filteredSongs == totalSongs {
		statusText = fmt.Sprintf(" CLiSpot | %d songs total", totalSongs)
	} else {
		statusText = fmt.Sprintf(" CLiSpot | %d/%d songs (filtered)", filteredSongs, totalSongs)
	}
	
	state := a.player.GetState()
	if state.IsPlaying || state.IsPaused {
		progress := ""
		if state.Duration > 0 {
			percent := float64(state.Position) / float64(state.Duration) * 100
			progress = fmt.Sprintf(" | %.1f%%", percent)
		}
		statusText += progress
	}
	
	if a.isSearchMode {
		statusText += " | [yellow]SEARCH MODE[white]"
	}
	
	a.statusBar.SetText(statusText)
}

func (a *App) playSelected() {
	currentIdx := a.songList.GetCurrentItem()
	if currentIdx >= 0 && currentIdx < len(a.filteredSongs) {
		song := a.filteredSongs[currentIdx]
		a.currentIdx = currentIdx
		
		if err := a.player.Play(song.FilePath); err != nil {
			a.showError(fmt.Sprintf("Error playing %s: %v", song.Title, err))
		} else {
			a.populateSongList() // Refresh to show playing indicator
		}
	}
}

func (a *App) togglePlayPause() {
	if a.player.GetCurrentSong() == "" && len(a.filteredSongs) > 0 {
		a.playSelected()
	} else {
		a.player.TogglePlayPause()
	}
}

func (a *App) nextSong() {
	if len(a.filteredSongs) == 0 {
		return
	}
	
	a.currentIdx = (a.currentIdx + 1) % len(a.filteredSongs)
	a.songList.SetCurrentItem(a.currentIdx)
	a.playSelected()
}

func (a *App) previousSong() {
	if len(a.filteredSongs) == 0 {
		return
	}
	
	a.currentIdx = (a.currentIdx - 1 + len(a.filteredSongs)) % len(a.filteredSongs)
	a.songList.SetCurrentItem(a.currentIdx)
	a.playSelected()
}

func (a *App) stopPlayback() {
	a.player.Stop()
	a.populateSongList() // Refresh to remove playing indicator
}

func (a *App) increaseVolume() {
	currentVolume := a.player.GetVolume()
	newVolume := currentVolume + 0.1
	if newVolume > 1.0 {
		newVolume = 1.0
	}
	a.player.SetVolume(newVolume)
}

func (a *App) decreaseVolume() {
	currentVolume := a.player.GetVolume()
	newVolume := currentVolume - 0.1
	if newVolume < 0.0 {
		newVolume = 0.0
	}
	a.player.SetVolume(newVolume)
}

func (a *App) enterSearchMode() {
	a.isSearchMode = true
	a.app.SetFocus(a.searchInput)
	a.searchInput.SetText("")
}

func (a *App) exitSearchMode() {
	a.isSearchMode = false
	a.app.SetFocus(a.songList)
	a.searchInput.SetText("")
}

func (a *App) performSearch() {
	query := strings.TrimSpace(a.searchInput.GetText())
	
	if query == "" {
		a.filteredSongs = a.songs
	} else {
		a.filteredSongs = make([]library.Song, 0)
		queryLower := strings.ToLower(query)
		
		for _, song := range a.songs {
			if strings.Contains(strings.ToLower(song.Title), queryLower) ||
			   strings.Contains(strings.ToLower(song.Artist), queryLower) ||
			   strings.Contains(strings.ToLower(song.Album), queryLower) {
				a.filteredSongs = append(a.filteredSongs, song)
			}
		}
	}
	
	a.populateSongList()
	if len(a.filteredSongs) > 0 {
		a.songList.SetCurrentItem(0)
	}
}

func (a *App) formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (a *App) showError(message string) {
	a.statusBar.SetText(fmt.Sprintf(" [red]ERROR:[white] %s", message))
}

func (a *App) updateLoop() {
	ticker := time.NewTicker(500 * time.Millisecond) // Standard update rate
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			a.app.QueueUpdateDraw(func() {
				state := a.player.GetState()
				
				a.progressBar.Update(state.Position, state.Duration)
				a.updateProgressPanel()
				
				a.updateInfoPanel()
				a.updateStatusBar()
				
				if a.player.GetCurrentSong() != "" && a.player.IsFinished() {
					a.handleSongFinished()
				}
			})
		}
	}
}

func (a *App) handleSongFinished() {
	state := a.player.GetState()
	
	if a.player.ShouldRepeat() {
		currentSong := state.CurrentSong
		if currentSong != "" {
			a.player.Play(currentSong)
		}
	} else {
		a.nextSong()
		
		if a.player.ShouldRepeatPlaylist() && a.currentIdx >= len(a.filteredSongs) {
			a.currentIdx = 0
			if len(a.filteredSongs) > 0 {
				a.player.Play(a.filteredSongs[0].FilePath)
				a.songList.SetCurrentItem(0)
				a.populateSongList()
			}
		}
	}
}

func (a *App) cycleRepeatMode() {
	newMode := a.player.CycleRepeatMode()
	a.settingsManager.Update(func(s *settings.Settings) {
		s.RepeatMode = newMode
	})
	
	var modeText string
	switch newMode {
	case settings.RepeatNone:
		modeText = "Repeat: Off"
	case settings.RepeatSingle:
		modeText = "Repeat: Single"
	case settings.RepeatAll:
		modeText = "Repeat: All"
	}
	
	originalUpdate := a.updateStatusBar
	a.statusBar.SetText(fmt.Sprintf(" [yellow]%s[white]", modeText))
	go func() {
		time.Sleep(2 * time.Second)
		a.app.QueueUpdateDraw(originalUpdate)
	}()
}

func (a *App) toggleProgressBar() {
	a.settingsManager.ToggleProgressBar()
	a.updateComponentVisibility()
	a.setupUI() // Rebuild UI with new visibility
	
	status := "off"
	if a.settingsManager.Get().ShowProgressBar {
		status = "on"
	}
	
	originalUpdate := a.updateStatusBar
	a.statusBar.SetText(fmt.Sprintf(" [yellow]Progress bar: %s[white]", status))
	go func() {
		time.Sleep(2 * time.Second)
		a.app.QueueUpdateDraw(originalUpdate)
	}()
}

func (a *App) updateComponentVisibility() {
	settings := a.settingsManager.Get()
	
	if a.progressBar != nil {
		a.progressBar.SetVisible(settings.ShowProgressBar)
	}
}

func (a *App) updateProgressPanel() {
	if a.progressPanel != nil && a.settingsManager.Get().ShowProgressBar {
		content := a.progressBar.Render()
		a.progressPanel.SetText(content)
	}
}

func (a *App) showSettingsInfo() {
	settings := a.settingsManager.Get()
	state := a.player.GetState()
	
	var info strings.Builder
	info.WriteString("[yellow]Settings:[white]\n\n")
	info.WriteString(fmt.Sprintf("Progress Bar: %s\n", boolToOnOff(settings.ShowProgressBar)))
	info.WriteString(fmt.Sprintf("Repeat Mode: %s\n", repeatModeToString(state.RepeatMode)))
	info.WriteString(fmt.Sprintf("Volume: %.0f%%\n", state.Volume*100))
	
	originalText := a.infoPanel.GetText(false)
	a.infoPanel.SetText(info.String())
	
	go func() {
		time.Sleep(4 * time.Second)
		a.app.QueueUpdateDraw(func() {
			a.infoPanel.SetText(originalText)
		})
	}()
}

func boolToOnOff(b bool) string {
	if b {
		return "[green]On[white]"
	}
	return "[red]Off[white]"
}

func repeatModeToString(mode settings.RepeatMode) string {
	switch mode {
	case settings.RepeatNone:
		return "[dim]None[white]"
	case settings.RepeatSingle:
		return "[yellow]Single[white]"
	case settings.RepeatAll:
		return "[green]All[white]"
	default:
		return "[dim]Unknown[white]"
	}
}