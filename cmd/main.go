package cmd

import (
	"flag"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	imgdiff "github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

func main() {

	scan := flag.Bool("scan", false, "Start the scan.")
	loc := flag.String("loc", "", "Folder location to scan, default is current directory.")

	flag.Parse()

	if false == *scan {
		return
	}

	if "" == *loc {
		*loc = "./"
	}

	var filenames []string

	err := filepath.Walk(*loc, func(path string, info os.FileInfo, err error) error {
		filenames = append(filenames, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	var imagePaths []string

	for _, filepath := range filenames {
		split := strings.Split(filepath, ".")

		if len(split) == 1 {
			continue
		}

		extension := split[1]

		if "png" == extension || "jpg" == extension || "jpeg" == extension {
			imagePaths = append(imagePaths, filepath)
		}
	}

	var (
		// diffsFound           = 0
		workingList = imagePaths
		// differences []imageDifference
		options *imgdiff.Options = &imgdiff.Options{
			Threshold: 1,
			DiffImage: false,
		}
	)

	for index, imagePath := range workingList {

		image1 := getImageFromFilePath(imagePath)
		workingList = remove(workingList, index)

		for sindex, simagePath := range workingList {

			image2 := getImageFromFilePath(simagePath)

			result := imgdiff.Diff(image1, image2, options)
			workingList = remove(workingList, sindex)

			log.Println(result.Equal)

		}

	}

}

func remove(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

func getImageFromFilePath(filePath string) image.Image {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	image, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return image
}

type imageDifference struct {
	img1  string
	img2  string
	diff  int
	match int
}
