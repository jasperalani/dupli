package main

import (
	"flag"
	"fmt"
	goimage "image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	copyfile "github.com/jasperalani/copyfile"
	aurora "github.com/logrusorgru/aurora"
	imgdiff "github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

var logEnabled = false

func main() {

	loc := flag.String("loc", "", "Folder location to scan, Omit for current directory.")
	log := flag.Bool("log", false, "Log the applications processes to a file.")

	flag.Parse()

	logEnabled = *log

	logDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	var directory string
	if "" == *loc {
		*loc = "./"
		directory = "current directory"
	} else {
		directory = *loc
		logDir = directory
	}

	fmt.Printf("%s %s", appName(), aurora.Yellow("Scanning "+directory+" ...\n").Bold())

	appendToLog(
		fmt.Sprintf("%s, %s, %s",
			"SCAN",
			time.Now().Format(time.RFC850),
			logDir))

	var filenames []string

	err = filepath.Walk(*loc, func(path string, info os.FileInfo, err error) error {
		filenames = append(filenames, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	var images []image

	for _, filepath := range filenames {
		split := strings.Split(filepath, ".")

		if len(split) == 1 {
			continue
		}

		extension := split[1]

		if "png" == extension || "jpg" == extension || "jpeg" == extension {
			images = append(images, image{
				filepath: filepath,
			})
		}
	}

	if 0 == len(images) {
		var directory string
		if "./" == *loc {
			directory = "the current directory."
		} else {
			directory = *loc
		}
		fmt.Printf("%s %s", appName(), aurora.Red("No images in "+directory+"\n").Bold())
		appendToLog(
			fmt.Sprintf("%s, %s, %s",
				"FINISH",
				time.Now().Format(time.RFC850),
				"DIR_NO_IMAGES"))
		return
	}

	var (
		imagesCompared = 0
		workingList    = images
		duplicates     []duplicate
		options        *imgdiff.Options = &imgdiff.Options{
			Threshold: 0.1,
		}
	)

	for index, image := range workingList {

		image.image = getImageFromFilePath(image.filepath)
		var sublist = workingList[index+1:]

		for _, subImage := range sublist {

			subImage.image = getImageFromFilePath(subImage.filepath)

			if !image.image.Bounds().Eq(subImage.image.Bounds()) {
				continue
			}

			appendToLog(fmt.Sprintf("%s, %s, %s, %s, %s, %s",
				"COMPARE",
				time.Now().Format(time.RFC850),
				getFileNameFromPath(image.filepath),
				getFileNameFromPath(subImage.filepath),
				image.image.Bounds(),
				subImage.image.Bounds()))

			result := imgdiff.Diff(image.image, subImage.image, options)

			imagesCompared++

			if result.Equal {

				appendToLog(fmt.Sprintf("%s, %s, %s, %s, %s, %s",
					"MATCH",
					time.Now().Format(time.RFC850),
					getFileNameFromPath(image.filepath),
					getFileNameFromPath(subImage.filepath),
					image.image.Bounds(),
					subImage.image.Bounds()))

				duplicates = append(duplicates, duplicate{
					img1: image,
					img2: subImage,
				})
			}

		}

	}

	if 0 == len(duplicates) {
		fmt.Printf("%s %s", appName(), aurora.Green("No duplicates found!\n").Bold())
		appendToLog(
			fmt.Sprintf("%s, %s, %s",
				"FINISH",
				time.Now().Format(time.RFC850),
				"NO_DUP_FOUND"))
		return
	}

	fmt.Printf("%s %s", appName(), aurora.Blue("Found "+strconv.Itoa(len(duplicates))+" duplicates\n").Bold())

	var duplicateDirectory string
	if "./" == *loc {
		duplicateDirectory = "duplicates"
	} else {
		duplicateDirectory = *loc + "/duplicates"
	}
	os.Mkdir(duplicateDirectory, 0777)

	fmt.Printf("%s %s", appName(), aurora.Green("Copying duplicates to folder...\n").Bold())

	for _, duplicate := range duplicates {
		// Copy duplicate
		duplicateFilename := getFileNameFromPath(duplicate.img2.filepath)
		var destination = duplicateDirectory + "/" + duplicateFilename
		err = copyfile.CopyFile(duplicate.img1.filepath, destination)
		if err != nil {
			panic(err)
		}
		appendToLog(
			fmt.Sprintf("%s, %s, %s, %s",
				"COPY",
				time.Now().Format(time.RFC850),
				duplicateFilename,
				duplicateDirectory))
	}

	fmt.Printf("%s %s", appName(), aurora.Red("Deleting original duplicates...\n").Bold())

	for _, duplicate := range duplicates {
		// Delete duplicate
		err = os.Remove(duplicate.img2.filepath)
		if err != nil {
			panic(err)
		}
		appendToLog(
			fmt.Sprintf("%s, %s, %s",
				"DELETE",
				time.Now().Format(time.RFC850),
				duplicate.img1.filepath))
	}

	fmt.Printf("%s %s", appName(), aurora.Green("Finished!\n").Bold())

}

func appName() string {
	return fmt.Sprintf("[%s]", aurora.Black("Dupli").Bold())
}

func appendToLog(text string) {
	if !logEnabled {
		return
	}
	f, err := os.OpenFile("dupli.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text + "\n"); err != nil {
		log.Println(err)
	}
}

func getFileNameFromPath(filepath string) string {
	split := strings.Split(filepath, "/")
	return split[len(split)-1]
}

func getImageFromFilePath(filepath string) goimage.Image {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	image, _, err := goimage.Decode(f)
	if err != nil {
		panic(err)
	}
	return image
}

func getImageConfigFromFilePath(filepath string) goimage.Config {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	image, _, err := goimage.DecodeConfig(f)
	if err != nil {
		panic(err)
	}
	return image
}

type image struct {
	filepath string
	image    goimage.Image
}

type duplicate struct {
	img1  image
	img2  image
	diff  int
	match int
}

type result struct {
	duplicates          []duplicate
	countDuplicates     int
	countImagesCompared int
}
