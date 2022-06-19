package config

import (
	"os"
	"path/filepath"
)

var (
	ShowsFolder     = getEnv("SHOWS_PATH", "./tv show")
	MoviesFolder    = getEnv("MOVIES_PATH", "./movies")
	DownloadsFolder = getEnv("DOWNLOADS_PATH", "./downloads")
	Debug           = getEnv("DEBUG", "false")
)

func Init() {
	var err error
	ShowsFolder, err = filepath.Abs(ShowsFolder)
	if err != nil {
		panic(err)
	}
	MoviesFolder, err = filepath.Abs(MoviesFolder)
	if err != nil {
		panic(err)
	}
	DownloadsFolder, err = filepath.Abs(DownloadsFolder)
	if err != nil {
		panic(err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
