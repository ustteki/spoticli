package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"clispot/internal/library"
	"clispot/internal/player"
	"clispot/internal/ui"
)

func main() {
	var musicDir string
	flag.StringVar(&musicDir, "dir", "./music", "Directory containing MP3 files")
	flag.Parse()

	
	absPath, err := filepath.Abs(musicDir)
	if err != nil {
		log.Fatalf("Error resolving music directory: %v", err)
	}

	
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Music directory '%s' does not exist. Creating it...\n", absPath)
		if err := os.MkdirAll(absPath, 0755); err != nil {
			log.Fatalf("Error creating music directory: %v", err)
		}
		fmt.Printf("Created music directory at: %s\n", absPath)
		fmt.Println("Add some MP3 files to this directory and run clispot again!")
		return
	}

	
	lib := library.NewLibrary(absPath)
	
	
	fmt.Printf("Scanning for music in: %s\n", absPath)
	songs, err := lib.ScanDirectory()
	if err != nil {
		log.Fatalf("Error scanning music directory: %v", err)
	}

	if len(songs) == 0 {
		fmt.Printf("No MP3 files found in '%s'\n", absPath)
		fmt.Println("Add some MP3 files and try again!")
		return
	}

	fmt.Printf("Found %d songs\n", len(songs))

	
	audioPlayer := player.NewPlayer()
	defer audioPlayer.Close()

	
	app := ui.NewApp(songs, audioPlayer)
	
	
	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}