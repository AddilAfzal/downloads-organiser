package downloads_organiser

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/addilafzal/downloads-organiser/config"
)

func HandleShow(show *TVShow) {
	showFolder := fmt.Sprintf("%s/%s", config.ShowsFolder, show.Name)
	if _, err := os.Stat(showFolder); os.IsNotExist(err) {
		log.Printf("Folder \"%s\" does not exist... creating \n", show.Name)
		err = os.Mkdir(showFolder, os.ModePerm)
		if err != nil {
			log.Printf("Failed to create directory. %s", err)
		}
	}

	showSeasonFolder := fmt.Sprintf("%s/%s", showFolder, show.Season)
	if _, err := os.Stat(showSeasonFolder); os.IsNotExist(err) {
		log.Printf("Folder \"%s\" does not exist... creating \n", show.Season)
		err = os.Mkdir(showSeasonFolder, os.ModePerm)
		if err != nil {
			log.Printf("Failed to create directory. %s", err)
		}
	}

	newPath := fmt.Sprintf("%s/%s", showSeasonFolder, show.FileName)
	log.Printf("Moving file to %s", newPath)
	err := MoveFile(show.FilePath, newPath)
	if err != nil {
		fmt.Println(err)
	}
}

func HandleMovie(movie *Movie) {
	newPath := fmt.Sprintf("%s/%s", config.MoviesFolder, movie.FileName)
	log.Printf("Moving file to %s", newPath)
	err := MoveFile(movie.FilePath, newPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileFolder := path.Dir(movie.FilePath)
	if fileFolder == config.DownloadsFolder {
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
