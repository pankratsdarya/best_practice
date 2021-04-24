package main

import (
	"flag"
	"fmt"
	"io/fs"
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
	myfs          FSImpl
)

type filesStruct struct {
	fileName    string
	fileSize    int64
	filePath    string
	fileChecked bool
}

type myDirEntry struct {
	Name  string
	Size  int64
	IsDir bool
}

type FSImpl struct {
}

func (fs *FSImpl) Remove(input string) error {
	return os.Remove(input)
}

func (fs *FSImpl) Chdir(input string) error {
	return os.Chdir(input)
}

func (fs *FSImpl) ReadDir(input string) ([]fs.DirEntry, error) {
	return os.ReadDir(input)
}

type FS interface {
	Remove()
	ReadDir()
	Chdir()
	MyDirEntry()
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

	allFiles, err := readingFiles(*dirPath)
	if err != nil {
		// здесь запись в лог не нужна, потому что это обработка ошибки,
		// вернувшейся из функции readingFiles, а там все ошибки уже в логах.
		fmt.Println("Can't read directory. App will close in 3 seconds.")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	for i := 0; i < len(allFiles); i++ {
		checkFiles(i)
	}

	fmt.Println("Checking complete")
}

// readingFiles reads all files in directory given and in its subdirectories
func readingFiles(directoryPath string) ([]filesStruct, error) {
	files, err := myfs.MyReadDir(directoryPath)
	if err != nil {
		hlog.WithFields(log.Fields{"subdir": directoryPath}).Error("Can't read directory")
		return nil, err
	}

	for _, file := range files {
		if file.IsDir {
			allFiles, err = readingFiles(strings.Join([]string{directoryPath, file.Name}, "\\"))
			if err != nil {
				hlog.WithFields(log.Fields{"file": strings.Join([]string{directoryPath, file.Name}, "\\")}).Warn("Can't read file")
				return nil, err
			}
		} else {
			var f filesStruct
			f.fileName = file.Name
			f.filePath = directoryPath
			f.fileChecked = false
			f.fileSize = file.Size
			allFiles = append(allFiles, f)
		}
	}
	return allFiles, nil
}

func (fs *FSImpl) MyReadDir(input string) ([]myDirEntry, error) {
	files, err := fs.ReadDir(input)
	if err != nil {
		hlog.WithFields(log.Fields{"subdir": input}).Error("Can't read directory")
		return nil, err
	}

	var dirFiles []myDirEntry
	var ent myDirEntry
	for _, file := range files {
		if file.IsDir() {
			ent.IsDir = true
		}
		fInfo, err := file.Info()
		if err != nil {

			hlog.WithFields(log.Fields{"file": strings.Join([]string{input, file.Name()}, "\\")}).Warn("Can't read file")
			return nil, err
		}
		ent.Name = file.Name()
		ent.Size = fInfo.Size()
		dirFiles = append(dirFiles, ent)
	}

	return dirFiles, nil
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
		if allFiles[num].fileName == allFiles[j].fileName && allFiles[num].fileSize == allFiles[j].fileSize {
			copiesNumber = append(copiesNumber, j)
			foundCopy = true
			allFiles[j].fileChecked = true
		}
	}
	if foundCopy {
		if allFiles[num].fileChecked {
			return
		}
		fmt.Println("Found copies: \n1.", allFiles[num].fileName, "    ", allFiles[num].filePath)
		for j := 0; j < len(copiesNumber); j++ {
			fmt.Print(j + 2)
			fmt.Println(". ", allFiles[copiesNumber[j]].fileName, "    ", allFiles[copiesNumber[j]].filePath)
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
			hlog.Warn("Wrong count entered for ", allFiles[number].fileName, " files. Files not deleted")
			return
		}
	}

	for k := 0; k < countDelete; k++ {
		fmt.Println("Enter number of file to delete. Enter 0 to save all files.")
		_, err := fmt.Scanln(&numberDelete)
		if err != nil || countDelete < numberDelete {
			fmt.Println("Wrong number. Files not deleted")
			hlog.Warn("Wrong number entered for ", allFiles[number].fileName, " files. Files not deleted")
			return
		}
		if numberDelete == 0 {
			return
		}
		err = myfs.Chdir(allFiles[copNum[numberDelete-2]].filePath)
		if err != nil {
			fmt.Println("Error changing directory.")
			hlog.WithFields(log.Fields{"file": strings.Join([]string{allFiles[copNum[numberDelete-2]].filePath, allFiles[copNum[numberDelete-2]].fileName}, "\\")}).Error("Error changing directory.")
			return
		}
		err = myfs.Remove(allFiles[copNum[numberDelete-2]].fileName)
		if err != nil {
			fmt.Println("File not deleted. Error occured.")
			hlog.WithFields(log.Fields{"file": strings.Join([]string{allFiles[copNum[numberDelete-2]].filePath, allFiles[copNum[numberDelete-2]].fileName}, "\\")}).Error("Can't delete file")
		} else {
			fmt.Println("File deleted.")
			hlog.WithFields(log.Fields{"file": strings.Join([]string{allFiles[copNum[numberDelete-2]].filePath, allFiles[copNum[numberDelete-2]].fileName}, "\\")}).Info("File deleted")
		}
	}
}
