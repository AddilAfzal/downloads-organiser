package downloads_organiser

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/addilafzal/downloads-organiser/config"
	"github.com/sirupsen/logrus"
)

type Asset interface {
	Handle()
}

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

func (s TVShow) Handle() {
	showFolder := fmt.Sprintf("%s/%s", config.ShowsFolder, s.Name)
	if _, err := os.Stat(showFolder); os.IsNotExist(err) {
		logrus.Infof("Folder \"%s\" does not exist... creating \n", s.Name)
		err = os.Mkdir(showFolder, os.ModePerm)
		if err != nil {
			logrus.Warnf("Failed to create directory. %s", err)
		}
	}

	showSeasonFolder := fmt.Sprintf("%s/%s", showFolder, s.Season)
	if _, err := os.Stat(showSeasonFolder); os.IsNotExist(err) {
		logrus.Infof("Folder \"%s\" does not exist... creating \n", s.Season)
		err = os.Mkdir(showSeasonFolder, os.ModePerm)
		if err != nil {
			logrus.Warnf("Failed to create directory. %s", err)
		}
	}

	newPath := fmt.Sprintf("%s/%s", showSeasonFolder, s.FileName)
	logrus.Infof("Moving file to %s", newPath)
	err := MoveFile(s.FilePath, newPath)
	if err != nil {
		fmt.Println(err)
	}
}

func (m Movie) Handle() {
	newPath := fmt.Sprintf("%s/%s", config.MoviesFolder, m.FileName)
	logrus.Infof("Moving file to %s", newPath)
	err := MoveFile(m.FilePath, newPath)
	if err != nil {
		logrus.Warningf("Failed to move file: %s", err)
		fmt.Printf("%+v", m)
		return
	}

	fileFolder := path.Dir(m.FilePath)
	if fileFolder == config.DownloadsFolder {
		// It's the downloads folder, we cant delete this
		return
	}

	// Check if empty
	if res, _ := IsEmpty(fileFolder); res {
		logrus.Infof("We can delete %s", fileFolder)
		err = os.Remove(fileFolder)
		if err != nil {
			log.Print(err)
			return
		}
	}
}
