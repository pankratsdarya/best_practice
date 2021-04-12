package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	delDuplicates *bool
	dirPath       *string
	allFiles      []filesStruct
	hlog          *log.Entry
)

type filesStruct struct {
	fileEntry   os.DirEntry
	fileSize    int64
	filePath    string
	fileChecked bool
}

func init() {
	delDuplicates = flag.Bool("delDuplicates", false, "Delete duplicates? false=No  true=Yes")
	dirPath = flag.String("dirPath", "F:\\test", "Path to directory for inspection. Program is configured to work on OS Windows.")
	flag.Parse()
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	standardFields := log.Fields{
		"directory": dirPath,
	}
	hlog = log.WithFields(standardFields)

	err := readingFiles(*dirPath)
	if err != nil {
		fmt.Println("Can't read directory. App will close in 3 seconds.")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	for i := 0; i < len(allFiles); i++ {
		checkFiles(i)
	}
}

// readingFiles reads all files in directory given and in its subdirectories
func readingFiles(directoryPath string) error {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		hlog.WithFields(log.Fields{"subdir": directoryPath}).Error("Can't read directory")
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = readingFiles(strings.Join([]string{directoryPath, file.Name()}, "\\"))
			if err != nil {
				return err
			}
		} else {
			var f filesStruct
			fInfo, err := file.Info()
			if err != nil {

				hlog.WithFields(log.Fields{"file": strings.Join([]string{directoryPath, file.Name()}, "\\")}).Warn("Can't read file")
				return err
			}
			f.fileEntry = file
			f.filePath = directoryPath
			f.fileChecked = false
			f.fileSize = fInfo.Size()
			allFiles = append(allFiles, f)
		}
	}
	return nil
}

// checkFiles checks if file on given position in slice have copies
// if delDuplicates flag is true, function ask which files to delete
func checkFiles(num int) {

	var copiesNumber []int
	foundCopy := false

	if allFiles[num].fileChecked {
		return
	}
	for j := num + 1; j < len(allFiles); j++ {
		if allFiles[num].fileEntry.Name() == allFiles[j].fileEntry.Name() && allFiles[num].fileSize == allFiles[j].fileSize {
			copiesNumber = append(copiesNumber, j)
			foundCopy = true
			allFiles[j].fileChecked = true
		}
	}
	if foundCopy {
		if allFiles[num].fileChecked {
			return
		}
		fmt.Println("Found copies: \n1.", allFiles[num].fileEntry.Name(), "    ", allFiles[num].filePath)
		for j := 0; j < len(copiesNumber); j++ {
			fmt.Print(j + 2)
			fmt.Println(". ", allFiles[copiesNumber[j]].fileEntry.Name(), "    ", allFiles[copiesNumber[j]].filePath)
		}
		if *delDuplicates {
			deleteDuplicates(copiesNumber, num)
		}
	}
}

func deleteDuplicates(copNum []int, number int) {
	countDelete := 1
	var numberDelete int
	if len(copNum) > 1 {
		fmt.Println("Enter count of files to delete. Enter 0 to save all files.")
		_, err := fmt.Scanln(&countDelete)
		if err != nil || countDelete > len(copNum) {
			fmt.Println("Wrong count. Files not deleted")
			hlog.Warn("Wrong count entered for ", allFiles[number].fileEntry.Name(), " files. Files not deleted")
			return
		}
	}

	for k := 0; k < countDelete; k++ {
		fmt.Println("Enter number of file to delete. Enter 0 to save all files.")
		_, err := fmt.Scanln(&numberDelete)
		if err != nil || countDelete < numberDelete {
			fmt.Println("Wrong number. Files not deleted")
			hlog.Warn("Wrong number entered for ", allFiles[number].fileEntry.Name(), " files. Files not deleted")
			return
		}
		if numberDelete == 0 {
			return
		}
		os.Chdir(allFiles[copNum[numberDelete-2]].filePath)
		err = os.Remove(allFiles[copNum[numberDelete-2]].fileEntry.Name())
		if err != nil {
			fmt.Println("File not deleted. Error occured.")
			hlog.WithFields(log.Fields{"file": strings.Join([]string{allFiles[copNum[numberDelete-2]].filePath, allFiles[copNum[numberDelete-2]].fileEntry.Name()}, "\\")}).Error("Can't delete file")
		} else {
			fmt.Println("File deleted.")
			hlog.WithFields(log.Fields{"file": strings.Join([]string{allFiles[copNum[numberDelete-2]].filePath, allFiles[copNum[numberDelete-2]].fileEntry.Name()}, "\\")}).Info("File deleted")
		}
	}
}
