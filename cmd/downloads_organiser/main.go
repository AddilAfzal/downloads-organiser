package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/addilafzal/downloads-organiser/config"
	. "github.com/addilafzal/downloads-organiser/internal/downloads_organiser"
	"github.com/rjeczalik/notify"
)

func main() {
	config.Init()
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
	})

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1000)

	watchFolder := fmt.Sprintf("%s/...", config.DownloadsFolder)
	logrus.Infof("Watching `%s`", watchFolder)
	if err := notify.Watch(watchFolder, c, notify.InCloseWrite, notify.InMovedTo, notify.InMoveSelf, notify.InCloseNowrite); err != nil {
		logrus.Fatal(err)
	}
	defer notify.Stop(c)

	for {
		// Block until an event is received.
		switch ei := <-c; ei.Event() {
		case notify.InCloseWrite:
			logrus.Info("File", ei.Path(), "was written to in the watched directory.")
			handle(ei)
			continue
		case notify.InMovedTo:
			logrus.Info("File", ei.Path(), "was swapped/moved into the watched directory.")
			handle(ei)
			continue
		default:
			if config.Debug == "true" {
				log.Println(ei)
				log.Println(ei.Event().String())
				log.Println(ei.Path())
				continue
			}
		}
	}
}

func handle(ei notify.EventInfo) {
	file, err := os.Open(ei.Path())
	if err != nil {
		logrus.Warn("Failed to open file", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		// error handling
	}

	if fileInfo.IsDir() {
		err = filepath.Walk(ei.Path(), handleFile)
		if err != nil {
			logrus.Warnf("Couldn't walk directory: %s", err)
		}
		return
	}

	strings.Split(ei.Path(), "/")
	handleFile(ei.Path(), fileInfo, nil)
}

func handleFile(path string, info fs.FileInfo, err error) error {
	// We only care about MKV files for now
	if strings.HasSuffix(info.Name(), ".mkv") || strings.HasSuffix(info.Name(), ".mp4") {
		fileName := info.Name()
		var asset Asset
		if showSeason := ReShow.FindAllStringSubmatch(fileName, 3); showSeason != nil {
			// It's a TV show
			show := strings.Replace(showSeason[0][1], ".", " ", -1) // Remove dots "."
			show = strings.Trim(show, " ")                          // Remove spaces
			season := showSeason[0][2]                              // with S (S07)

			asset = &TVShow{
				FileName: fileName,
				FilePath: path,
				Name:     strings.Title(show),
				Season:   strings.Title(season),
			}
		} else if movieYearQuality := ReMovie.FindAllStringSubmatch(fileName, 3); movieYearQuality != nil {
			// It's a Movie
			movie := strings.Replace(movieYearQuality[0][1], ".", " ", -1) // Remove dots "."
			movie = strings.Trim(movie, " ")                               // Remove spaces
			year := movieYearQuality[0][2]
			quality := movieYearQuality[0][3]

			asset = &Movie{
				FileName: fileName,
				FilePath: path,
				Name:     movie,
				Year:     year,
				Quality:  quality,
			}
		} else {
			logrus.Warnf("What is this file? %s \n", fileName)
			return nil
		}
		asset.Handle()
		return nil
	}
	logrus.Info("Not .mkv file, skipping")
	return nil
}
