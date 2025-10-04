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
	
	// Library management
	library      *library.Library
	currentItems []library.LibraryItem
	
	// Settings and components
	settingsManager *settings.Manager
	progressBar     *progressbar.ProgressBar
	
	// UI Components
	songList        *tview.List
	infoPanel       *tview.TextView
	statusBar       *tview.TextView
	helpText        *tview.TextView
	searchInput     *tview.InputField
	progressPanel   *tview.TextView
	breadcrumb      *tview.TextView
	
	// State
	isSearchMode bool
	filteredSongs []library.Song
}


// NewApp creates a new application instance
func NewApp(songs []library.Song, audioPlayer *player.Player, lib *library.Library) *App {
	// Initialize settings manager
	settingsManager := settings.NewManager()
	
	app := &App{
		app:             tview.NewApplication(),
		songs:           songs,
		player:          audioPlayer,
		currentIdx:      0,
		filteredSongs:   songs,
		library:         lib,
		settingsManager: settingsManager,
		progressBar:     progressbar.NewProgressBar(80),  // Default width
	}
	
	// Load current items from library
	items, err := lib.GetCurrentItems()
	if err != nil {
		// Fallback to songs only
		app.currentItems = make([]library.LibraryItem, 0)
		for _, song := range songs {
			app.currentItems = append(app.currentItems, library.LibraryItem{
				Type: library.ItemTypeSong,
				Name: fmt.Sprintf("%s - %s", song.Artist, song.Title),
				Path: song.FilePath,
				Song: &song,
			})
		}
	} else {
		app.currentItems = items
	}
	
	return app
}


// Run starts the application
func (a *App) Run() error {
	a.setupUI()
	a.setupKeyBindings()
	a.populateLibraryList()
	a.updateBreadcrumb()
	a.updateInfoPanel()
	a.updateStatusBar()
	
	// Start a goroutine to update the UI periodically
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
	a.breadcrumb = tview.NewTextView().SetDynamicColors(true)
	
	
	a.songList.SetBorder(true).SetTitle(" Library Browser ")
	a.infoPanel.SetBorder(true).SetTitle(" Now Playing ")
	a.statusBar.SetBorder(false)
	a.helpText.SetBorder(true).SetTitle(" Controls ")
	a.searchInput.SetBorder(true).SetTitle(" Search ")
	a.progressPanel.SetBorder(false)
	a.breadcrumb.SetBorder(true).SetTitle(" Location ")
	
	
	helpStr := `[yellow]Controls:[white]
[green]Space[white] - Play/Pause    [green]Enter[white] - Play/Browse
[green]↑/↓[white] - Navigate        [green]n[white] - Next song
[green]p[white] - Previous song     [green]/[white] - Search
[green]Esc[white] - Exit search     [green]q[white] - Quit
[green]+/-[white] - Volume up/down  [green]s[white] - Stop
[green]L[white] - Repeat mode       [green]B[white] - Toggle progress
[green]Backspace[white] - Go back   [green]?[white] - Settings info`
	
	a.helpText.SetText(helpStr)
	
	
	a.searchInput.SetText("")
	
	
	a.updateComponentVisibility()
	
	
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.breadcrumb, 2, 0, false).
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
			return event 
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
			a.handleSelection()
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			a.navigateBack()
			return nil
		}
		return event
	})
	
	// Song list selection handler
	a.songList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		a.handleSelection()
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
	state := a.player.GetState()
	
	if state.CurrentSong == "" {
		// Show currently selected item info
		currentIdx := a.songList.GetCurrentItem()
		if currentIdx >= 0 && currentIdx < len(a.currentItems) {
			item := a.currentItems[currentIdx]
			
			if item.Type == library.ItemTypeFolder {
				// Show folder info
				var info strings.Builder
				if item.Name == ".." {
					info.WriteString("[blue]📁 Parent Directory[white]\n\n")
					info.WriteString("[dim]Press Enter to go back[white]")
				} else {
					info.WriteString(fmt.Sprintf("[blue]📁 Playlist: %s[white]\n\n", item.Name))
					info.WriteString(fmt.Sprintf("[green]Songs:[white] %d\n\n", item.SongCount))
					info.WriteString("[dim]Press Enter to browse[white]")
				}
				a.infoPanel.SetText(info.String())
			} else if item.Song != nil {
				// Show song info
				song := item.Song
				var info strings.Builder
				info.WriteString("[yellow]Selected:[white]\n")
				
				// Add album art if available
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
			a.infoPanel.SetText("[red]No item selected[white]")
		}
	} else {
		// Show currently playing song info
		var currentSong *library.Song
		for _, song := range a.songs {
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
			
			// Add album art if available
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
			a.populateSongList() 
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
	// Don't change selection, just refresh the display
	a.populateLibraryList()
	a.updateInfoPanel()
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
	ticker := time.NewTicker(500 * time.Millisecond) 
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
	a.setupUI() 
	
	
	a.populateSongList()
	a.updateInfoPanel()
	
	
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

// populateLibraryList populates the list with current library items (folders and songs)
func (a *App) populateLibraryList() {
	a.songList.Clear()
	
	for i, item := range a.currentItems {
		var mainText, secondaryText string
		
		if item.Type == library.ItemTypeFolder {
			if item.Name == ".." {
				mainText = "[blue]📁 .. (Back)[white]"
				secondaryText = ""
			} else {
				mainText = fmt.Sprintf("[blue]📁 %s[white]", item.Name)
				secondaryText = fmt.Sprintf("%d songs", item.SongCount)
			}
		} else {
			// Song item
			song := item.Song
			mainText = fmt.Sprintf("%s - %s", song.Artist, song.Title)
			secondaryText = fmt.Sprintf("%s | %s", song.Album, a.formatDuration(song.Duration))
			
			// Highlight currently playing song
			if a.player.GetCurrentSong() == song.FilePath {
				mainText = fmt.Sprintf("[yellow]♪ %s[white]", mainText)
				a.songList.SetCurrentItem(i)
			}
		}
		
		a.songList.AddItem(mainText, secondaryText, 0, nil)
	}
}

// updateBreadcrumb updates the breadcrumb navigation
func (a *App) updateBreadcrumb() {
	if a.library == nil {
		a.breadcrumb.SetText(" [dim]No library loaded[white]")
		return
	}
	
	relativePath := a.library.GetRelativePath()
	if relativePath == "/" {
		a.breadcrumb.SetText(" [cyan]Music Library[white] [dim]>[white] [yellow]Root[white]")
	} else {
		a.breadcrumb.SetText(fmt.Sprintf(" [cyan]Music Library[white] [dim]>[white] [yellow]%s[white]", relativePath))
	}
}

// handleSelection handles both folder navigation and song playing
func (a *App) handleSelection() {
	currentIdx := a.songList.GetCurrentItem()
	if currentIdx < 0 || currentIdx >= len(a.currentItems) {
		return
	}
	
	item := a.currentItems[currentIdx]
	
	if item.Type == library.ItemTypeFolder {
		// Navigate to folder
		err := a.library.NavigateToFolder(item.Path)
		if err != nil {
			a.statusBar.SetText(fmt.Sprintf(" [red]Error: %v[white]", err))
			return
		}
		
		// Update current items
		items, err := a.library.GetCurrentItems()
		if err != nil {
			a.statusBar.SetText(fmt.Sprintf(" [red]Error loading folder: %v[white]", err))
			return
		}
		
		a.currentItems = items
		a.populateLibraryList()
		a.updateBreadcrumb()
		a.updateInfoPanel()
		
	} else {
		// Play song
		a.playSelectedSong(item.Song)
	}
}

// navigateBack goes back to parent directory
func (a *App) navigateBack() {
	if a.library == nil || !a.library.CanGoBack() {
		return
	}
	
	currentPath := a.library.GetCurrentPath()
	parentPath := filepath.Dir(currentPath)
	
	err := a.library.NavigateToFolder(parentPath)
	if err != nil {
		a.statusBar.SetText(fmt.Sprintf(" [red]Error: %v[white]", err))
		return
	}
	
	// Update current items
	items, err := a.library.GetCurrentItems()
	if err != nil {
		a.statusBar.SetText(fmt.Sprintf(" [red]Error loading folder: %v[white]", err))
		return
	}
	
	a.currentItems = items
	a.populateLibraryList()
	a.updateBreadcrumb()
	a.updateInfoPanel()
}

// playSelectedSong plays a specific song
func (a *App) playSelectedSong(song *library.Song) {
	if song == nil {
		return
	}
	
	// Find the song in the full song list to get the index
	for i, s := range a.songs {
		if s.FilePath == song.FilePath {
			a.currentIdx = i
			break
		}
	}
	
	// Play the song
	err := a.player.Play(song.FilePath)
	if err != nil {
		a.statusBar.SetText(fmt.Sprintf(" [red]Error playing: %v[white]", err))
		return
	}
	
	a.populateLibraryList() // Refresh to show playing indicator
	a.updateInfoPanel()
}

// getCurrentPlayableSongs returns songs available in the current context
func (a *App) getCurrentPlayableSongs() []library.Song {
	// If we're in a folder, get songs from that folder
	if a.library != nil {
		folderSongs := a.library.GetSongsInCurrentFolder()
		if len(folderSongs) > 0 {
			return folderSongs
		}
	}
	
	// Fallback to filtered songs or all songs
	if len(a.filteredSongs) > 0 {
		return a.filteredSongs
	}
	
	return a.songs
}

// playSpecificSong plays a specific song and updates UI
func (a *App) playSpecificSong(song *library.Song) {
	if song == nil {
		return
	}
	
	err := a.player.Play(song.FilePath)
	if err != nil {
		a.statusBar.SetText(fmt.Sprintf(" [red]Error playing: %v[white]", err))
		return
	}
	
	// Update UI to reflect the new playing song
	a.populateLibraryList()
	a.updateInfoPanel()
	
	// Try to highlight the playing song in the list
	for i, item := range a.currentItems {
		if item.Type == library.ItemTypeSong && item.Song != nil && item.Song.FilePath == song.FilePath {
			a.songList.SetCurrentItem(i)
			break
		}
	}
}