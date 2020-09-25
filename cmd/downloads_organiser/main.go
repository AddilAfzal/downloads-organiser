package main

import (
	"downloadsOrganiser/config"
	. "downloadsOrganiser/internal/downloads_organiser"
	"fmt"
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

// The tv show season format. E.g. S07E01 (will fail without the 0 in S07)
var reShow = regexp.MustCompile(`(.*)(S[0-9]+)E[0-9]+`)
var reMovie = regexp.MustCompile(`^([^()\n]*)\(?([1-2][0-9][0-9][0-9])\)?.*(1080p|2160p|720p).*$`)

func main() {
	var ok bool
	if config.MoviesFolder, ok = os.LookupEnv("MOVIES_PATH"); !ok {
		config.MoviesFolder = "/media/storage/movies"
	}
	if config.ShowsFolder, ok = os.LookupEnv("SHOWS_PATH"); !ok {
		config.ShowsFolder = "/media/storage/tv shows"
	}
	if config.DownloadsFolder, ok = os.LookupEnv("DOWNLOADS_PATH"); !ok {
		config.DownloadsFolder = "/home/storage/downloads"
	}

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	watchFolder := fmt.Sprintf("%s/...", config.DownloadsFolder)
	log.Printf("Watching `%s`", watchFolder)
	if err := notify.Watch(watchFolder, c, notify.InCloseWrite, notify.InMovedTo); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	for {
		// Block until an event is received.
		switch ei := <-c; ei.Event() {
		case notify.InCloseWrite:
			log.Println("File", ei.Path(), "was written to in the watched directory.")
			handleFile(ei)
			continue
		case notify.InMovedTo:
			log.Println("File", ei.Path(), "was swapped/moved into the watched directory.")
			handleFile(ei)
			continue
		}
	}
}

func handleFile(ei notify.EventInfo) {
	// If mkv file.
	if strings.HasSuffix(ei.Path(), ".mkv") {
		fileName := path.Base(ei.Path())

		// Parse file name.
		showSeason := reShow.FindAllStringSubmatch(fileName, 3)

		if showSeason != nil {
			// It's a TV show
			show := strings.Replace(showSeason[0][1], ".", " ", -1) // Remove dots "."
			show = strings.Trim(show, " ")                          // Remove spaces
			season := showSeason[0][2]                              // with S (S07)

			s := &TVShow{
				FileName: fileName,
				FilePath: ei.Path(),
				Name:     show,
				Season:   season,
			}
			HandleShow(s)
		} else {
			// It's a film maybe?
			movieYearQuality := reMovie.FindAllStringSubmatch(fileName, 3)
			if movieYearQuality != nil {
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
				log.Printf("What is this file? %s \n", fileName)
			}

			log.Println(movieYearQuality)
		}
	} else {
		log.Print("Not .mkv file, skipping")
	}
}
