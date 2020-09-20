package main

import (
	"fmt"
	"github.com/rjeczalik/notify"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var showsFolder = ""     // without trailing /
var moviesFolder = ""    // without trailing /
var downloadsFolder = "" // without trailing /

type TVShow struct {
	FileName string
	FilePath string
	Name     string
	Season   string
}

type Movie struct {
	FileName string
	FilePath string
	Name     string
	Year     string
	Quality  string
}

func main() {
	var ok bool
	if moviesFolder, ok = os.LookupEnv("MOVIES_PATH"); !ok {
		moviesFolder = "/media/storage/movies"
	}
	if showsFolder, ok = os.LookupEnv("SHOWS_PATH"); !ok {
		showsFolder = "/media/storage/tv shows"
	}
	if downloadsFolder, ok = os.LookupEnv("DOWNLOADS_PATH"); !ok {
		downloadsFolder = "/media/storage/downloads"
	}

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	// The tv show season format. E.g. S07E01 (will fail without the 0 in S07)
	reShow := regexp.MustCompile(`(.*)(S[0-9]+)E[0-9]+`)
	reMovie := regexp.MustCompile(`^([^()\n]*)\(?([1-2][0-9][0-9][0-9])\)?.*(1080p|2160p|720p).*$`)

	if err := notify.Watch("./test/...", c, notify.InCloseWrite, notify.InMovedTo); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	for {
		// Block until an event is received.
		switch ei := <-c; ei.Event() {
		case notify.InCloseWrite:
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
					handleShow(s)
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
						handleMovie(m)
					} else {
						log.Printf("What is this file? %s \n", fileName)
					}

					log.Println(movieYearQuality)
				}
			} else {
				log.Print("Not .mkv file, skipping")
			}

			//log.Println("Editing of", ei.Path(), "file is done.")
			continue
		case notify.InMovedTo:
			log.Println("File", ei.Path(), "was swapped/moved into the watched directory.")
			continue
		}
	}
}

func handleShow(show *TVShow) {
	showFolder := fmt.Sprintf("%s/%s", showsFolder, show.Name)
	if _, err := os.Stat(showFolder); os.IsNotExist(err) {
		log.Printf("Folder \"%s\" does not exist... creating \n", show.Name)
		os.Mkdir(showFolder, os.ModePerm)
	}

	showSeasonFolder := fmt.Sprintf("%s/%s", showFolder, show.Season)
	if _, err := os.Stat(showSeasonFolder); os.IsNotExist(err) {
		log.Printf("Folder \"%s\" does not exist... creating \n", show.Season)
		os.Mkdir(showSeasonFolder, os.ModePerm)
	}

	newPath := fmt.Sprintf("%s/%s", showSeasonFolder, show.FileName)
	log.Printf("Moving file to %s", newPath)
	err := os.Rename(show.FilePath, newPath)
	if err != nil {
		fmt.Println(err)
	}
}

func handleMovie(movie *Movie) {
	newPath := fmt.Sprintf("%s/%s", moviesFolder, movie.FileName)
	log.Printf("Moving file to %s", newPath)
	err := os.Rename(movie.FilePath, newPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileFolder := path.Dir(movie.FilePath)
	if fileFolder == downloadsFolder {
		// It's the downloads folder, we cant delete this
		return
	}

	// Check if empty
	if res, _ := IsEmpty(fileFolder); res {
		log.Printf("We can delete %s", fileFolder)
		err = os.Remove(fileFolder)
		if err != nil {
			log.Print(err)
			return
		}
	}
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
