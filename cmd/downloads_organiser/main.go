package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"path"
	"strings"

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
			handleFile(ei)
			continue
		case notify.InMovedTo:
			logrus.Info("File", ei.Path(), "was swapped/moved into the watched directory.")
			handleFile(ei)
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

func handleFile(ei notify.EventInfo) {
	// We only care about MKV files for now
	if strings.HasSuffix(ei.Path(), ".mkv") {
		fileName := path.Base(ei.Path())

		if showSeason := ReShow.FindAllStringSubmatch(fileName, 3); showSeason != nil {
			// It's a TV show
			show := strings.Replace(showSeason[0][1], ".", " ", -1) // Remove dots "."
			show = strings.Trim(show, " ")                          // Remove spaces
			season := showSeason[0][2]                              // with S (S07)

			s := &TVShow{
				FileName: fileName,
				FilePath: ei.Path(),
				Name:     strings.Title(show),
				Season:   strings.Title(season),
			}
			HandleShow(s)
		} else if movieYearQuality := ReMovie.FindAllStringSubmatch(fileName, 3); movieYearQuality != nil {
			// It's a Movie
			movie := strings.Replace(movieYearQuality[0][1], ".", " ", -1) // Remove dots "."
			movie = strings.Trim(movie, " ")                               // Remove spaces
			year := movieYearQuality[0][2]
			quality := movieYearQuality[0][3]

			m := &Movie{
				FileName: fileName,
				FilePath: ei.Path(),
				Name:     movie,
				Year:     year,
				Quality:  quality,
			}
			HandleMovie(m)
		} else {
			logrus.Warnf("What is this file? %s \n", fileName)
		}
	} else {
		logrus.Info("Not .mkv file, skipping")
	}
}
